package vault

import (
	vault "github.com/hashicorp/vault/api"
	"github.com/jack-mcveigh/secretly"
)

// Config provides both Vault KV V1 and secretly wrapper configurations.
type Config struct {
	// Token is the Vault Auth Token.
	Token string

	// MountPath is the location where the target KV secrets engine resides in Vault.
	MountPath string

	// VaultConfig is the config for the Vault client.
	// If the configuration is nil,
	// Vault will use configuration from DefaultConfig(),
	// which is the recommended starting configuration.
	VaultConfig *vault.Config

	secretly.Config
}
