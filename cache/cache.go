package cache

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/fs"
	"nextunit/op2aws/awsvault"
	"nextunit/op2aws/opaws"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/service/sts"
)

const FILEMODE = os.FileMode(int(0777))

type AWSCredentialsCacheClient struct {
	osClient    AWSCredentialsCacheOsClient
	path        string
	vault       string
	item        string
	mfa         string
	assume_role string
}

type AWSCredentialsCacheOsClient interface {
	Stat(name string) (fs.FileInfo, error)
	MkdirAll(path string, perm fs.FileMode) error
	WriteFile(filename string, data []byte, perm fs.FileMode) error
	ReadFile(filename string) ([]byte, error)
	Remove(name string) error
}

func (cache AWSCredentialsCacheClient) checkCacheDir() {
	_, err := cache.osClient.Stat(cache.path)
	if err == nil {
		return
	}

	cache.osClient.MkdirAll(cache.path, FILEMODE)
}

func (cache AWSCredentialsCacheClient) getFilePath() string {
	filehash := md5.Sum([]byte(fmt.Sprintf("%s-%s-%s-%s", cache.vault, cache.item, cache.mfa, cache.assume_role)))
	return fmt.Sprintf("%s/%x", cache.path, string(filehash[:]))
}

func (cache AWSCredentialsCacheClient) Store(credentials *sts.Credentials) error {
	cache.checkCacheDir()
	filepath := cache.getFilePath()

	content, err := json.Marshal(credentials)
	if err != nil {
		return err
	}

	return cache.osClient.WriteFile(filepath, content, FILEMODE)
}

func (cache AWSCredentialsCacheClient) GetCache() (*sts.Credentials, error) {
	cache.checkCacheDir()
	filepath := cache.getFilePath()

	_, err := cache.osClient.Stat(filepath)
	if err != nil {
		return nil, nil
	}

	content, err := cache.osClient.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	credentials := &sts.Credentials{}
	json.Unmarshal(content, credentials)

	if credentials.Expiration.Before(time.Now()) {
		cache.osClient.Remove(filepath)
		return nil, nil
	}

	return credentials, nil
}

func (cache *AWSCredentialsCacheClient) GenerateFromOP(client awsvault.Vault) {
	cache.vault = client.GetVault()
	cache.item = client.GetItem()
}

func (cache *AWSCredentialsCacheClient) GenerateFromOPAWS(client *opaws.OpAWS) {
	cache.mfa = client.GetMFA()
	cache.assume_role = client.GetAssumeRole()
}

func (cache *AWSCredentialsCacheClient) Vault(vault string) {
	cache.vault = vault
}

func (cache *AWSCredentialsCacheClient) Item(item string) {
	cache.item = item
}

func (cache *AWSCredentialsCacheClient) MFA(mfa string) {
	cache.mfa = mfa
}

func (cache *AWSCredentialsCacheClient) AssumeRole(assume_role string) {
	cache.assume_role = assume_role
}

func New(osClient AWSCredentialsCacheOsClient, path string) *AWSCredentialsCacheClient {
	return &AWSCredentialsCacheClient{osClient: osClient, path: path}
}
