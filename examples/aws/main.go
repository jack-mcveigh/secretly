package main

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/jack-mcveigh/secretly"
)

type Secrets struct {
	DatabaseUsername string `type:"text"`
	DatabasePassword string `type:"text" version:"AWSCURRENT"`
}

type SecretManagerClient struct {
	client *secretsmanager.Client
}

func (c *SecretManagerClient) GetSecret(ctx context.Context, name, version string) ([]byte, error) {
	// Use the latest secret version
	// if the default version was determined from the secrets spec.
	if version == secretly.DefaultVersion {
		version = "AWSCURRENT"
	}

	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(name),
		VersionStage: aws.String(version),
	}

	result, err := c.client.GetSecretValue(ctx, input)
	if err != nil {
		return nil, err
	}

	return []byte(*result.SecretString), nil
}

func main() {
	ctx := context.Background()

	const region = "us-east1"

	config, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		log.Fatal(err)
	}

	client := secretsmanager.NewFromConfig(config)

	smc := SecretManagerClient{client: client}

	var secrets Secrets
	if err := secretly.Process(ctx, &secrets, smc.GetSecret); err != nil {
		panic(err)
	}
}
