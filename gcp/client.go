package gcp

import (
	"context"
	"fmt"
	"strconv"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	gax "github.com/googleapis/gax-go/v2"
	"github.com/jack-mcveigh/secretly"
	"google.golang.org/api/option"
)

const secretVersionsFormat = "projects/%s/secrets/%s/versions/%s"

// gcpsmc describes required GCP Secret Manager client methods
type gcpsmc interface {
	AccessSecretVersion(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error)
	Close() error
}

// Config provides both GCP Secret Manager client and secretly wrapper configurations.
type Config struct {
	// ProjectID identifies the GCP project from which to retrieve the secrets.
	ProjectID string

	SecretlyConfig secretly.Config
}

// Client is the GCP Secret Manager Client wrapper.
// Implements secretly.Client
type Client struct {
	// client is the GCP Secret Manager client.
	client gcpsmc

	// projectId identifies the GCP project from which to retrieve the secrets.
	projectId string

	// secretCache is the cache that stores secrets => versions => content
	// to reduce secret manager accesses.
	secretCache secretly.SecretCache
}

// Compile time check to assert that client implements secretly.Client
var _ secretly.Client = (*Client)(nil)

// NewClient returns a GCP client wrapper
// with the options applied.
// Will error if authentication with the secret manager fails.
func NewClient(ctx context.Context, cfg Config, opts ...option.ClientOption) (*Client, error) {
	smc, err := secretmanager.NewClient(ctx, opts...)
	if err != nil {
		return nil, err
	}

	var sc secretly.SecretCache
	if cfg.SecretlyConfig.DisableCaching {
		sc = secretly.NewNoOpSecretCache()
	} else {
		sc = secretly.NewSecretCache()
	}

	c := &Client{
		client:      smc,
		projectId:   cfg.ProjectID,
		secretCache: sc,
	}
	return c, nil
}

// Wrap wraps the GCP client.
func Wrap(client *secretmanager.Client, cfg Config) *Client {
	var sc secretly.SecretCache
	if cfg.SecretlyConfig.DisableCaching {
		sc = secretly.NewNoOpSecretCache()
	} else {
		sc = secretly.NewSecretCache()
	}

	c := &Client{
		client:      client,
		projectId:   cfg.ProjectID,
		secretCache: sc,
	}
	return c
}

// Process resolves the provided specification
// using GCP Secret Manager.
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
// from GCP Secret Manager.
func (c *Client) GetSecret(ctx context.Context, name string) ([]byte, error) {
	if b, hit := c.secretCache.Get(name, "latest"); hit {
		return b, nil
	}
	b, err := c.getSecretVersion(ctx, name, "latest")
	c.secretCache.Add(name, "latest", b)
	return b, err
}

// GetSecretWithVersion retrieves the specific secret version for name
// from GCP Secret Manager.
func (c *Client) GetSecretWithVersion(ctx context.Context, name, version string) ([]byte, error) {
	switch version {
	case secretly.DefaultVersion, "latest":
		return c.GetSecret(ctx, name)
	default:
		_, err := strconv.ParseUint(version, 10, 0)
		if err != nil {
			return nil, secretly.ErrInvalidSecretVersion
		}
	}

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

// getSecret retrieves the a specific version of the secret from the GCP Secret Manager.
func (c *Client) getSecretVersion(ctx context.Context, name, version string) ([]byte, error) {
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf(secretVersionsFormat, c.projectId, name, version),
	}

	resp, err := c.client.AccessSecretVersion(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.GetPayload().GetData(), nil
}

// Close releases resources consumed by the client.
func (c *Client) Close() error {
	return c.client.Close()
}
