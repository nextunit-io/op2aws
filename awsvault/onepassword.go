package awsvault

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const (
	CLI_COMMAND                         = "op"
	OP_GET_ITEM_PATH                    = "op://%s/%s/%s"
	AWS_ACCESS_KEY_FIELD_DEFAULT        = "aws_access_key_id"
	AWS_SECRET_ACCESS_KEY_FIELD_DEFAULT = "aws_secret_access_key"
	AWS_MFA_FIELD_DEFAULT               = "TODO"
)

type OpInterface interface {
	GetName() string

	OpVault | OpItem | OpEntry
}

type OpEntry struct {
	Id      string `json:"id"`
	Section struct {
		Id string `json:"id"`
	} `json:"section"`
	Type      string `json:"type"`
	Label     string `json:"label"`
	Reference string `json:"reference"`
}

type OpVault struct {
	Id      string `json:"id"`
	Name    string `json:"name"`
	Version int    `json:"content_version,omitempty"`
}

type OpUrl struct {
	Primary bool   `json:"primary"`
	Href    string `json:"href"`
}

type OpItem struct {
	Id                    string    `json:"id"`
	Title                 string    `json:"title"`
	Version               int       `json:"version,omitempty"`
	Vault                 OpVault   `json:"vault"`
	Category              string    `json:"category"`
	LastEditedBy          string    `json:"last_edited_by"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
	AdditionalInformation string    `json:"additional_information"`
	Urls                  []OpUrl   `json:"urls"`
}

type OnePassword struct {
	commandLineClient CommandInterface

	vault string
	item  string

	accessKeyField       string
	secretAccessKeyField string
	mfaField             string

	Vault
}

func (e OpEntry) GetName() string {
	return e.Label
}

func (v OpVault) GetName() string {
	return v.Name
}

func (i OpItem) GetName() string {
	return i.Title
}

func getOutput(cmd CmdInterface) (string, error) {
	stdout, err := cmd.Output()

	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(stdout)), nil
}

func (client *OnePassword) getItem(path string) (string, error) {
	cmd := client.commandLineClient.Command(CLI_COMMAND, "read", path)
	return getOutput(cmd)
}

func GetVaults(commandLineClient CommandInterface) ([]OpVault, error) {
	cmd := commandLineClient.Command(CLI_COMMAND, "vault", "list", "--format", "json")
	output, err := getOutput(cmd)
	if err != nil {
		return nil, err
	}
	var vault []OpVault

	err = json.Unmarshal([]byte(output), &vault)
	if err != nil {
		return nil, err
	}

	return vault, nil
}

func GetEntries(commandLineClient CommandInterface, vault, item string) ([]OpEntry, error) {
	cmd := commandLineClient.Command(CLI_COMMAND, "item", "get", item, "--vault", vault, "--format", "json")
	output, err := getOutput(cmd)
	if err != nil {
		return nil, err
	}
	var items struct {
		Fields []OpEntry `json:"fields"`
	}

	err = json.Unmarshal([]byte(output), &items)
	if err != nil {
		return nil, err
	}

	return items.Fields, nil
}

func GetItems(commandLineClient CommandInterface, vault string) ([]OpItem, error) {
	cmd := commandLineClient.Command(CLI_COMMAND, "item", "list", "--vault", vault, "--format", "json")
	output, err := getOutput(cmd)
	if err != nil {
		return nil, err
	}
	var items []OpItem

	err = json.Unmarshal([]byte(output), &items)
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (client OnePassword) GetAccessKeyId() (string, error) {
	return client.getItem(fmt.Sprintf(OP_GET_ITEM_PATH, client.vault, client.item, client.accessKeyField))
}

func (client OnePassword) GetSecretAccessKey() (string, error) {
	return client.getItem(fmt.Sprintf(OP_GET_ITEM_PATH, client.vault, client.item, client.secretAccessKeyField))
}

func (client OnePassword) GetOTP() (string, error) {
	// TODO: check if there is an option to retrieve a otp inside of a specific label when it comes to multiple otp inside of one item
	cmd := client.commandLineClient.Command(CLI_COMMAND, "item", "get", client.item, "--vault", client.vault, "--otp")
	return getOutput(cmd)
}

func (client OnePassword) VaultAvailable() bool {
	cmd := client.commandLineClient.Command(CLI_COMMAND)
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

func NewOnePasswordVault(commandLineClient CommandInterface, vault, item string) *OnePassword {
	return &OnePassword{
		commandLineClient:    commandLineClient,
		vault:                vault,
		item:                 item,
		accessKeyField:       AWS_ACCESS_KEY_FIELD_DEFAULT,
		secretAccessKeyField: AWS_SECRET_ACCESS_KEY_FIELD_DEFAULT,
		mfaField:             AWS_MFA_FIELD_DEFAULT,
	}
}
