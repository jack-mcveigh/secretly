package main

import (
	"log"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/jack-mcveigh/secretly"
	secretlyaws "github.com/jack-mcveigh/secretly/aws"
)

const awsSecretVersionsFilePath = "versions.json"

type SecretConfig struct {
	// The secret stores text data and is named "Service_Integration_Token"
	// in AWS Secrets Manager. Since "split_words" is enabled, version info can be loaded
	// from a config file by including the field name, converted to PascalCase to
	// Snake_Case, as a key: "Service_Integration_Token".
	ServiceIntegrationToken string `split_words:"true"`

	// The secret stores a json map and is named "My-Database-Credentials"
	// in AWS Secrets Manager. The field to extract from the json secret is named
	// "Username". Version info from a config can be loaded by the config including the
	// key "My-Database-Credentials_Username". Version info from a config can be loaded
	// by exporting the variable "My_Database_Credentials_Username". Note, an underscore
	// separates the name, "My_Database_Credentials", and the key,
	// "Username", since split_words is set to true.
	DatabaseUsername string `type:"json" name:"My-Database-Credentials" key:"Username" split_words:"true"`

	// The secret stores a json map and is named "My-Database-Credentials"
	// in AWS Secrets Manager. The field to extract from the json secret is named
	// "Password". Version info from a config can be loaded by the config including the
	// key "My-Database-Credentials_Password". Version info from a config can be loaded
	// by exporting the variable "My_Database_Credentials_Password". Note, an underscore
	// separates the name, "My_Database_Credentials", and the key,
	// "Password", since split_words is set to true.
	DatabasePassword string `type:"json" name:"My-Database-Credentials" key:"Password" split_words:"true"`
}

func main() {
	s := session.Must(session.NewSession())

	client := secretlyaws.NewClient(s)

	// Or initialize by wrapping your own AWS Secrets Manager client.
	//
	// client := secretlyaws.Wrap(secretsmanager.New(s))

	var sc SecretConfig
	err := client.Process(&sc, secretly.ApplyConfig(awsSecretVersionsFilePath))
	if err != nil {
		log.Fatalf("Failed to process SecretConfig: %v", err)
	}

	log.Printf("Username: %s", sc.DatabaseUsername)
	log.Printf("Password: %s", sc.DatabasePassword)
}
