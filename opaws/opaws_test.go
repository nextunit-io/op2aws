package opaws

import (
	"fmt"
	"nextunit/op2aws/awsvault"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
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
	OpAWSInput
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

	client := New(&awsVaultTest{}, &opAwsInputTest{})

	client.AssumeRole("test-assume-role")
	_, err := client.GetCredentials()
	if err != nil {
		t.Errorf("Error occured at GetCredentials: %+v", err)
	}
}

func TestUsingAssumeRoleAndCallingTheAwsApi(t *testing.T) {
	setupTestCase()
	t.Helper()

	client := New(&awsVaultTest{}, &opAwsInputTest{})

	client.AssumeRole("test-assume-role")
	client.GetCredentials()

	if assumeRoleCallCount != 1 {
		t.Errorf("AssumeRole was not executed - Expected: %d - Actual %d", 1, assumeRoleCallCount)
	}
}

func TestUsingAssumeRoleWithCorrectAssumeRoleInput(t *testing.T) {
	setupTestCase()
	t.Helper()

	client := New(&awsVaultTest{}, &opAwsInputTest{})

	client.AssumeRole("test-assume-role")
	client.GetCredentials()

	if *assumeRoleInput.RoleArn != "test-assume-role" {
		t.Errorf("Input for RoleArn is not valid - Expected %s - Actual %s", "test-assume-role", *assumeRoleInput.RoleArn)
	}

	if *assumeRoleInput.RoleSessionName != DEFAULT_SESSION_NAME {
		t.Errorf("Input for RoleSessionName is not valid - Expected %s - Actual %s", DEFAULT_SESSION_NAME, *assumeRoleInput.RoleSessionName)
	}

	if assumeRoleInput.SerialNumber != nil {
		t.Errorf("SerialNumber should be 'nil' - Acutal: %s", *assumeRoleInput.SerialNumber)
	}
}

func TestUsingAssumeRoleAndMfaWithCorrectAssumeRoleInput(t *testing.T) {
	setupTestCase()
	t.Helper()

	client := New(&awsVaultTest{}, &opAwsInputTest{})

	client.AssumeRole("test-assume-role")
	client.UseMFA("test-mfa")
	client.GetCredentials()

	if *assumeRoleInput.RoleArn != "test-assume-role" {
		t.Errorf("Input for RoleArn is not valid - Expected %s - Actual %s", "test-assume-role", *assumeRoleInput.RoleArn)
	}

	if *assumeRoleInput.RoleSessionName != DEFAULT_SESSION_NAME {
		t.Errorf("Input for RoleSessionName is not valid - Expected %s - Actual %s", DEFAULT_SESSION_NAME, *assumeRoleInput.RoleSessionName)
	}

	if getOtpCallCount != 1 {
		t.Errorf("GetOTP function was not called but was expected to")
	}

	if *assumeRoleInput.SerialNumber != "test-mfa" {
		t.Errorf("Input for SerialNumber is not valid - Expected %s - Actual %s", "test-mfa", *assumeRoleInput.SerialNumber)
	}

	if *assumeRoleInput.TokenCode != "otp-default" {
		t.Errorf("Input for SerialNumber is not valid - Expected %s - Actual %s", "otp-default", *assumeRoleInput.TokenCode)
	}
}

func TestUsingAssumeRoleAndMfaWithAssumeRoleError(t *testing.T) {
	setupTestCase()
	t.Helper()

	assumeRoleReturnValue = nil
	client := New(&awsVaultTest{}, &opAwsInputTest{})

	client.AssumeRole("test-assume-role")
	client.UseMFA("test-mfa")
	_, err := client.GetCredentials()

	if assumeRoleCallCount != 1 {
		t.Errorf("AssumeRole was not executed - Expected: %d - Actual %d", 1, assumeRoleCallCount)
	}
	if err == nil {
		t.Errorf("Expacting an error, when AssumeRole is returning an error")
	}
}

func TestUsingAssumeRoleAndMfaWithOtpError(t *testing.T) {
	setupTestCase()
	t.Helper()

	getOtpReturnValue = nil
	client := New(&awsVaultTest{}, &opAwsInputTest{})

	client.AssumeRole("test-assume-role")
	client.UseMFA("test-mfa")
	_, err := client.GetCredentials()

	if assumeRoleCallCount != 0 {
		t.Errorf("AssumeRole was executed - Expected: %d - Actual %d", 0, assumeRoleCallCount)
	}
	if err == nil {
		t.Errorf("Expacting an error, when GetOtp is returning an error")
	}
}

func TestUsingAssumeRoleAndMfaWithSecretAccessKeyError(t *testing.T) {
	setupTestCase()
	t.Helper()

	getSecretAccessKeyReturnValue = nil
	client := New(&awsVaultTest{}, &opAwsInputTest{})

	client.AssumeRole("test-assume-role")
	client.UseMFA("test-mfa")
	_, err := client.GetCredentials()

	if assumeRoleCallCount != 0 {
		t.Errorf("AssumeRole was executed - Expected: %d - Actual %d", 0, assumeRoleCallCount)
	}
	if getOtpCallCount != 0 {
		t.Errorf("GetOtp was executed - Expected: %d - Actual %d", 0, getOtpCallCount)
	}
	if err == nil {
		t.Errorf("Expacting an error, when GetSecretAccessKey is returning an error")
	}
}

func TestUsingAssumeRoleAndMfaWithAccessKeyIdError(t *testing.T) {
	setupTestCase()
	t.Helper()

	getAccessKeyIdReturnValue = nil
	client := New(&awsVaultTest{}, &opAwsInputTest{})

	client.AssumeRole("test-assume-role")
	client.UseMFA("test-mfa")
	_, err := client.GetCredentials()

	if assumeRoleCallCount != 0 {
		t.Errorf("AssumeRole was executed - Expected: %d - Actual %d", 0, assumeRoleCallCount)
	}
	if getOtpCallCount != 0 {
		t.Errorf("GetOtp was executed - Expected: %d - Actual %d", 0, getOtpCallCount)
	}
	if getSecretAccessKeyCallCount != 0 {
		t.Errorf("GetSecretAccessKey was executed - Expected: %d - Actual %d", 0, getSecretAccessKeyCallCount)
	}
	if err == nil {
		t.Errorf("Expacting an error, when GetAccessKeyId is returning an error")
	}
}

func TestUsingGenerateSessionTokenWithoutErrors(t *testing.T) {
	setupTestCase()
	t.Helper()

	client := New(&awsVaultTest{}, &opAwsInputTest{})

	_, err := client.GetCredentials()
	if err != nil {
		t.Errorf("Error occured at GetCredentials: %+v", err)
	}
}

func TestUsingGenerateSessionTokenAndCallingTheAwsApi(t *testing.T) {
	setupTestCase()
	t.Helper()

	client := New(&awsVaultTest{}, &opAwsInputTest{})

	client.GetCredentials()

	if getSessionTokenCallCount != 1 {
		t.Errorf("AssumeRole was not executed - Expected: %d - Actual %d", 1, getSessionTokenCallCount)
	}
}

func TestUsingGenerateSessionTokenWithCorrectAssumeRoleInput(t *testing.T) {
	setupTestCase()
	t.Helper()

	client := New(&awsVaultTest{}, &opAwsInputTest{})

	client.GetCredentials()

	if getSessionTokenInput.SerialNumber != nil {
		t.Errorf("SerialNumber should be 'nil' - Acutal: %s", *getSessionTokenInput.SerialNumber)
	}
}

func TestUsingGenerateSessionTokenAndMfaWithCorrectAssumeRoleInput(t *testing.T) {
	setupTestCase()
	t.Helper()

	client := New(&awsVaultTest{}, &opAwsInputTest{})

	client.UseMFA("test-mfa")
	client.GetCredentials()

	if getOtpCallCount != 1 {
		t.Errorf("GetOTP function was not called but was expected to")
	}

	if *getSessionTokenInput.SerialNumber != "test-mfa" {
		t.Errorf("Input for SerialNumber is not valid - Expected %s - Actual %s", "test-mfa", *getSessionTokenInput.SerialNumber)
	}

	if *getSessionTokenInput.TokenCode != "otp-default" {
		t.Errorf("Input for SerialNumber is not valid - Expected %s - Actual %s", "otp-default", *getSessionTokenInput.TokenCode)
	}
}

func TestUsingGenerateSessionTokenAndMfaWithAssumeRoleError(t *testing.T) {
	setupTestCase()
	t.Helper()

	getSessionTokenReturnValue = nil
	client := New(&awsVaultTest{}, &opAwsInputTest{})

	client.UseMFA("test-mfa")
	_, err := client.GetCredentials()

	if getSessionTokenCallCount != 1 {
		t.Errorf("GetSessionToken was not executed - Expected: %d - Actual %d", 1, getSessionTokenCallCount)
	}
	if err == nil {
		t.Errorf("Expacting an error, when GetSessionToken is returning an error")
	}
}

func TestUsingGenerateSessionTokenAndMfaWithOtpError(t *testing.T) {
	setupTestCase()
	t.Helper()

	getOtpReturnValue = nil
	client := New(&awsVaultTest{}, &opAwsInputTest{})

	client.UseMFA("test-mfa")
	_, err := client.GetCredentials()

	if getSessionTokenCallCount != 0 {
		t.Errorf("GetSessionToken was executed - Expected: %d - Actual %d", 0, getSessionTokenCallCount)
	}
	if err == nil {
		t.Errorf("Expacting an error, when GetOtp is returning an error")
	}
}

func TestUsingGenerateSessionTokenAndMfaWithSecretAccessKeyError(t *testing.T) {
	setupTestCase()
	t.Helper()

	getSecretAccessKeyReturnValue = nil
	client := New(&awsVaultTest{}, &opAwsInputTest{})

	client.UseMFA("test-mfa")
	_, err := client.GetCredentials()

	if getSessionTokenCallCount != 0 {
		t.Errorf("GetSessionToken was executed - Expected: %d - Actual %d", 0, getSessionTokenCallCount)
	}
	if getOtpCallCount != 0 {
		t.Errorf("GetOtp was executed - Expected: %d - Actual %d", 0, getOtpCallCount)
	}
	if err == nil {
		t.Errorf("Expacting an error, when GetSecretAccessKey is returning an error")
	}
}

func TestUsingGenerateSessionTokenAndMfaWithAccessKeyIdError(t *testing.T) {
	setupTestCase()
	t.Helper()

	getAccessKeyIdReturnValue = nil
	client := New(&awsVaultTest{}, &opAwsInputTest{})

	client.AssumeRole("test-assume-role")
	client.UseMFA("test-mfa")
	_, err := client.GetCredentials()

	if getSessionTokenCallCount != 0 {
		t.Errorf("GetSessionToken was executed - Expected: %d - Actual %d", 0, getSessionTokenCallCount)
	}
	if getOtpCallCount != 0 {
		t.Errorf("GetOtp was executed - Expected: %d - Actual %d", 0, getOtpCallCount)
	}
	if getSecretAccessKeyCallCount != 0 {
		t.Errorf("GetSecretAccessKey was executed - Expected: %d - Actual %d", 0, getSecretAccessKeyCallCount)
	}
	if err == nil {
		t.Errorf("Expacting an error, when GetAccessKeyId is returning an error")
	}
}
