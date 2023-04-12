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

Below is an example structure definition detailing default behavior, and the available tags:

```go
type Specification struct {
    // A secret named "TextSecret" that stores text data.
    TextSecret            string `type:"text"`

    // A secret named "TextSecretVersion" that stores text data. Rather than retrieving the latest version, retrieve version 1.
    TextSecretWithVersion1 string `type:"text" version:"1"`

    // A secret named "Split_Text_Secret" that stores text data.
    SplitTextSecret       string `type:"text" split_words:"true"`
    
    // A secret named "Json_Secret" that stores mapped data including a key "Json_Secret_Key".
    JsonSecretKey          int `type:"json" secret_name:"Json_Secret" split_words:"true"`

    // A secret named "Json_Secret" that stores mapped data with a key "Json_Secret_Key_2".
    JsonSecretExplicitKey  float64 `type:"json" secret_name:"Json_Secret" key_name:"Json_Secret_Key_2"`

    // Ignored.
    IgnoredField          string `ignored:"true"`

    // Also ignored.
    ignoredField          string
}
```

### Supported Field types

* __text__ - Plain text. Any secret value can be read as plain text.

    Example secret that stores text data:

    ```text
    sensitive data
    ```

* __json__ - JSON map. The secret stores JSON data; read a specific field form the JSON map.

    _Example secret that stores a JSON map:_

    ```json
    {
        "sensitive-field-1": "sensitive data"
    }
    ```

* __yaml__ - YAML map. The secret stores YAML data; read a specific field form the YAML map.

    _Example secret that stores a YAML map:_

    ```yaml
    sensitive-field-1: sensitive data
    ```

### Secret Versioning

Secretly provides two options for specifying secret versions other than the __version__ tag:

1. Read secret versions from a config file:
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
            "My-DB-Credentials_password": {
                "version": "5"
            }
        }
        ```

    * example.go

        ```go
        ...

        type Secrets struct {
            DatabaseUsername string `type:"yaml" secret_name:"My-DB-Credentials" key_name:"username" split_words:"true"`

            DatabasePassword string `type:"yaml" secret_name:"My-DB-Credentials" key_name:"password" split_words:"true"`
        }

        func example(client secretly.Client) Secrets {
            var s Secrets
            
            err := client.Process(&s, secretly.WithVersionsFromConfig("versions.json"))
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
        export EXAMPLE_MY_DB_CREDENTIALS_USERNAME=latest
        export EXAMPLE_MY_DB_CREDENTIALS_PASSWORD=5
        ```

    * example.go

        ```go
        ...

        type Secrets struct {
            DatabaseUsername string `type:"yaml" secret_name:"My-DB-Credentials" key_name:"username" split_words:"true"`

            DatabasePassword string `type:"yaml" secret_name:"My-DB-Credentials" key_name:"password" split_words:"true"`
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
