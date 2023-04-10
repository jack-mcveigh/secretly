package gcp

import (
	"context"
	"fmt"
	"strconv"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/googleapis/gax-go/v2"
	"github.com/jack-mcveigh/secretly"
	"github.com/jack-mcveigh/secretly/internal"
)

const secretVersionsFormat = "projects/%s/secrets/%s/versions/%s"

type (
	Client interface {
		AccessSecretVersion(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error)
		Close() error
	}

	client struct {
		client    Client
		projectID string

		secretCache internal.SecretCache
	}
)

// Compile time check that client implements secretly.Client
var _ secretly.Client = (*client)(nil)

// NewClient constructs a GCP client with the projectID
// TODO: support options for secretmanager.NewClient
func NewClient(projectID string) (*client, error) {
	smc, err := secretmanager.NewClient(context.TODO())
	if err != nil {
		return nil, err
	}

	c := &client{
		client:      smc,
		projectID:   projectID,
		secretCache: internal.NewSecretCache(),
	}
	return c, nil
}

func (c *client) Process(spec any, opts ...internal.ProcessOption) error {
	fields, err := internal.Process(spec)
	if err != nil {
		return err
	}

	for _, f := range fields {
		b, err := c.GetSecretVersion(context.TODO(), f.SecretName, f.SecretVersion)
		if err != nil {
			return err
		}

		err = f.Set(b)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *client) GetSecret(ctx context.Context, name string) ([]byte, error) {
	b, err := c.getSecretVersion(ctx, name, "latest")
	c.secretCache.Add(name, "latest", b)
	return b, err
}

func (c *client) GetSecretVersion(ctx context.Context, name, version string) ([]byte, error) {
	switch version {
	case "0":
		version = "latest"
	case "latest":
	default:
		_, err := strconv.ParseUint(version, 10, 0)
		if err != nil {
			return nil, internal.ErrInvalidSecretVersion
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

// getSecret retrieves the a specific version of the secret from the GCP Secret Manager
func (c *client) getSecretVersion(ctx context.Context, name, version string) ([]byte, error) {
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf(secretVersionsFormat, c.projectID, name, version),
	}

	resp, err := c.client.AccessSecretVersion(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.GetPayload().GetData(), nil
}

func (c *client) Close() error {
	return c.client.Close()
}
