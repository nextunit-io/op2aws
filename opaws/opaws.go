package opaws

import (
	"nextunit/op2aws/awsvault"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
)

var DEFAULT_SESSION_NAME = "op2aws-session"

type OpAWS struct {
	opClient  awsvault.Vault
	awsClient OpAWSInput

	mfa         string
	assume_role string
}

func (client OpAWS) generateStsClient() (stsiface.STSAPI, error) {
	accessKeyId, err := client.opClient.GetAccessKeyId()
	if err != nil {
		return nil, err
	}

	secretAccessKey, err := client.opClient.GetSecretAccessKey()
	if err != nil {
		return nil, err
	}

	return client.awsClient.NewSts(client.awsClient.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(accessKeyId, secretAccessKey, ""),
	})), nil
}

func (client OpAWS) generateSessionToken() (*sts.Credentials, error) {
	stsClient, err := client.generateStsClient()
	if err != nil {
		return nil, err
	}

	input := &sts.GetSessionTokenInput{}

	if len(client.mfa) != 0 {
		otp, err := client.opClient.GetOTP()
		if err != nil {
			return nil, err
		}

		input.SerialNumber = &client.mfa
		input.TokenCode = &otp
	}

	output, err := stsClient.GetSessionToken(input)
	if err != nil {
		return nil, err
	}

	return output.Credentials, nil
}

func (client OpAWS) generateAssumedRoleCredentials() (*sts.Credentials, error) {
	stsClient, err := client.generateStsClient()
	if err != nil {
		return nil, err
	}

	input := &sts.AssumeRoleInput{RoleArn: &client.assume_role, RoleSessionName: &DEFAULT_SESSION_NAME}

	if len(client.mfa) != 0 {
		otp, err := client.opClient.GetOTP()
		if err != nil {
			return nil, err
		}

		input.SerialNumber = &client.mfa
		input.TokenCode = &otp
	}

	output, err := stsClient.AssumeRole(input)
	if err != nil {
		return nil, err
	}

	return output.Credentials, nil
}

// TODO: Missing - static credentials
func (client OpAWS) GetCredentials() (*sts.Credentials, error) {
	if len(client.assume_role) == 0 {
		return client.generateSessionToken()
	}

	return client.generateAssumedRoleCredentials()
}

func (client OpAWS) GetMFA() string {
	return client.mfa
}

func (client OpAWS) GetAssumeRole() string {
	return client.assume_role
}

func (client *OpAWS) UseMFA(mfa string) {
	client.mfa = mfa
}

func (client *OpAWS) AssumeRole(assume_role string) {
	client.assume_role = assume_role
}

func New(opClient awsvault.Vault, awsClient OpAWSInput) *OpAWS {
	return &OpAWS{opClient: opClient, awsClient: awsClient}
}
