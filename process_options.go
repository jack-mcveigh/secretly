package secretly

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type (
	unmarshalFunc func([]byte, any) error

	secretConfig struct {
		Type       string `json:"type" yaml:"type"`
		Name       string `json:"name" yaml:"name"`
		Key        string `json:"key" yaml:"key"`
		Version    string `json:"version" yaml:"version"`
		SplitWords bool   `json:"split_words" yaml:"split_words"`
	}
)

// ApplyConfig returns an ProcessOption which overwrites
// the specified/default field values with the provided config.
// Can be used to overwrite any of the configurable field values.
//
// Types of config files are determined by their extensions.
// Accepted config file types are:
//  1. JSON (.json)
//  2. YAML (.yaml,.yml)
func ApplyConfig(filePath string) ProcessOption {
	return func(fields []Field) error {
		b, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}

		switch ext := filepath.Ext(filePath); ext {
		case ".json":
			err = setFieldsWithConfig(json.Unmarshal, b, fields)
		case ".yaml", ".yml":
			err = setFieldsWithConfig(yaml.Unmarshal, b, fields)
		default:
			err = fmt.Errorf("file type \"%s\" not supported", ext)
		}
		return err
	}
}

// setFieldsWithConfig overwrites fields applying unmarshal to the bytes, b.
func setFieldsWithConfig(unmarshal unmarshalFunc, b []byte, fields []Field) error {
	secretConfigMap := make(map[string]secretConfig, len(fields))

	err := unmarshal(b, &secretConfigMap)
	if err != nil {
		return err
	}

	for i, f := range fields {
		if sc, ok := secretConfigMap[f.Name()]; ok {
			if sc.Type != "" {
				fields[i].SecretType = sc.Type
			}
			if sc.Name != "" {
				fields[i].SecretName = sc.Name
			}
			if sc.Key != "" {
				fields[i].MapKeyName = sc.Key
			}
			if sc.Version != "" {
				fields[i].SecretVersion = sc.Version
			}
			if sc.SplitWords {
				fields[i].SplitWords = sc.SplitWords
			}
		}
	}

	return nil
}

// WithVersionsFromEnv returns an ProcessOption which overwrites
// the specified/default secret versions with versions from the environment.
// Environment variables are to be named with the following logic:
//
//	if prefix
//		uppercase( prefix + "_" + field.Name() ) + "_VERSION"
//	else
//		uppercase( field.Name() ) + "_VERSION"
func WithVersionsFromEnv(prefix string) ProcessOption {
	return func(fields []Field) error {
		if prefix != "" {
			prefix += "_"
		}

		for i, field := range fields {
			name := strings.ReplaceAll(field.Name(), "-", "_")
			key := strings.ToUpper(prefix + name + "_VERSION")
			if v, ok := os.LookupEnv(key); ok {
				fields[i].SecretVersion = v // TODO: Support types other than string
			}
		}
		return nil
	}
}
