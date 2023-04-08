package gcp

import (
	"context"
	"fmt"
	"strconv"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/jack-mcveigh/secretly/internal"
)

type Client struct {
	client    *secretmanager.Client
	projectID string

	secretCache map[string]map[string][]byte
}

func NewClient(projectID string) (*Client, error) {
	smc, err := secretmanager.NewClient(context.TODO())
	if err != nil {
		return nil, err
	}

	c := &Client{
		client:      smc,
		projectID:   projectID,
		secretCache: make(map[string]map[string][]byte),
	}
	return c, nil
}

func (c *Client) Process(spec any, opts ...internal.ProcessOption) error {
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

func (c *Client) GetSecret(ctx context.Context, name string) ([]byte, error) {
	return c.getSecretVersion(ctx, name, "latest")
}

func (c *Client) GetSecretVersion(ctx context.Context, name, version string) ([]byte, error) {
	_, err := strconv.ParseUint(version, 10, 0)
	if err != nil {
		return nil, internal.ErrInvalidSecretVersion
	}

	if version == "0" {
		version = "latest"
	}

	if b, hit := c.getSecretFromCache(name, version); hit {
		return b, nil
	}

	b, err := c.getSecretVersion(ctx, name, version)
	if err != nil {
		return nil, err
	}

	c.addSecretToCache(name, version, b)

	return b, nil
}

func (c *Client) addSecretToCache(name, version string, b []byte) {
	if c.secretCache[name] == nil {
		c.secretCache[name] = make(map[string][]byte)
	}

	c.secretCache[name][version] = b
}

func (c *Client) getSecretFromCache(name, version string) ([]byte, bool) {
	if c.secretCache[name] == nil {
		return nil, false
	}
	if b, ok := c.secretCache[name][version]; ok {
		return b, true
	}
	return nil, false
}

func (c *Client) getSecretVersion(ctx context.Context, name, version string) ([]byte, error) {
	format := "projects/%s/secrets/%s/versions/%s"
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf(format, c.projectID, name, version),
	}

	resp, err := c.client.AccessSecretVersion(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.GetPayload().GetData(), nil
}

func (c *Client) Close() error {
	return c.client.Close()
}
