package cache

import (
	"io/fs"
	"io/ioutil"
	"os"
)

type AWSCredentialsCacheOsClientDefault struct {
	AWSCredentialsCacheOsClient
}

func (AWSCredentialsCacheOsClientDefault) Stat(name string) (fs.FileInfo, error) {
	return os.Stat(name)
}

func (AWSCredentialsCacheOsClientDefault) MkdirAll(path string, perm fs.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (AWSCredentialsCacheOsClientDefault) WriteFile(filename string, data []byte, perm fs.FileMode) error {
	return ioutil.WriteFile(filename, data, perm)
}

func (AWSCredentialsCacheOsClientDefault) ReadFile(filename string) ([]byte, error) {
	return ioutil.ReadFile(filename)
}

func (AWSCredentialsCacheOsClientDefault) Remove(name string) error {
	return os.Remove(name)
}
