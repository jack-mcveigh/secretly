package main

import (
	"log"

	vault "github.com/hashicorp/vault/api"
	secretlyvault "github.com/jack-mcveigh/secretly/vault"
)

const (
	vaultToken     = "a-fake-token"
	vaultMountPath = "a-fake-mount-path"
	vaultAddress   = "www.google.com"
)

type SecretConfig struct {
	// The secret stores text data and is named "Service_Integration_Token"
	// in Vault. Since "split_words" is enabled, version info can be loaded
	// from a config file by including the field name, converted to PascalCase to
	// Snake_Case, as a key: "Service_Integration_Token".
	ServiceIntegrationToken string `split_words:"true"`

	// The secret stores a json map and is named "My-Database-Credentials"
	// in Vault. The field to extract from the json secret is named
	// "Username". Version info from a config can be loaded by the config including the
	// key "My-Database-Credentials_Username". Version info from a config can be loaded
	// by exporting the variable "My_Database_Credentials_Username". Note, an underscore
	// separates the name, "My_Database_Credentials", and the key,
	// "Username", since split_words is set to true.
	DatabaseUsername string `type:"json" name:"My-Database-Credentials" key:"Username" split_words:"true"`

	// The secret stores a json map and is named "My-Database-Credentials"
	// in Vault. The field to extract from the json secret is named
	// "Password". Version info from a config can be loaded by the config including the
	// key "My-Database-Credentials_Password". Version info from a config can be loaded
	// by exporting the variable "My_Database_Credentials_Password". Note, an underscore
	// separates the name, "My_Database_Credentials", and the key,
	// "Password", since split_words is set to true.
	DatabasePassword string `type:"json" name:"My-Database-Credentials" key:"Password" split_words:"true"`
}

func main() {
	cfg := vault.DefaultConfig()
	cfg.Address = vaultAddress

	client, err := secretlyvault.NewKVv1Client(secretlyvault.Config{
		Token:       vaultToken,
		MountPath:   vaultMountPath,
		VaultConfig: cfg,
	})
	if err != nil {
		log.Fatalf("Failed to initialize vault KV v1 secret engine client: %v", err)
	}

	// Or initialize by wrapping your own Vault KV V1 Secret Engine client.
	//
	// vc, err := vault.NewClient(cfg)
	// if err != nil {
	// 	log.Fatalf("Failed to initialize vault KV V1 secret engine client: %v", err)
	// }
	// vc.SetToken(vaultToken)
	// client = secretlyvault.WrapKVv1(vc.KVv1(vaultMountPath), secretlyvault.Config{})

	var sc SecretConfig
	err = client.Process(&sc)
	if err != nil {
		log.Fatalf("Failed to process SecretConfig: %v", err)
	}

	log.Printf("Username: %s", sc.DatabaseUsername)
	log.Printf("Password: %s", sc.DatabasePassword)
}
