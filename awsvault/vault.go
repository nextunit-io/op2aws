package awsvault

import "os/exec"

type Vault interface {
	GetAccessKeyId() (string, error)
	GetItem() string
	GetOTP() (string, error)
	GetSecretAccessKey() (string, error)
	GetVault() string
	SetDefaults(accessKeyField, secretAccessKeyField, mfaField string)
	VaultAvailable() bool
}

type CommandInterface interface {
	Command(name string, arg ...string) CmdInterface
}

type CmdInterface interface {
	Output() ([]byte, error)
}

type CommandClientDefault struct {
	CommandInterface
}

func (CommandClientDefault) Command(name string, arg ...string) CmdInterface {
	return exec.Command(name, arg...)
}
