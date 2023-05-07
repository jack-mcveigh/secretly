package aws

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/jack-mcveigh/secretly"
)

// AWS Staging Labels are used to access secret versions by an alias,
// retrieving versions such as the latest version.
const (
	AWSCURRENT  = "AWSCURRENT"  // latest
	AWSPREVIOUS = "AWSPREVIOUS" // latest - 1
	AWSPENDING  = "AWSPENDING"  // Temporary while secret is being rotated
)

// awssmc describes required AWS Secrets Manager client methods
type awssmc interface {
	GetSecretValueWithContext(ctx aws.Context, input *secretsmanager.GetSecretValueInput, opts ...request.Option) (*secretsmanager.GetSecretValueOutput, error)
}

// Client is the AWS Secrets Manager Client wrapper.
// Implements secretly.Client
type Client struct {
	client      awssmc
	secretCache secretly.SecretCache
}

// Compile time check to assert that client implements secretly.Client
var _ secretly.Client = (*Client)(nil)

// NewClient returns an AWS AWS Secrets Manager client wrapper
// with the configs applied.
// Will error if authentication with the secrets manager fails.
func NewClient(p client.ConfigProvider, cfgs ...*aws.Config) (*Client, error) {
	smc := secretsmanager.New(p, cfgs...)

	c := &Client{
		client:      smc,
		secretCache: secretly.NewSecretCache(),
	}
	return c, nil
}

// NewClient wraps the AWS Secrets Manager client.
func Wrap(client *secretsmanager.SecretsManager) *Client {
	c := &Client{
		client:      client,
		secretCache: secretly.NewSecretCache(),
	}
	return c
}

// Process resolves the provided specification
// using AWS Secrets Manager.
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
		VersionStage: aws.String(label),
	})
	c.secretCache.Add(name, label, b)
	return b, err
}

// GetSecretWithVersion retrieves the specific secret version for name
// from AWS Secrets Manager.
//
// Note: The version provided can be either a version id or
// one of the default version staging labels,
// [AWSCURRENT], [AWSPREVIOUS], or [AWSPENDING].
func (c *Client) GetSecretWithVersion(ctx context.Context, name, versionOrVersionStage string) ([]byte, error) {
	switch versionStage := versionOrVersionStage; versionStage {
	case "0", AWSCURRENT:
		return c.GetSecret(ctx, name)
	case AWSPENDING, AWSPREVIOUS:
		return c.getSecretWithStagingLabel(ctx, name, versionStage)
	}

	version := versionOrVersionStage

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
