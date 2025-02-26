package main

import (
	"context"
	"fmt"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/jack-mcveigh/secretly"
)

type Secrets struct {
	DatabaseUsername string `type:"text"`
	DatabasePassword string `type:"text" version:"latest"`
}

type SecretManagerClient struct {
	client *secretmanager.Client

	projectID string
}

func (c *SecretManagerClient) GetSecret(ctx context.Context, name, version string) ([]byte, error) {
	// Use the latest secret version
	// if the default version was determined from the secrets spec.
	if version == secretly.DefaultVersion {
		version = "latest"
	}

	accessRequest := &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("projects/%s/secrets/%s/versions/%s", c.projectID, name, version),
	}

	result, err := c.client.AccessSecretVersion(ctx, accessRequest)
	if err != nil {
		return nil, err
	}

	return result.Payload.Data, nil
}

func main() {
	ctx := context.Background()

	const projectID = "example-project-id"

	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		panic(err)
	}

	smc := SecretManagerClient{client: client, projectID: projectID}

	var secrets Secrets
	if err := secretly.Process(ctx, &secrets, smc.GetSecret); err != nil {
		panic(err)
	}
}
