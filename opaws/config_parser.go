package opaws

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"
)

type ConfigParserInterface interface {
	Parse(data string) (*AWSConfigModel, error)
}

type ConfigParser struct {
	ConfigParserInterface
}

func (c ConfigParser) Parse(data string) (*AWSConfigModel, error) {
	regexp := regexp.MustCompile("(\\[[^\\]]*\\][^\\[]*)")
	profileStrings := regexp.FindAll([]byte(data), -1)
	profiles := &AWSConfigModel{
		Profile: map[string]AWSConfigCredentialsModel{},
	}

	for _, v := range profileStrings {
		name, p, err := c.parseProfile(string(v))
		if err != nil {
			return nil, err
		}

		profiles.Profile[name] = *p
	}

	return profiles, nil
}

func (ConfigParser) parseProfile(data string) (string, *AWSConfigCredentialsModel, error) {
	returnValue := &AWSConfigCredentialsModel{}

	r := regexp.MustCompile("\\[([^\\]]*)\\]([^\\[]*)$")
	result := r.FindSubmatch([]byte(data))
	name := string(result[1])
	body := result[2]

	scanner := bufio.NewScanner(strings.NewReader(string(body)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		regex := regexp.MustCompile("([^=]*)=(.*)$")
		res := regex.FindSubmatch([]byte(line))

		value := strings.TrimSpace(string(res[2]))

		switch t := strings.TrimSpace(string(res[1])); t {
		case "role_arn":
			returnValue.RoleArn = value
		case "source_profile":
			returnValue.SourceProfile = value
		case "credential_process":
			returnValue.CredentialProcess = value
		case "aws_access_key_id":
			returnValue.AwsAccessKeyId = value
		case "aws_secret_access_key":
			returnValue.AwsSecretAccessKey = value
		case "aws_session_token":
			returnValue.AwsSessionToken = value
		case "aws_security_token":
			returnValue.AwsSecurityToken = value
		case "x_principal_arn":
			returnValue.XPrincipalArn = value
		case "x_security_token_expires":
			returnValue.XSecurityTokenExpires = value
		default:
			fmt.Println("Unknown: " + t)
		}
	}

	return name, returnValue, nil
}
