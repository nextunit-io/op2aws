package opaws

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"nextunit/op2aws/awsvault"
	"nextunit/op2aws/config"
	"os"
	"strings"
)

type AWSConfig struct {
	path   string
	client AwsConfigInterface
}

type AwsConfigInterface interface {
	Stat(name string) (fs.FileInfo, error)
	IsNotExist(err error) bool
	WriteFile(filename string, data []byte, perm fs.FileMode) error
	OpenFile(name string, flag int, perm fs.FileMode) (AwsConfigFileInterface, error)
}

type AwsConfigFileInterface interface {
	Close() error
	WriteString(s string) (n int, err error)
}

type AwsConfigClientDefault struct {
	AwsConfigInterface
}

var (
	PROFILE_TEMPLATE = "\n\n[profile %s]\n    credential_process = sh -c '\"%s\" \"%s\" \"%s\" \"%s\"%s'"
	AWS_FILE_PATH    = fmt.Sprintf("%s/.aws/config", os.Getenv("HOME"))
)

func (AwsConfigClientDefault) Stat(name string) (fs.FileInfo, error) {
	return os.Stat(name)
}

func (AwsConfigClientDefault) IsNotExist(err error) bool {
	return os.IsNotExist(err)
}

func (AwsConfigClientDefault) WriteFile(filename string, data []byte, perm fs.FileMode) error {
	return ioutil.WriteFile(filename, data, perm)
}

func (AwsConfigClientDefault) OpenFile(name string, flag int, perm fs.FileMode) (AwsConfigFileInterface, error) {
	return os.OpenFile(name, flag, perm)
}

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

	if labelAccessKey != awsvault.AWS_ACCESS_KEY_FIELD_DEFAULT && labelAccessKey != "" {
		additionalOptionsArray = append(additionalOptionsArray, fmt.Sprintf("\"-k\" \"%s\"", labelAccessKey))
	}

	if labelSecretAccessKey != awsvault.AWS_SECRET_ACCESS_KEY_FIELD_DEFAULT && labelSecretAccessKey != "" {
		additionalOptionsArray = append(additionalOptionsArray, fmt.Sprintf("\"-s\" \"%s\"", labelSecretAccessKey))
	}

	additionalOptions := strings.TrimSpace(strings.Join(additionalOptionsArray[:], " "))
	if len(additionalOptions) > 0 {
		additionalOptions = " " + additionalOptions
	}

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
	_, err := c.client.Stat(c.path)
	if err != nil {
		if c.client.IsNotExist(err) {
			err = c.client.WriteFile(c.path, []byte(body), 0644)
			return nil
		} else {
			return err
		}
	}

	file, err := c.client.OpenFile(c.path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := file.WriteString(body); err != nil {
		return err
	}
	return nil
}

func NewAwsConfig(client AwsConfigInterface, path string) *AWSConfig {
	return &AWSConfig{
		client: client,
		path:   path,
	}
}
