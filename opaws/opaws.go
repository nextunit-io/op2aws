package opaws

import (
	"nextunit/op2aws/cache"
	"nextunit/op2aws/onepassword"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

type OpAWS struct {
	cacheClient *cache.CacheClient
	opClient    *onepassword.OnePassword
	mfa         string
	assume_role string
}

func (client OpAWS) generateStsClient() (*sts.STS, error) {
	accessKeyId, err := client.opClient.GetAccessKeyId()
	if err != nil {
		return nil, err
	}

	secretAccessKey, err := client.opClient.GetSecretAccessKey()
	if err != nil {
		return nil, err
	}

	return sts.New(session.New(&aws.Config{
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

	input := &sts.AssumeRoleInput{RoleArn: &client.assume_role}

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

func (client OpAWS) GetCredentials() (*sts.Credentials, error) {
	if len(client.assume_role) == 0 {
		return client.generateSessionToken()
	}

	return client.generateAssumedRoleCredentials()
}

func (client OpAWS) UseMFA(mfa string) {
	client.mfa = mfa
}

func (client OpAWS) AssumeRole(mfa string) {
	client.mfa = mfa
}

func New(cacheClient *cache.CacheClient, opClient *onepassword.OnePassword) *OpAWS {
	return &OpAWS{cacheClient: cacheClient, opClient: opClient}
}
