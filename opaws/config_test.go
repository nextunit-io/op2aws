package opaws_test

import (
	"fmt"
	"io/fs"
	"nextunit/op2aws/awsvault"
	"nextunit/op2aws/opaws"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testGetProfileInput struct {
	profileName          string
	vault                string
	item                 string
	assumeRole           string
	mfa                  string
	labelAccessKey       string
	labelSecretAccessKey string

	expectedOutput string
}

type testAwsConfigMock struct {
	opaws.AwsConfigInterface
}

type testAwsFileInfoMock struct {
	fs.FileInfo
}

type writeFileInputModel struct {
	filename string
	data     []byte
	perm     fs.FileMode
}

type openFileInputModel struct {
	name string
	flag int
	perm fs.FileMode
}

type awsConfigFileMock struct {
	opaws.AwsConfigFileInterface
}

var (
	fileInfoReturnValue        fs.FileInfo
	errorIsNotExistReturnValue bool
	writeFileReturnValue       error
	openFileReturnValue        opaws.AwsConfigFileInterface
	closeReturnValue           error
	writeStringReturnValue     int

	fileInfoCallCount        int
	errorIsNotExistCallCount int
	writeFileCallCount       int
	openFileCallCount        int
	closeCallCount           int
	writeStringCallCount     int

	fileInfoInput        string
	errorIsNotExistInput error
	writeFileInput       []writeFileInputModel
	openFileInput        []openFileInputModel
	writeStringInput     string

	testCases = []testGetProfileInput{
		{
			profileName:          "test-profile",
			vault:                "test-vault",
			item:                 "test-item",
			assumeRole:           "testAssumeRole",
			mfa:                  "testMfa",
			labelAccessKey:       "testLabelAccessKey",
			labelSecretAccessKey: "testLabelSecretAccessKey",
			expectedOutput:       "\n\n[profile test-profile]\n    credential_process = sh -c '\"op2aws\" \"cli\" \"test-vault\" \"test-item\" \"-a\" \"testAssumeRole\" \"-m\" \"testMfa\" \"-k\" \"testLabelAccessKey\" \"-s\" \"testLabelSecretAccessKey\"'",
		},
		{
			profileName:          "test-profile",
			vault:                "test-vault",
			item:                 "test-item",
			assumeRole:           "testAssumeRole",
			mfa:                  "testMfa",
			labelAccessKey:       awsvault.AWS_ACCESS_KEY_FIELD_DEFAULT,
			labelSecretAccessKey: awsvault.AWS_SECRET_ACCESS_KEY_FIELD_DEFAULT,
			expectedOutput:       "\n\n[profile test-profile]\n    credential_process = sh -c '\"op2aws\" \"cli\" \"test-vault\" \"test-item\" \"-a\" \"testAssumeRole\" \"-m\" \"testMfa\"'",
		},
		{
			profileName:    "test-profile",
			vault:          "test-vault",
			item:           "test-item",
			assumeRole:     "testAssumeRole",
			mfa:            "testMfa",
			expectedOutput: "\n\n[profile test-profile]\n    credential_process = sh -c '\"op2aws\" \"cli\" \"test-vault\" \"test-item\" \"-a\" \"testAssumeRole\" \"-m\" \"testMfa\"'",
		},
		{
			profileName:    "test-profile",
			vault:          "test-vault",
			item:           "test-item",
			assumeRole:     "testAssumeRole",
			expectedOutput: "\n\n[profile test-profile]\n    credential_process = sh -c '\"op2aws\" \"cli\" \"test-vault\" \"test-item\" \"-a\" \"testAssumeRole\"'",
		},
		{
			profileName:    "test-profile",
			vault:          "test-vault",
			item:           "test-item",
			mfa:            "testMfa",
			expectedOutput: "\n\n[profile test-profile]\n    credential_process = sh -c '\"op2aws\" \"cli\" \"test-vault\" \"test-item\" \"-m\" \"testMfa\"'",
		},
		{
			profileName:    "test-profile",
			vault:          "test-vault",
			item:           "test-item",
			expectedOutput: "\n\n[profile test-profile]\n    credential_process = sh -c '\"op2aws\" \"cli\" \"test-vault\" \"test-item\"'",
		},
	}
)

func setupTestCases() {
	// Set defaults
	fileInfoReturnValue = &testAwsFileInfoMock{}
	errorIsNotExistReturnValue = false
	writeFileReturnValue = nil
	openFileReturnValue = &awsConfigFileMock{}
	closeReturnValue = nil
	writeStringReturnValue = 2

	fileInfoCallCount = 0
	errorIsNotExistCallCount = 0
	writeFileCallCount = 0
	openFileCallCount = 0
	closeCallCount = 0
	writeStringCallCount = 0

	fileInfoInput = ""
	errorIsNotExistInput = nil
	writeFileInput = []writeFileInputModel{}
	openFileInput = []openFileInputModel{}
	writeStringInput = ""
}

func (awsConfigFileMock) Close() error {
	closeCallCount++
	return closeReturnValue
}

func (awsConfigFileMock) WriteString(s string) (n int, err error) {
	writeStringCallCount++
	writeStringInput = s
	if writeStringReturnValue == -1 {
		return -1, fmt.Errorf("test error WriteString")
	}

	return writeStringReturnValue, nil
}

func (testAwsConfigMock) Stat(name string) (fs.FileInfo, error) {
	fileInfoCallCount++
	fileInfoInput = name
	if fileInfoReturnValue == nil {
		return nil, fmt.Errorf("test error Stat")
	}

	return fileInfoReturnValue, nil
}

func (testAwsConfigMock) IsNotExist(err error) bool {
	errorIsNotExistCallCount++
	errorIsNotExistInput = err

	return errorIsNotExistReturnValue
}

func (testAwsConfigMock) WriteFile(filename string, data []byte, perm fs.FileMode) error {
	writeFileCallCount++
	writeFileInput = append(writeFileInput, writeFileInputModel{
		filename: filename,
		data:     data,
		perm:     perm,
	})

	return writeFileReturnValue
}

func (testAwsConfigMock) OpenFile(name string, flag int, perm fs.FileMode) (opaws.AwsConfigFileInterface, error) {
	openFileCallCount++
	openFileInput = append(openFileInput, openFileInputModel{
		name: name,
		flag: flag,
		perm: perm,
	})

	if openFileReturnValue == nil {
		return nil, fmt.Errorf("test error OpenFile")
	}

	return openFileReturnValue, nil
}

func TestGetProfileBody(t *testing.T) {
	t.Helper()

	for i, v := range testCases {
		t.Run(fmt.Sprintf("Run case %d", i), func(t *testing.T) {
			output := opaws.GetProfileBody(
				v.profileName,
				v.vault,
				v.item,
				v.assumeRole,
				v.mfa,
				v.labelAccessKey,
				v.labelSecretAccessKey,
			)

			assert.Equal(t, v.expectedOutput, output)
		})
	}
}

func TestWriteProfileFileExists(t *testing.T) {
	assert := assert.New(t)
	t.Helper()
	setupTestCases()

	client := opaws.NewAwsConfig(&testAwsConfigMock{}, "test-path")

	err := client.WriteProfile("test-body")

	assert.Nil(err, "There should be no error")
	assert.Equal(1, fileInfoCallCount, "client.Stat should be called one time")
	assert.Equal(0, errorIsNotExistCallCount, "client.IsNotExist should not be called")
	assert.Equal(0, writeFileCallCount, "client.WriteFile should not be called")
	assert.Equal(1, openFileCallCount, "client.OpenFile should be called one time")
	assert.Equal(1, closeCallCount, "file.Close should be called one time")
	assert.Equal(1, writeStringCallCount, "file.WriteString should be called one time")

	assert.Equal("test-path", fileInfoInput)
	assert.Equal(openFileInputModel{
		name: "test-path",
		flag: os.O_APPEND | os.O_WRONLY,
		perm: 0644,
	}, openFileInput[0])
	assert.Equal("test-body", writeStringInput)
}
