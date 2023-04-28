package aws

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/jack-mcveigh/secretly"
)

var (
	AWSCURRENT  = "AWSCURRENT"
	AWSPREVIOUS = "AWSPREVIOUS"
	AWSPENDING  = "AWSPENDING"
)

// awssmc describes required AWS Secrets Manager client methods
type awssmc interface {
	GetSecretValueWithContext(ctx aws.Context, input *secretsmanager.GetSecretValueInput, opts ...request.Option) (*secretsmanager.GetSecretValueOutput, error)
}

type Client struct {
	client      awssmc
	secretCache secretly.SecretCache
}

func NewClient(p client.ConfigProvider, cfgs ...*aws.Config) (*Client, error) {
	smc := secretsmanager.New(p, cfgs...)

	c := &Client{
		client:      smc,
		secretCache: secretly.NewSecretCache(),
	}
	return c, nil
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

// GetSecret retrieves the secret labeled AWSCURRENT for name
// from AWS Secrets Manager.
func (c *Client) GetSecret(ctx context.Context, name string) ([]byte, error) {
	return c.getSecretWithStagingLabel(ctx, name, AWSCURRENT)
}

// getSecretWithStagingLabel retrieves the secret labeled, label, for name
// from AWS Secrets Manager.
func (c *Client) getSecretWithStagingLabel(ctx context.Context, name, label string) ([]byte, error) {
	if b, hit := c.secretCache.Get(name, label); hit {
		return b, nil
	}
	b, err := c.getSecretVersion(ctx, &secretsmanager.GetSecretValueInput{
		SecretId:     &name,
		VersionStage: &label,
	})
	c.secretCache.Add(name, label, b)
	return b, err
}

// GetSecretWithVersion retrieves the specific secret version for name
// from AWS Secrets Manager.
func (c *Client) GetSecretWithVersion(ctx context.Context, name, version string) ([]byte, error) {
	switch version {
	case "0", AWSCURRENT:
		return c.GetSecret(ctx, name)
	case AWSPENDING, AWSPREVIOUS:
		return c.getSecretWithStagingLabel(ctx, name, version)
	}

	if b, hit := c.secretCache.Get(name, version); hit {
		return b, nil
	}

	b, err := c.getSecretVersion(ctx, &secretsmanager.GetSecretValueInput{
		SecretId:  &name,
		VersionId: &version,
	})
	if err != nil {
		return nil, err
	}

	c.secretCache.Add(name, version, b)
	return b, nil
}

// getSecret retrieves the a specific version of the secret from the AWS Secrets Manager.
func (c *Client) getSecretVersion(ctx context.Context, input *secretsmanager.GetSecretValueInput) ([]byte, error) {
	output, err := c.client.GetSecretValueWithContext(ctx, input)
	if err != nil {
		return nil, err
	}

	return []byte(*output.SecretString), nil
}
