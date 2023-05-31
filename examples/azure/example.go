package main

import (
	"context"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/jack-mcveigh/secretly"
	secretlyazure "github.com/jack-mcveigh/secretly/azure"
)

const (
	azureVaultURI          = "fake.vault.azure.net"
	secretVersionsFilePath = "versions.json"
)

type SecretConfig struct {
	// The secret stores text data and is named "Service_Integration_Token"
	// in Azure Key Vault. Since "split_words" is enabled, version info can be loaded
	// from a config file by including the field name, converted to PascalCase to
	// Snake_Case, as a key: "Service_Integration_Token".
	ServiceIntegrationToken string `split_words:"true"`

	// The secret stores a json map and is named "My-Database-Credentials"
	// in Azure Key Vault. The field to extract from the json secret is named
	// "Username". Version info from a config can be loaded by the config including the
	// key "My-Database-Credentials_Username". Version info from a config can be loaded
	// by exporting the variable "My_Database_Credentials_Username". Note, an underscore
	// separates the name, "My_Database_Credentials", and the key,
	// "Username", since split_words is set to true.
	DatabaseUsername string `type:"json" name:"My-Database-Credentials" key:"Username" split_words:"true"`

	// The secret stores a json map and is named "My-Database-Credentials"
	// in Azure Key Vault. The field to extract from the json secret is named
	// "Password". Version info from a config can be loaded by the config including the
	// key "My-Database-Credentials_Password". Version info from a config can be loaded
	// by exporting the variable "My_Database_Credentials_Password". Note, an underscore
	// separates the name, "My_Database_Credentials", and the key,
	// "Password", since split_words is set to true.
	DatabasePassword string `type:"json" name:"My-Database-Credentials" key:"Password" split_words:"true"`
}

func main() {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatalf("Failed to obtain a credential: %v", err)
	}

	client, err := secretlyazure.NewClient(context.Background(), secretlyazure.Config{
		VaultURI:   azureVaultURI,
		Credential: cred,
	})
	if err != nil {
		log.Fatalf("Failed to initialize azure key vault secrets client: %v", err)
	}

	// Or initialize by wrapping your own Azure keyvault secrets client.
	//
	// azsc, err := azsecrets.NewClient(azureVaultURI, cred, nil)
	// if err != nil {
	// 	log.Fatalf("Failed to initialize azure key vault secrets client: %v", err)
	// }
	// client := secretlyazure.Wrap(azsc, secretlyazure.Config{})

	var sc SecretConfig
	err = client.Process(&sc, secretly.ApplyPatch(secretVersionsFilePath))
	if err != nil {
		log.Fatalf("Failed to process SecretConfig: %v", err)
	}

	log.Printf("Username: %s", sc.DatabaseUsername)
	log.Printf("Password: %s", sc.DatabasePassword)
}
