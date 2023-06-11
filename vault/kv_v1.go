package vault

import (
	"context"
	"encoding/json"
	"errors"

	vault "github.com/hashicorp/vault/api"
	"github.com/jack-mcveigh/secretly"
)

var ErrSpecificVersionPassedToKVv1 = errors.New("KVv1 does not accept versioning")

// vaultKVv1 describes required Vault KVv1 Secrets Engine methods
type vaultKVv1 interface {
	Get(ctx context.Context, secretPath string) (*vault.KVSecret, error)
}

// Client is the Vault KVv1 Secrets Engine wrapper.
// Implements secretly.Client
//
// Note: (*KVv1Client).GetSecretVersion does not accept versioning
// other than the default version.
// (This is a limitation of the secret engine,
// use KVv2 if you want secret versioning.)
type KVv1Client struct {
	// client is the Vault KVv1 Secrets Engine.
	client vaultKVv1

	// secretCache is the cache that stores secrets => versions => content
	// to reduce secret manager accesses.
	secretCache secretly.SecretCache
}

// Compile time check to assert that client implements secretly.Client
var _ secretly.Client = (*KVv1Client)(nil)

// NewKVv1Client returns a Vault KVv1 Secrets Engine wrapper.
func NewKVv1Client(cfg Config) (*KVv1Client, error) {
	client, err := vault.NewClient(cfg.VaultConfig)
	if err != nil {
		return nil, err
	}

	if cfg.Token != "" {
		client.SetToken(cfg.Token)
	}

	var sc secretly.SecretCache
	if cfg.SecretlyConfig.DisableCaching {
		sc = secretly.NewNoOpSecretCache()
	} else {
		sc = secretly.NewSecretCache()
	}

	c := &KVv1Client{
		client:      client.KVv1(cfg.MountPath),
		secretCache: sc,
	}
	return c, nil
}

// WrapKVv2 wraps the Vault KVv1 Secrets Engine client.
func WrapKVv1(client *vault.KVv1, cfg Config) *KVv1Client {
	var sc secretly.SecretCache
	if cfg.SecretlyConfig.DisableCaching {
		sc = secretly.NewNoOpSecretCache()
	} else {
		sc = secretly.NewSecretCache()
	}

	c := &KVv1Client{
		client:      client,
		secretCache: sc,
	}
	return c
}

// Process resolves the provided specification
// using Vault KVv1 Secrets Engine.
// ProcessOptions can be provided
// to add additional processing for the fields,
// like reading version info from the env or a file.
//
// (*Client).Process is a convenience
// for calling secretly.Process with the Client.
func (c *KVv1Client) Process(spec any, opts ...secretly.ProcessOption) error {
	return secretly.Process(c, spec, opts...)
}

// GetSecret retrieves the latest secret for name
// from Vault KVv1 Secrets Engine.
func (c *KVv1Client) GetSecret(ctx context.Context, name string) ([]byte, error) {
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

// GetSecretWithVersion behaves the same as GetSecret
// but has a side effect of returning [ErrSpecificVersionPassedToKVv1]
// when a non default secret version is passed.
func (c *KVv1Client) GetSecretWithVersion(ctx context.Context, name, version string) ([]byte, error) {
	if version != secretly.DefaultVersion {
		return nil, ErrSpecificVersionPassedToKVv1
	}
	return c.GetSecret(ctx, name)
}
