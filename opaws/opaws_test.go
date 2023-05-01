package opaws_test

import (
	"fmt"
	"nextunit/op2aws/awsvault"
	"nextunit/op2aws/opaws"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
	"github.com/stretchr/testify/assert"
)

var (
	getAccessKeyIdReturnValue     *string
	getSecretAccessKeyReturnValue *string
	getOtpReturnValue             *string
	assumeRoleReturnValue         *sts.AssumeRoleOutput
	getSessionTokenReturnValue    *sts.GetSessionTokenOutput

	getAccessKeyIdCallCount     int
	getSecretAccessKeyCallCount int
	getOtpCallCount             int
	assumeRoleCallCount         int
	getSessionTokenCallCount    int

	assumeRoleInput      *sts.AssumeRoleInput
	getSessionTokenInput *sts.GetSessionTokenInput
)

type awsVaultTest struct {
	awsvault.Vault
}

type opAwsInputTest struct {
	opaws.OpAWSInput
}

type stsApiTest struct {
	stsiface.STSAPI
}

func setupTestCase() {
	// Set default
	getAccessKeyIdReturnValueString := "access-key-id-default"
	getSecretAccessKeyReturnValueString := "secret-access-key-default"
	getOtpReturnValueString := "otp-default"
	sourceIdentityString := "test-source-identity"

	getAccessKeyIdReturnValue = &getAccessKeyIdReturnValueString
	getSecretAccessKeyReturnValue = &getSecretAccessKeyReturnValueString
	getOtpReturnValue = &getOtpReturnValueString
	assumeRoleReturnValue = &sts.AssumeRoleOutput{
		SourceIdentity: &sourceIdentityString,
	}
	getSessionTokenReturnValue = &sts.GetSessionTokenOutput{}

	getAccessKeyIdCallCount = 0
	getSecretAccessKeyCallCount = 0
	getOtpCallCount = 0
	assumeRoleCallCount = 0
	getSessionTokenCallCount = 0

	assumeRoleInput = nil
	getSessionTokenInput = nil
}

func (stsApiTest) AssumeRole(input *sts.AssumeRoleInput) (*sts.AssumeRoleOutput, error) {
	assumeRoleInput = input

	assumeRoleCallCount++
	if assumeRoleReturnValue == nil {
		return nil, fmt.Errorf("Test error")
	}
	return assumeRoleReturnValue, nil
}

func (stsApiTest) GetSessionToken(input *sts.GetSessionTokenInput) (*sts.GetSessionTokenOutput, error) {
	getSessionTokenInput = input

	getSessionTokenCallCount++
	if getSessionTokenReturnValue == nil {
		return nil, fmt.Errorf("Test error")
	}
	return getSessionTokenReturnValue, nil
}

func (awsVaultTest) GetAccessKeyId() (string, error) {
	getAccessKeyIdCallCount++
	if getAccessKeyIdReturnValue == nil {
		return "", fmt.Errorf("Test error")
	}
	return *getAccessKeyIdReturnValue, nil
}

func (awsVaultTest) GetSecretAccessKey() (string, error) {
	getSecretAccessKeyCallCount++
	if getSecretAccessKeyReturnValue == nil {
		return "", fmt.Errorf("Test error")
	}
	return *getSecretAccessKeyReturnValue, nil
}

func (awsVaultTest) GetOTP() (string, error) {
	getOtpCallCount++
	if getOtpReturnValue == nil {
		return "", fmt.Errorf("Test error")
	}
	return *getOtpReturnValue, nil
}

func (opAwsInputTest) NewSession(cfgs ...*aws.Config) *session.Session {
	return session.New()
}

func (opAwsInputTest) NewSts(p client.ConfigProvider, cfgs ...*aws.Config) stsiface.STSAPI {
	return &stsApiTest{}
}

func TestUsingAssumeRoleWithoutErrors(t *testing.T) {
	setupTestCase()
	t.Helper()

	client := opaws.New(&awsVaultTest{}, &opAwsInputTest{})

	client.AssumeRole("test-assume-role")
	_, err := client.GetCredentials()

	assert.Nil(t, err, "Error occured at AssumeRole")
}

func TestUsingAssumeRoleAndCallingTheAwsApi(t *testing.T) {
	setupTestCase()
	t.Helper()

	client := opaws.New(&awsVaultTest{}, &opAwsInputTest{})

	client.AssumeRole("test-assume-role")
	client.GetCredentials()

	assert.Equal(t, 1, assumeRoleCallCount, "AssumeRole should called one time")
}

func TestUsingAssumeRoleWithCorrectAssumeRoleInput(t *testing.T) {
	setupTestCase()
	assert := assert.New(t)
	t.Helper()

	client := opaws.New(&awsVaultTest{}, &opAwsInputTest{})

	client.AssumeRole("test-assume-role")
	client.GetCredentials()

	assert.Equalf("test-assume-role", *assumeRoleInput.RoleArn, "Input for RoleArn is not valid - Expected %s - Actual %s", "test-assume-role", *assumeRoleInput.RoleArn)
	assert.Equalf(opaws.DEFAULT_SESSION_NAME, *assumeRoleInput.RoleSessionName, "Input for RoleArn is not valid - Expected %s - Actual %s", opaws.DEFAULT_SESSION_NAME, *assumeRoleInput.RoleSessionName)
	assert.Nilf(assumeRoleInput.SerialNumber, "SerialNumber should be 'nil' - Acutal: %s", assumeRoleInput.SerialNumber)
}

func TestUsingAssumeRoleAndMfaWithCorrectAssumeRoleInput(t *testing.T) {
	setupTestCase()
	assert := assert.New(t)

	t.Helper()

	client := opaws.New(&awsVaultTest{}, &opAwsInputTest{})

	client.AssumeRole("test-assume-role")
	client.UseMFA("test-mfa")
	client.GetCredentials()

	assert.Equalf("test-assume-role", *assumeRoleInput.RoleArn, "Input for RoleArn is not valid - Expected %s - Actual %s", "test-assume-role", *assumeRoleInput.RoleArn)
	assert.Equalf(opaws.DEFAULT_SESSION_NAME, *assumeRoleInput.RoleSessionName, "Input for RoleArn is not valid - Expected %s - Actual %s", opaws.DEFAULT_SESSION_NAME, *assumeRoleInput.RoleSessionName)

	assert.Equalf(1, getOtpCallCount, "GetOTP function should be exactly one time called. Called: %d", getOtpCallCount)
	assert.Equalf("test-mfa", *assumeRoleInput.SerialNumber, "Input for SerialNumber is not valid - Expected %s - Actual %s", "test-mfa", *assumeRoleInput.SerialNumber)
	assert.Equalf("otp-default", *assumeRoleInput.TokenCode, "Input for TokenCode is not valid - Expected %s - Actual %s", "otp-default", *assumeRoleInput.TokenCode)
}

func TestUsingAssumeRoleAndMfaWithAssumeRoleError(t *testing.T) {
	setupTestCase()
	t.Helper()

	assumeRoleReturnValue = nil
	client := opaws.New(&awsVaultTest{}, &opAwsInputTest{})

	client.AssumeRole("test-assume-role")
	client.UseMFA("test-mfa")
	_, err := client.GetCredentials()

	assert.Equal(t, 1, assumeRoleCallCount, "AssumeRole should called one time")
	assert.ErrorContains(t, err, "Test error")
}

func TestUsingAssumeRoleAndMfaWithOtpError(t *testing.T) {
	setupTestCase()
	t.Helper()

	getOtpReturnValue = nil
	client := opaws.New(&awsVaultTest{}, &opAwsInputTest{})

	client.AssumeRole("test-assume-role")
	client.UseMFA("test-mfa")
	_, err := client.GetCredentials()

	assert.Equal(t, 0, assumeRoleCallCount, "AssumeRole should not be called")
	assert.Equal(t, 1, getOtpCallCount, "GetOtp should be called")
	assert.ErrorContains(t, err, "Test error")
}

func TestUsingAssumeRoleAndMfaWithSecretAccessKeyError(t *testing.T) {
	setupTestCase()
	t.Helper()

	getSecretAccessKeyReturnValue = nil
	client := opaws.New(&awsVaultTest{}, &opAwsInputTest{})

	client.AssumeRole("test-assume-role")
	client.UseMFA("test-mfa")
	_, err := client.GetCredentials()

	assert.Equal(t, 0, assumeRoleCallCount, "AssumeRole should not be called")
	assert.Equal(t, 0, getOtpCallCount, "GetOtp should not be called")
	assert.Equal(t, 1, getSecretAccessKeyCallCount, "GetSecretAccessKey should be called")
	assert.Equal(t, 1, getAccessKeyIdCallCount, "GetAccessKeyId should be called")
	assert.ErrorContains(t, err, "Test error")
}

func TestUsingAssumeRoleAndMfaWithAccessKeyIdError(t *testing.T) {
	setupTestCase()
	t.Helper()

	getAccessKeyIdReturnValue = nil
	client := opaws.New(&awsVaultTest{}, &opAwsInputTest{})

	client.AssumeRole("test-assume-role")
	client.UseMFA("test-mfa")
	_, err := client.GetCredentials()

	assert.Equal(t, 0, assumeRoleCallCount, "AssumeRole should not be called")
	assert.Equal(t, 0, getOtpCallCount, "GetOtp should not be called")
	assert.Equal(t, 0, getSecretAccessKeyCallCount, "GetSecretAccessKey should not be called")
	assert.Equal(t, 1, getAccessKeyIdCallCount, "GetAccessKeyId should be called")
	assert.ErrorContains(t, err, "Test error")
}

func TestUsingGenerateSessionTokenWithoutErrors(t *testing.T) {
	setupTestCase()
	t.Helper()

	client := opaws.New(&awsVaultTest{}, &opAwsInputTest{})

	_, err := client.GetCredentials()
	assert.Nil(t, err, "Error occured at GenerateSessionToken")
}

func TestUsingGenerateSessionTokenAndCallingTheAwsApi(t *testing.T) {
	setupTestCase()
	t.Helper()

	client := opaws.New(&awsVaultTest{}, &opAwsInputTest{})

	client.GetCredentials()
	assert.Equal(t, 1, getSessionTokenCallCount, "GetSessionToken should be called")
}

func TestUsingGenerateSessionTokenWithCorrectAssumeRoleInput(t *testing.T) {
	setupTestCase()
	t.Helper()

	client := opaws.New(&awsVaultTest{}, &opAwsInputTest{})

	client.GetCredentials()
	assert.Nil(t, getSessionTokenInput.SerialNumber, "SerialNumber should not be set")
}

func TestUsingGenerateSessionTokenAndMfaWithCorrectAssumeRoleInput(t *testing.T) {
	setupTestCase()
	assert := assert.New(t)
	t.Helper()

	client := opaws.New(&awsVaultTest{}, &opAwsInputTest{})

	client.UseMFA("test-mfa")
	client.GetCredentials()

	assert.Equalf(1, getOtpCallCount, "GetOTP function should be exactly one time called. Called: %d", getOtpCallCount)
	assert.Equalf("test-mfa", *getSessionTokenInput.SerialNumber, "Input for SerialNumber is not valid - Expected %s - Actual %s", "test-mfa", *getSessionTokenInput.SerialNumber)
	assert.Equalf("otp-default", *getSessionTokenInput.TokenCode, "Input for TokenCode is not valid - Expected %s - Actual %s", "otp-default", *getSessionTokenInput.TokenCode)
}

func TestUsingGenerateSessionTokenAndMfaWithGetSessionTokenError(t *testing.T) {
	setupTestCase()
	t.Helper()

	getSessionTokenReturnValue = nil
	client := opaws.New(&awsVaultTest{}, &opAwsInputTest{})

	client.UseMFA("test-mfa")
	_, err := client.GetCredentials()

	assert.Equal(t, 1, getSessionTokenCallCount, "GetSessionToken should be called")
	assert.Equal(t, 1, getOtpCallCount, "GetOtp should be called")
	assert.Equal(t, 1, getSecretAccessKeyCallCount, "GetSecretAccessKey should be called")
	assert.Equal(t, 1, getAccessKeyIdCallCount, "GetAccessKeyId should be called")
	assert.ErrorContains(t, err, "Test error")
}

func TestUsingGenerateSessionTokenAndMfaWithOtpError(t *testing.T) {
	setupTestCase()
	t.Helper()

	getOtpReturnValue = nil
	client := opaws.New(&awsVaultTest{}, &opAwsInputTest{})

	client.UseMFA("test-mfa")
	_, err := client.GetCredentials()

	assert.Equal(t, 0, getSessionTokenCallCount, "GetSessionToken should not be called")
	assert.Equal(t, 1, getOtpCallCount, "GetOtp should be called")
	assert.Equal(t, 1, getSecretAccessKeyCallCount, "GetSecretAccessKey should be called")
	assert.Equal(t, 1, getAccessKeyIdCallCount, "GetAccessKeyId should be called")
	assert.ErrorContains(t, err, "Test error")
}

func TestUsingGenerateSessionTokenAndMfaWithSecretAccessKeyError(t *testing.T) {
	setupTestCase()
	t.Helper()

	getSecretAccessKeyReturnValue = nil
	client := opaws.New(&awsVaultTest{}, &opAwsInputTest{})

	client.UseMFA("test-mfa")
	_, err := client.GetCredentials()

	assert.Equal(t, 0, getSessionTokenCallCount, "GetSessionToken should not be called")
	assert.Equal(t, 0, getOtpCallCount, "GetOtp should not be called")
	assert.Equal(t, 1, getSecretAccessKeyCallCount, "GetSecretAccessKey should be called")
	assert.Equal(t, 1, getAccessKeyIdCallCount, "GetAccessKeyId should be called")
	assert.ErrorContains(t, err, "Test error")
}

func TestUsingGenerateSessionTokenAndMfaWithAccessKeyIdError(t *testing.T) {
	setupTestCase()
	t.Helper()

	getAccessKeyIdReturnValue = nil
	client := opaws.New(&awsVaultTest{}, &opAwsInputTest{})

	client.AssumeRole("test-assume-role")
	client.UseMFA("test-mfa")
	_, err := client.GetCredentials()

	assert.Equal(t, 0, getSessionTokenCallCount, "GetSessionToken should not be called")
	assert.Equal(t, 0, getOtpCallCount, "GetOtp should not be called")
	assert.Equal(t, 0, getSecretAccessKeyCallCount, "GetSecretAccessKey should not be called")
	assert.Equal(t, 1, getAccessKeyIdCallCount, "GetAccessKeyId should be called")
	assert.ErrorContains(t, err, "Test error")
}
