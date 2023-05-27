package vault

import (
	"context"
	"encoding/json"
	"strconv"

	vault "github.com/hashicorp/vault/api"
	"github.com/jack-mcveigh/secretly"
)

// vaultKVv2 describes required Vault KVv2 Secrets Engine methods
type vaultKVv2 interface {
	Get(ctx context.Context, secretPath string) (*vault.KVSecret, error)
	GetVersion(ctx context.Context, secretPath string, version int) (*vault.KVSecret, error)
}

// KVv2Client is the Vault KVv2 Secrets Engine wrapper.
// Implements secretly.KVv2Client
type KVv2Client struct {
	// client is the Vault KVv2 Secrets Engine.
	client vaultKVv2

	// secretCache is the cache that stores secrets => versions => content
	// to reduce secret manager accesses.
	secretCache secretly.SecretCache
}

// Compile time check to assert that client implements secretly.Client
var _ secretly.Client = (*KVv2Client)(nil)

// NewKVv2Client returns a Vault KVv2 Secrets Engine wrapper.
func NewKVv2Client(cfg Config) (*KVv2Client, error) {
	client, err := vault.NewClient(cfg.VaultConfig)
	if err != nil {
		return nil, err
	}

	if cfg.Token != "" {
		client.SetToken(cfg.Token)
	}

	var sc secretly.SecretCache
	if cfg.DisableCaching {
		sc = secretly.NewNoOpSecretCache()
	} else {
		sc = secretly.NewSecretCache()
	}

	c := &KVv2Client{
		client:      client.KVv2(cfg.MountPath),
		secretCache: sc,
	}
	return c, nil
}

// WrapKVv2 wraps the Vault KVv2 Secrets Engine client.
func WrapKVv2(client *vault.KVv2, cfg Config) *KVv2Client {
	var sc secretly.SecretCache
	if cfg.DisableCaching {
		sc = secretly.NewNoOpSecretCache()
	} else {
		sc = secretly.NewSecretCache()
	}

	c := &KVv2Client{
		client:      client,
		secretCache: sc,
	}
	return c
}

// Process resolves the provided specification
// using Vault KVv2 Secrets Engine.
// ProcessOptions can be provided
// to add additional processing for the fields,
// like reading version info from the env or a file.
//
// (*Client).Process is a convenience
// for calling secretly.Process with the Client.
func (c *KVv2Client) Process(spec any, opts ...secretly.ProcessOption) error {
	return secretly.Process(c, spec, opts...)
}

// GetSecret retrieves the latest secret for name
// from Vault KVv2 Secrets Engine.
func (c *KVv2Client) GetSecret(ctx context.Context, name string) ([]byte, error) {
	if b, hit := c.secretCache.Get(name, secretly.DefaultVersion); hit {
		return b, nil
	}

	secret, err := c.client.Get(ctx, name)
	if err != nil {
		return nil, err
	}

	b, err := json.Marshal(secret.Data)
	if err != nil {
		return nil, err
	}

	c.secretCache.Add(name, secretly.DefaultVersion, b)
	return b, nil
}

// GetSecretWithVersion retrieves the specific secret version for name
// from Vault KVv2 Secrets Engine.
func (c *KVv2Client) GetSecretWithVersion(ctx context.Context, name, version string) ([]byte, error) {
	if version == secretly.DefaultVersion {
		return c.GetSecret(ctx, name)
	}

	if b, hit := c.secretCache.Get(name, version); hit {
		return b, nil
	}

	v, err := strconv.Atoi(version)
	if err != nil {
		return nil, err
	}

	secret, err := c.client.GetVersion(ctx, name, v)
	if err != nil {
		return nil, err
	}

	b, err := json.Marshal(secret.Data)
	if err != nil {
		return nil, err
	}

	c.secretCache.Add(name, version, b)
	return b, nil
}
