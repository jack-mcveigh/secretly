package azure

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/keyvault/azsecrets"
	"github.com/jack-mcveigh/secretly"
)

// azuresc describes required Azure Key Vault Secrets client methods
type azuresc interface {
	GetSecret(ctx context.Context, name string, version string, options *azsecrets.GetSecretOptions) (azsecrets.GetSecretResponse, error)
}

// Config provides both Azure Key Vault Secrets client and secretly wrapper configurations.
type Config struct {
	VaultURI string

	Credential azcore.TokenCredential

	Options *azsecrets.ClientOptions

	secretly.Config
}

// Client is the Azure Key Vault Secrets client wrapper.
// Implements secretly.Client
type Client struct {
	// client is the Azure Key Vault Secrets client.
	client azuresc

	// secretCache is the cache that stores secrets => versions => content
	// to reduce secret client accesses.
	secretCache secretly.SecretCache
}

// Compile time check to assert that client implements secretly.Client
var _ secretly.Client = (*Client)(nil)

// NewClient returns a Azure Key Vault Secrets client wrapper
// with the options applied.
// Will error if authentication with the secret manager fails.
func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	azsc, err := azsecrets.NewClient(cfg.VaultURI, cfg.Credential, cfg.Options)
	if err != nil {
		return nil, err
	}

	var sc secretly.SecretCache
	if cfg.DisableCaching {
		sc = secretly.NewNoOpSecretCache()
	} else {
		sc = secretly.NewSecretCache()
	}

	c := &Client{
		client:      azsc,
		secretCache: sc,
	}
	return c, nil
}

// Wrap wraps the Azure Key Vault Secrets client.
func Wrap(client *azsecrets.Client, cfg Config) *Client {
	var sc secretly.SecretCache
	if cfg.DisableCaching {
		sc = secretly.NewNoOpSecretCache()
	} else {
		sc = secretly.NewSecretCache()
	}

	c := &Client{
		client:      client,
		secretCache: sc,
	}
	return c
}

// Process resolves the provided specification
// using Azure Key Vault Secrets.
// ProcessOptions can be provided
// to add additional processing for the fields,
// like reading version info from the env or a file.
//
// (*Client).Process is a convenience
// for calling secretly.Process with the Client.
func (c *Client) Process(spec any, opts ...secretly.ProcessOption) error {
	return secretly.Process(c, spec, opts...)
}

// GetSecret retrieves the latest secret version for name
// from Azure Key Vault Secrets.
func (c *Client) GetSecret(ctx context.Context, name string) ([]byte, error) {
	if b, hit := c.secretCache.Get(name, ""); hit {
		return b, nil
	}
	b, err := c.getSecretVersion(ctx, name, "")
	c.secretCache.Add(name, "", b)
	return b, err
}

// GetSecretWithVersion retrieves the specific secret version for name
// from Azure Key Vault Secrets.
func (c *Client) GetSecretWithVersion(ctx context.Context, name, version string) ([]byte, error) {
	fmt.Println("here1")
	switch version {
	case secretly.DefaultVersion, "":
		fmt.Println("here2")
		return c.GetSecret(ctx, name)
	}

	fmt.Println("here3")
	if b, hit := c.secretCache.Get(name, version); hit {
		return b, nil
	}

	b, err := c.getSecretVersion(ctx, name, version)
	if err != nil {
		return nil, err
	}

	c.secretCache.Add(name, version, b)
	return b, nil
}

// getSecret retrieves the a specific version of the secret from the Azure Key Vault Secrets.
func (c *Client) getSecretVersion(ctx context.Context, name, version string) ([]byte, error) {
	resp, err := c.client.GetSecret(ctx, name, version, nil)
	if err != nil {
		return nil, err
	}
	return []byte(*resp.Value), nil
}
