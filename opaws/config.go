package opaws

import (
	"fmt"
	"io/ioutil"
	"nextunit/op2aws/awsvault"
	"nextunit/op2aws/config"
	"os"
	"strings"
)

type AWSConfig struct {
	path string
}

var (
	PROFILE_TEMPLATE = "\n\n[profile %s]\n    credential_process = sh -c '\"%s\" \"%s\" \"%s\" \"%s\"%s'"
	AWS_FILE_PATH    = fmt.Sprintf("%s/.aws/config", os.Getenv("HOME"))
)

func (c AWSConfig) GetPath() string {
	return c.path
}

func GetProfileBody(profileName, vault, item, assumeRole, mfa, labelAccessKey, labelSecretAccessKey string) string {
	additionalOptionsArray := []string{}

	if assumeRole != "" {
		additionalOptionsArray = append(additionalOptionsArray, fmt.Sprintf("\"-a\" \"%s\"", assumeRole))
	}

	if mfa != "" {
		additionalOptionsArray = append(additionalOptionsArray, fmt.Sprintf("\"-m\" \"%s\"", mfa))
	}

	if labelAccessKey != awsvault.AWS_ACCESS_KEY_FIELD_DEFAULT {
		additionalOptionsArray = append(additionalOptionsArray, fmt.Sprintf("\"-k\" \"%s\"", labelAccessKey))
	}

	if labelSecretAccessKey != awsvault.AWS_SECRET_ACCESS_KEY_FIELD_DEFAULT {
		additionalOptionsArray = append(additionalOptionsArray, fmt.Sprintf("\"-s\" \"%s\"", labelSecretAccessKey))
	}

	additionalOptions := strings.TrimSpace(strings.Join(additionalOptionsArray[:], " "))
	if len(additionalOptions) > 0 {
		additionalOptions = " " + additionalOptions
	}

	fmt.Println(additionalOptions)

	return fmt.Sprintf(
		PROFILE_TEMPLATE,
		profileName,
		config.COMMAND_ROOT,
		config.COMMAND_CLI,
		vault,
		item,
		additionalOptions,
	)
}

func (c AWSConfig) WriteProfile(body string) error {
	_, err := os.Stat(c.path)
	if err != nil {
		if os.IsNotExist(err) {
			err = ioutil.WriteFile(c.path, []byte(body), 0644)
			return nil
		} else {
			return err
		}
	}

	file, err := os.OpenFile(c.path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := file.WriteString(body); err != nil {
		return err
	}
	return nil
}

func (config AWSConfig) Add(vault, item, mfaArn, assumeRoleArn string) error {
	fmt.Print(vault, item, mfaArn, assumeRoleArn)
	return nil
}

func NewAwsConfig(path string) *AWSConfig {
	return &AWSConfig{
		path: path,
	}
}
