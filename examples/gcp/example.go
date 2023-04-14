package main

import (
	"context"
	"log"

	"github.com/jack-mcveigh/secretly"
	secretlygcp "github.com/jack-mcveigh/secretly/gcp"
)

const (
	gcpProjectId              = "project-id-12345"
	gcpSecretVersionsFilePath = "versions.json"
)

type SecretConfig struct {
	// The secret stores text data and is named "Service_Integration_Token"
	// in GCP Secret Manager. Since "split_words" is enabled, version info can be loaded
	// from a config file by including the field name, converted to PascalCase to
	// Snake_Case, as a key: "Service_Integration_Token".
	ServiceIntegrationToken string `split_words:"true"`

	// The secret stores a json map and is named "My-Database-Credentials"
	// in GCP Secret Manager. The field to extract from the json secret is named
	// "Username". Version info from a config can be loaded by the config including the
	// key "My-Database-Credentials_Username". Version info from a config can be loaded
	// by exporting the variable "My_Database_Credentials_Username". Note, an underscore
	// separates the name, "My_Database_Credentials", and the key,
	// "Username", since split_words is set to true.
	DatabaseUsername string `type:"json" name:"My-Database-Credentials" key:"Username" split_words:"true"`

	// The secret stores a json map and is named "My-Database-Credentials"
	// in GCP Secret Manager. The field to extract from the json secret is named
	// "Password". Version info from a config can be loaded by the config including the
	// key "My-Database-Credentials_Password". Version info from a config can be loaded
	// by exporting the variable "My_Database_Credentials_Password". Note, an underscore
	// separates the name, "My_Database_Credentials", and the key,
	// "Password", since split_words is set to true.
	DatabasePassword string `type:"json" name:"My-Database-Credentials" key:"Password" split_words:"true"`
}

func main() {
	client, err := secretlygcp.NewClient(context.Background(), gcpProjectId)
	if err != nil {
		log.Fatalf("Failed to initialize gcp secret manager client: %v", err)
	}

	var sc SecretConfig
	err = client.Process(&sc, secretly.WithVersionsFromConfig(gcpSecretVersionsFilePath))
	if err != nil {
		log.Fatalf("Failed to process SecretConfig: %v", err)
	}

	log.Printf("Username: %s", sc.DatabaseUsername)
	log.Printf("Password: %s", sc.DatabasePassword)
}
