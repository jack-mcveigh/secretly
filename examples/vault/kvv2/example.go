package main

import (
	"log"

	vault "github.com/hashicorp/vault/api"
	"github.com/jack-mcveigh/secretly"
	secretlyvault "github.com/jack-mcveigh/secretly/vault"
)

const (
	vaultToken             = "a-fake-token"
	vaultMountPath         = "a-fake-mount-path"
	vaultAddress           = "www.google.com"
	secretVersionsFilePath = "versions.json"
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

	client, err := secretlyvault.NewKVv2Client(secretlyvault.Config{
		Token:       vaultToken,
		MountPath:   vaultMountPath,
		VaultConfig: cfg,
	})
	if err != nil {
		log.Fatalf("Failed to initialize vault KV v2 secret engine client: %v", err)
	}

	// Or initialize by wrapping your own Vault KV V2 Secret Engine client.
	//
	// vc, err := vault.NewClient(cfg)
	// if err != nil {
	// 	log.Fatalf("Failed to initialize vault KV V2 secret engine client: %v", err)
	// }
	// vc.SetToken(vaultToken)
	// client := secretlyvault.WrapKVv2(vc.KVv2(vaultMountPath), secretlyvault.Config{})

	var sc SecretConfig
	err = client.Process(&sc, secretly.ApplyPatch(secretVersionsFilePath))
	if err != nil {
		log.Fatalf("Failed to process SecretConfig: %v", err)
	}

	log.Printf("Username: %s", sc.DatabaseUsername)
	log.Printf("Password: %s", sc.DatabasePassword)
}
