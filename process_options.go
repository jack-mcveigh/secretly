package secretly

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

type (
	// ProcessOptions are optional modifiers for secret processing.
	ProcessOption func([]Field) error

	unmarshalFunc func([]byte, any) error

	secretConfig struct {
		Type       string `json:"type" yaml:"type"`
		Name       string `json:"name" yaml:"name"`
		Key        string `json:"key" yaml:"key"`
		Version    string `json:"version" yaml:"version"`
		SplitWords bool   `json:"split_words" yaml:"split_words"`
	}
)

var ErrInvalidFileType = errors.New("invalid file type")

// ApplyPatch returns an ProcessOption which overwrites
// the specified/default field values with the provided patch.
// Can be used to overwrite any of the configurable field values.
//
// Types of patch files are determined by their extensions.
// Accepted patch file types are:
//  1. JSON (.json)
//  2. YAML (.yaml,.yml)
func ApplyPatch(filePath string) ProcessOption {
	return func(fields []Field) error {
		b, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("ApplyPatch: %w", err)
		}

		switch ext := filepath.Ext(filePath); ext {
		case ".json":
			err = setFieldsWithPatch(json.Unmarshal, b, fields)
		case ".yaml", ".yml":
			err = setFieldsWithPatch(yaml.Unmarshal, b, fields)
		default:
			err = fmt.Errorf("%w: %s", ErrInvalidFileType, ext)
		}
		return fmt.Errorf("ApplyPatch: %w", err)
	}
}

// setFieldsWithPatch overwrites fields applying unmarshal to the bytes, b.
func setFieldsWithPatch(unmarshal unmarshalFunc, b []byte, fields []Field) error {
	secretConfigMap := make(map[string]secretConfig, len(fields))

	err := unmarshal(b, &secretConfigMap)
	if err != nil {
		return err
	}

	for idx, f := range fields {
		sc, ok := secretConfigMap[f.Name()]
		if !ok {
			continue
		}

		if sc.Type != "" {
			fields[idx].SecretType = sc.Type
		}
		if sc.Name != "" {
			fields[idx].SecretName = sc.Name
		}
		if sc.Key != "" {
			fields[idx].MapKeyName = sc.Key
		}
		if sc.Version != "" {
			fields[idx].SecretVersion = sc.Version
		}
		if sc.SplitWords {
			fields[idx].SplitWords = sc.SplitWords
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
