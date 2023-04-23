package awsvault

type Vault interface {
	GetAccessKeyId() (string, error)
	GetItem() string
	GetOTP() (string, error)
	GetSecretAccessKey() (string, error)
	GetVault() string
	SetDefaults(accessKeyField, secretAccessKeyField, mfaField string)
	VaultAvailable() bool
}
