# Secretly

[![GoDoc](https://godoc.org/github.com/jack-mcveigh/secretly?status.svg)](https://godoc.org/github.com/jack-mcveigh/secretly)
[![Tests workflow](https://img.shields.io/github/actions/workflow/status/jack-mcveigh/secretly/unit-test-base.yaml?branch=main&longCache=true&label=tests&logo=github&logoColor=fff)](https://github.com/jack-mcveigh/secretly/actions?query=workflow%3ATest%20Base)
[![Go Report Card](https://goreportcard.com/badge/github.com/jack-mcveigh/secretly)](https://goreportcard.com/report/github.com/jack-mcveigh/secretly)
[![License: MIT](https://img.shields.io/badge/license-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

___Secretly___ was created to allow Go applications to easily interface with popular secret management services and reduce secret ingestion boiler-plate. In-memory secret caching is included to reduce the number of operations against the secret management service, when dealing with secrets that store map data in the form of JSON and YAML.

Below is a list of the currently supported secret management services:

* [Google Cloud Platform's Secret Manager](https://cloud.google.com/secret-manager)

If there's a secret management service missing that you'd like to see, create a Feature Request!

## Usage

See the brief overview below or check out our [examples](examples).

## Overview

### Tag Support

#### Available Tags

* __type__ - The secret content's structure.
  * _Valid Values_: "text", "json", "yaml"
  * _Default_: "text"
* __name__ - The secret's name
* __key__ - The specific field to extract from the secret's content. Note: Requires type "json" or "yaml".
  * _Default_: The struct field name (split if __split_words__ is true).
* __version__ - The version of the secret to retrieve.
  * _Default_: 0 (translates to the latest version within the client wrappers, e.g. with GCP Secret Manager, 0 -> "latest")
* __split_words__ - If the field name is used as the secret __name__ and/or __key__, split it with underscores. If set to true and a process option is provided that combines __name__ and __key__, the __name__ and __key__ will be separated with an underscore.
  * _Default_: false

Below is an example structure definition detailing default behavior, and the available tags:

```go
type Specification struct {
    // The latest version of a secret named "TextSecret" that stores text data.
    TextSecret string `type:"text"`

    // The first version of a secret named "TextSecretVersion" that stores text data.
    // Rather than retrieving the latest version, retrieve version 1.
    TextSecretWithVersion1 string `type:"text" version:"1"`

    // The latest version of a secret named "Split_Text_Secret" that stores text data.
    SplitTextSecret string `type:"text" split_words:"true"`

    // The latest version of a secret named "Json_Secret" that stores json data
    // including a key "Json_Secret_Key".
    JsonSecretKey int `type:"json" name:"Json_Secret" split_words:"true"`

    // The latest version of a secret named "Yaml_Secret" that stores yaml data
    // including a key "Yaml_Secret_Key".
    YamlSecretExplicitKey float64 `type:"json" name:"Yaml_Secret" key:"Yaml_Secret_Key"`

    // Ignored.
    IgnoredField string `ignored:"true"`

    // Also ignored.
    ignoredField string

    // The fields from the nested struct are also processed.
    SubSpecification SubSpecification
}

type SubSpecification struct {
    // The latest version of a secret named "SubTextSecret" that stores text data.
    // Since the type is not specified, the default type, text, is used.
    SubTextSecret string

    // The latest version of a secret named "SubJsonSecretAndKey" that stores yaml data
    // including a key "SubJsonSecretAndKey".
    SubJsonSecretAndKey string `type:"json"`
}
```

### Supported Field types

* __text__ - Plain text. Any secret value can be read as plain text.

    Example secrets that stores text data:

    ```text
    sensitive data
    ```

* __json__ - JSON map. The secret stores JSON data; read a specific field from the JSON map. Note: If you want to read the entire json object, use the text type.

    _Example secret that stores a JSON map:_

    ```json
    {
        "sensitive-field-1": "sensitive data"
    }
    ```

* __yaml__ - YAML map. The secret stores YAML data; read a specific field from the YAML map. Note: If you want to read the entire yaml mapping, use the text type.

    _Example secret that stores a YAML map:_

    ```yaml
    sensitive-field-1: sensitive data
    ```

### Secret Versioning

Secretly provides two options for specifying secret versions other than the __version__ tag:

1. Read secret versions (and all other field values) from a config file:
    * Supported config file types:
        * JSON (ext: .json)
        * YAML (ext: .yaml OR .yml)

    _Example of reading secret versions from a JSON config file:_

    * versions.json

        ```json
        {
            "My-DB-Credentials_username": {
                "version": "latest"
            },
            "My-DB-Credentialspassword": {
                "version": "5"
            }
        }
        ```

    * example.go

        ```go
        ...

        type Secrets struct {
            DatabaseUsername string `type:"yaml" name:"My-DB-Credentials" key:"username" split_words:"true"`

            DatabasePassword string `type:"yaml" name:"My-DB-Credentials" key:"password"`
        }

        func example(client secretly.Client) Secrets {
            var s Secrets
            
            err := client.Process(&s, secretly.ApplyConfig("versions.json"))
            if err != nil {
                log.Fatal(err)
            }
            
            return s
        }

        ...
        ```

2. Read secret versions from environment variables:

    _Example of reading secret versions from environment variables:_

    * Export environment variables:

        ```bash
        export EXAMPLE_MY_DB_CREDENTIALS_USERNAME_VERSION=latest
        export EXAMPLE_MYDBCREDENTIALSPASSWORD_VERSION=5
        ```

    * example.go

        ```go
        ...

        type Secrets struct {
            DatabaseUsername string `type:"yaml" name:"My-DB-Credentials" key:"username" split_words:"true"`

            DatabasePassword string `type:"yaml" name:"My-DB-Credentials" key:"password"`
        }

        func example(client secretly.Client) Secrets {
            var s Secrets
            
            err := client.Process(&s, secretly.WithVersionsFromEnv("EXAMPLE"))
            if err != nil {
                log.Fatal(err)
            }
            
            return s
        }

        ...
        ```
