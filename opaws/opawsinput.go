package opaws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
)

type OpAWSInput interface {
	NewSts(p client.ConfigProvider, cfgs ...*aws.Config) stsiface.STSAPI
	NewSession(cfgs ...*aws.Config) *session.Session
}

type OpAwsDefaultInput struct {
	OpAWSInput
}

func (OpAwsDefaultInput) NewSts(p client.ConfigProvider, cfgs ...*aws.Config) stsiface.STSAPI {
	return sts.New(p, cfgs...)
}

func (OpAwsDefaultInput) NewSession(cfgs ...*aws.Config) *session.Session {
	return session.New(cfgs...)
}
