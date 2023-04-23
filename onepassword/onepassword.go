package onepassword

import (
	"fmt"
	"os/exec"
	"strings"
)

var (
	CLI_COMMAND                         = "op"
	OP_GET_ITEM_PATH                    = "op://%s/%s/%s"
	AWS_ACCESS_KEY_FIELD_DEFAULT        = "aws_access_key_id"
	AWS_SECRET_ACCESS_KEY_FIELD_DEFAULT = "aws_secret_access_key"
	AWS_MFA_FIELD_DEFAULT               = "TODO"
)

type OnePassword struct {
	vault string
	item  string

	accessKeyField       string
	secretAccessKeyField string
	mfaField             string
}

func getOutput(cmd *exec.Cmd) (string, error) {
	stdout, err := cmd.Output()

	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(stdout)), nil
}

func (client *OnePassword) getItem(path string) (string, error) {
	cmd := exec.Command(CLI_COMMAND, "read", path)
	return getOutput(cmd)
}

func (client *OnePassword) GetAccessKeyId() (string, error) {
	return client.getItem(fmt.Sprintf(OP_GET_ITEM_PATH, client.vault, client.item, client.accessKeyField))
}

func (client *OnePassword) GetSecretAccessKey() (string, error) {
	return client.getItem(fmt.Sprintf(OP_GET_ITEM_PATH, client.vault, client.item, client.secretAccessKeyField))
}

func (client *OnePassword) GetOTP() (string, error) {
	// TODO: check if there is an option to retrieve a otp inside of a specific label when it comes to multiple otp inside of one item
	cmd := exec.Command(CLI_COMMAND, "item", "get", client.item, "--vault", client.vault, "--otp")
	return getOutput(cmd)
}

func (client *OnePassword) CLIAvailable() bool {
	cmd := exec.Command(CLI_COMMAND)
	_, err := cmd.Output()

	if err != nil {
		return false
	}

	return true
}

func (client *OnePassword) GetVault() string {
	return client.vault
}

func (client *OnePassword) GetItem() string {
	return client.item
}

func (client *OnePassword) SetDefaults(accessKeyField, secretAccessKeyField, mfaField string) {
	client.accessKeyField = accessKeyField
	client.secretAccessKeyField = secretAccessKeyField
	client.mfaField = mfaField
}

func New(vault, item string) *OnePassword {
	return &OnePassword{
		vault:                vault,
		item:                 item,
		accessKeyField:       AWS_ACCESS_KEY_FIELD_DEFAULT,
		secretAccessKeyField: AWS_SECRET_ACCESS_KEY_FIELD_DEFAULT,
		mfaField:             AWS_MFA_FIELD_DEFAULT,
	}
}
