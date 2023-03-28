package gcp

import (
	"context"
	"fmt"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

type Client struct {
	client    *secretmanager.Client
	projectID string
}

func NewClient(projectID string) (*Client, error) {
	smc, err := secretmanager.NewClient(context.TODO())
	if err != nil {
		return nil, err
	}

	c := &Client{
		client:    smc,
		projectID: projectID,
	}

	return c, nil
}

func (m *Client) GetSecret(ctx context.Context, name, version string) ([]byte, error) {
	format := "projects/%s/secrets/%s/versions/%s"
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf(format, m.projectID, name, version),
	}

	resp, err := m.client.AccessSecretVersion(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.GetPayload().GetData(), nil
}

func (m *Client) Close() error {
	return m.client.Close()
}
