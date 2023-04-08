package main

import (
	"log"

	"github.com/jack-mcveigh/secretly"
	"github.com/jack-mcveigh/secretly/gcp"
)

const (
	gcpProjectId              = "project-id-12345"
	gcpSecretVersionsFilePath = "versions.json"
)

type SecretConfig struct {
	ServiceIntegrationToken string `split_words:"true"`
	DatabaseUsername        string `type:"json" secret_name:"My-Database-Credentials" key_name:"Username" split_words:"true"`
	DatabasePassword        string `type:"json" secret_name:"My-Database-Credentials" key_name:"Password" split_words:"true"`
}

func main() {
	client, err := gcp.NewClient(gcpProjectId)
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
