package cache

type CacheClient struct {
	path        string
	vault       string
	item        string
	mfa         string
	assume_role string
}

func (cache CacheClient) Vault(vault string) {
	cache.vault = vault
}

func (cache CacheClient) Item(item string) {
	cache.item = item
}

func (cache CacheClient) MFA(mfa string) {
	cache.mfa = mfa
}

func (cache CacheClient) AssumeRole(assume_role string) {
	cache.assume_role = assume_role
}

func New(path string) *CacheClient {
	return &CacheClient{path: path}
}
