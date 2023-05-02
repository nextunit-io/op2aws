package cache_test

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"nextunit/op2aws/awsvault"
	"nextunit/op2aws/cache"
	"nextunit/op2aws/opaws"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/stretchr/testify/assert"
)

var (
	expirationDate time.Time

	statReturnValue      fs.FileInfo
	mkdirAllReturnValue  error
	writeFileReturnValue error
	readFileReturnValue  []byte
	removeReturnValue    error

	statCallCount      int
	mkdirAllCallCount  int
	writeFileCallCount int
	readFileCallCount  int
	removeCallCount    int

	statInput     []string
	mkdirAllInput struct {
		path string
		perm fs.FileMode
	}
	writeFileInput struct {
		filename string
		data     []byte
		perm     fs.FileMode
	}
	readFileInput string
	removeInput   string

	testCasesGetCache = []testCaseModel{
		{
			Vault:            "test-vault",
			Item:             "test-item",
			Mfa:              "test-mfa",
			AssumeRole:       "test-assume-role",
			ExpectedFileName: "test-path/41f9ea0d0f7c3b470458d338de7ee037",
		},
		{
			Vault:            "test-vault-1",
			Item:             "test-item-1",
			Mfa:              "test-mfa-1",
			AssumeRole:       "test-assume-role-1",
			ExpectedFileName: "test-path/97451959f989e485ea05856a73a48366",
		},
	}
)

type testCaseModel struct {
	Vault            string
	Item             string
	Mfa              string
	AssumeRole       string
	ExpectedFileName string

	AwsVault awsvault.Vault
	OpAws    *opaws.OpAWS
}

type testAwsFileInfoMock struct {
	fs.FileInfo
}

type testCredentialsCacheOsClientMock struct {
	cache.AWSCredentialsCacheOsClient
}

func (testCredentialsCacheOsClientMock) Stat(name string) (fs.FileInfo, error) {
	statCallCount++
	statInput = append(statInput, name)
	if statReturnValue == nil {
		return nil, fmt.Errorf("test-error stat")
	}
	return statReturnValue, nil
}

func (testCredentialsCacheOsClientMock) MkdirAll(path string, perm fs.FileMode) error {
	mkdirAllCallCount++
	mkdirAllInput = struct {
		path string
		perm fs.FileMode
	}{
		path: path,
		perm: perm,
	}
	return mkdirAllReturnValue
}

func (testCredentialsCacheOsClientMock) WriteFile(filename string, data []byte, perm fs.FileMode) error {
	writeFileCallCount++
	writeFileInput = struct {
		filename string
		data     []byte
		perm     fs.FileMode
	}{
		filename: filename,
		data:     data,
		perm:     perm,
	}
	return writeFileReturnValue
}

func (testCredentialsCacheOsClientMock) ReadFile(filename string) ([]byte, error) {
	readFileCallCount++
	readFileInput = filename
	if readFileReturnValue == nil {
		return nil, fmt.Errorf("test-error ReadFile")
	}
	return readFileReturnValue, nil
}

func (testCredentialsCacheOsClientMock) Remove(name string) error {
	removeCallCount++
	removeInput = name
	return removeReturnValue
}

func init() {
	setupTestCases()
	vault := awsvault.NewOnePasswordVault(&awsvault.CommandClientDefault{}, "test-vault-3", "test-item-3")
	opClient1 := opaws.New(vault, &opaws.OpAwsDefaultInput{})
	opClient1.AssumeRole("test-assume-role-3")
	opClient1.UseMFA("test-mfa-3")
	testCasesGetCache = append(testCasesGetCache, testCaseModel{
		AwsVault:         vault,
		OpAws:            opClient1,
		ExpectedFileName: "test-path/2faabbccae1170d3ec64bcc9fe523d0e",
	})
}

func setupTestCases() {
	expirationDate := time.Now().Add(time.Hour)
	currentTimeString, _ := expirationDate.UTC().MarshalText()

	statReturnValue = &testAwsFileInfoMock{}
	mkdirAllReturnValue = nil
	writeFileReturnValue = nil
	readFileReturnValue = []byte(fmt.Sprintf("{\"AccessKeyId\":\"access-key-id\",\"Expiration\":\"%s\",\"SecretAccessKey\":\"secret-access-key\",\"SessionToken\":\"session-token\"}", string(currentTimeString)))
	removeReturnValue = nil

	statCallCount = 0
	mkdirAllCallCount = 0
	writeFileCallCount = 0
	readFileCallCount = 0
	removeCallCount = 0

	statInput = []string{}
	mkdirAllInput = struct {
		path string
		perm fs.FileMode
	}{}
	writeFileInput = struct {
		filename string
		data     []byte
		perm     fs.FileMode
	}{}
	readFileInput = ""
	removeInput = ""
}

func TestGetCache(t *testing.T) {
	t.Helper()

	for i, v := range testCasesGetCache {
		t.Run(fmt.Sprintf("Running GetCache test with valid credentials %d", i), func(t *testing.T) {
			assert := assert.New(t)
			t.Helper()
			setupTestCases()

			client := cache.New(&testCredentialsCacheOsClientMock{}, "test-path")

			if v.AwsVault == nil {
				client.Vault(v.Vault)
				client.Item(v.Item)
			} else {
				client.GenerateFromOP(v.AwsVault)
			}
			if v.OpAws == nil {
				client.MFA(v.Mfa)
				client.AssumeRole(v.AssumeRole)
			} else {
				client.GenerateFromOPAWS(v.OpAws)
			}

			credentials, err := client.GetCache()

			assert.Nil(err)

			stsCredentials := &sts.Credentials{}
			json.Unmarshal(readFileReturnValue, stsCredentials)

			assert.Equal(stsCredentials, credentials)

			assert.Equal("test-path", statInput[0], "First check is checking if directory path is existing")
			assert.Equal(0, mkdirAllCallCount, "MkdirAll should not be called, because the Stat function is not returning an error")
			assert.Equal(v.ExpectedFileName, statInput[1], "Second is checking the filepath for cache file is existing")
			assert.Equal(1, readFileCallCount, "ReadFile should be called")

			assert.Equal(v.ExpectedFileName, readFileInput)
			assert.Equal(0, removeCallCount, "Remove call should not be executed, since the credentials are still valid")
		})
		t.Run(fmt.Sprintf("Running GetCache test with invalid credentials %d", i), func(t *testing.T) {
			assert := assert.New(t)
			t.Helper()
			setupTestCases()

			currentTime := time.Now().Add(-1 * time.Hour)
			currentTimeString, _ := currentTime.UTC().MarshalText()
			readFileReturnValue = []byte(fmt.Sprintf("{\"AccessKeyId\":\"access-key-id\",\"Expiration\":\"%s\",\"SecretAccessKey\":\"secret-access-key\",\"SessionToken\":\"session-token\"}", string(currentTimeString)))

			client := cache.New(&testCredentialsCacheOsClientMock{}, "test-path")

			if v.AwsVault == nil {
				client.Vault(v.Vault)
				client.Item(v.Item)
			} else {
				client.GenerateFromOP(v.AwsVault)
			}
			if v.OpAws == nil {
				client.MFA(v.Mfa)
				client.AssumeRole(v.AssumeRole)
			} else {
				client.GenerateFromOPAWS(v.OpAws)
			}

			credentials, err := client.GetCache()

			assert.Nil(err)
			assert.Nil(credentials)

			assert.Equal("test-path", statInput[0], "First check is checking if directory path is existing")
			assert.Equal(0, mkdirAllCallCount, "MkdirAll should not be called, because the Stat function is not returning an error")
			assert.Equal(v.ExpectedFileName, statInput[1], "Second is checking the filepath for cache file is existing")
			assert.Equal(1, readFileCallCount, "ReadFile should be called")

			assert.Equal(v.ExpectedFileName, readFileInput)
			assert.Equal(1, removeCallCount, "Remove call should not be executed, since the credentials are still valid")
			assert.Equal(v.ExpectedFileName, removeInput)
		})
	}
}

func TestStore(t *testing.T) {
	t.Helper()
	credentials := &sts.Credentials{}

	for i, v := range testCasesGetCache {
		t.Run(fmt.Sprintf("Running GetCache test with valid credentials %d", i), func(t *testing.T) {
			assert := assert.New(t)
			t.Helper()
			setupTestCases()

			client := cache.New(&testCredentialsCacheOsClientMock{}, "test-path")

			if v.AwsVault == nil {
				client.Vault(v.Vault)
				client.Item(v.Item)
			} else {
				client.GenerateFromOP(v.AwsVault)
			}
			if v.OpAws == nil {
				client.MFA(v.Mfa)
				client.AssumeRole(v.AssumeRole)
			} else {
				client.GenerateFromOPAWS(v.OpAws)
			}

			err := client.Store(credentials)

			assert.Nil(err)

			stsCredentials, err := json.Marshal(credentials)

			assert.Equal("test-path", statInput[0], "First check is checking if directory path is existing")
			assert.Equal(0, mkdirAllCallCount, "MkdirAll should not be called, because the Stat function is not returning an error")
			assert.Equal(1, writeFileCallCount, "WriteFile should be called")

			assert.Equal(v.ExpectedFileName, writeFileInput.filename)
			assert.Equal(stsCredentials, writeFileInput.data)
		})
	}
}
