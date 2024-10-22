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
	ProcessOption func(fields) error

	unmarshalFunc func([]byte, any) error

	secretConfig struct {
		Type       secretType `json:"type" yaml:"type"`
		Name       string     `json:"name" yaml:"name"`
		Key        string     `json:"key" yaml:"key"`
		Version    string     `json:"version" yaml:"version"`
		SplitWords bool       `json:"split_words" yaml:"split_words"`
	}
)

var ErrInvalidFileType = errors.New("invalid file type")

// WithDefaultVersion overwrites the default version, [secretly.DefaultVersion],
// with the provided version.
// Use this to set the default version to aliases like "latest" or "AWSCURRENT".
func WithDefaultVersion(version string) ProcessOption {
	return func(fields fields) error {
		for i, f := range fields {
			if f.secretVersion == DefaultVersion {
				fields[i].secretVersion = version
			}
		}

		return nil
	}
}

// WithCache caches secrets in memory
// to avoid unnecessary calls to the secret manager.
// Do not use this option if you want your application
// to handle secrets changes without restarting.
func WithCache() ProcessOption {
	return func(fields fields) error {
		cache := newCache()

		for i := range fields {
			fields[i].cache = cache
		}

		return nil
	}
}

// WithPatch returns an ProcessOption which overwrites
// the specified/default field values with the provided patch.
// Can be used to overwrite any of the configurable field values.
//
// Must be written in either YAML or YAML compatible JSON
func WithPatch(patch []byte) ProcessOption {
	return func(fields fields) error {
		return setFieldsWithPatch(yaml.Unmarshal, patch, fields)
	}
}

// WithPatchFile returns an ProcessOption which overwrites
// the specified/default field values with the provided patch.
// Can be used to overwrite any of the configurable field values.
//
// Types of patch files are determined by their extensions.
// Accepted patch file types are:
//  1. JSON (.json)
//  2. YAML (.yaml,.yml)
func WithPatchFile(filePath string) ProcessOption {
	return func(fields fields) error {
		b, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("reading patch file: %w", err)
		}

		switch ext := filepath.Ext(filePath); ext {
		case ".json":
			err = setFieldsWithPatch(json.Unmarshal, b, fields)
		case ".yaml", ".yml":
			err = setFieldsWithPatch(yaml.Unmarshal, b, fields)
		default:
			err = fmt.Errorf("%w: %s", ErrInvalidFileType, ext)
		}
		if err != nil {
			return fmt.Errorf("applying patch: %w", err)
		}

		return nil
	}
}

// setFieldsWithPatch overwrites fields applying unmarshal to the bytes, b.
func setFieldsWithPatch(unmarshal unmarshalFunc, b []byte, fields fields) error {
	secretConfigMap := make(map[string]secretConfig, len(fields))

	err := unmarshal(b, &secretConfigMap)
	if err != nil {
		return err
	}

	for i, f := range fields {
		fmt.Println(f.Name())
		sc, ok := secretConfigMap[f.Name()]
		if !ok {
			continue
		}

		if sc.Type != "" {
			fields[i].secretType = sc.Type
		}
		if sc.Name != "" {
			fields[i].secretName = sc.Name
		}
		if sc.Key != "" {
			fields[i].mapKeyName = sc.Key
		}
		if sc.Version != "" {
			fields[i].secretVersion = sc.Version
		}
		if sc.SplitWords {
			fields[i].splitWords = sc.SplitWords
		}
	}

	return nil
}

// WithVersionsFromEnv returns an ProcessOption which overwrites
// the specified/default secret versions with versions from the environment.
// Environment variables are to be named with the following logic:
//
//	if prefix
//		uppercase( prefix + "_" + field.FullName() ) + "_VERSION"
//	else
//		uppercase( field.FullName() ) + "_VERSION"
func WithVersionsFromEnv(prefix string) ProcessOption {
	return func(fields fields) error {
		if prefix != "" {
			prefix += "_"
		}

		for i, field := range fields {
			name := strings.ReplaceAll(field.Name(), "-", "_")
			key := strings.ToUpper(prefix + name + "_VERSION")

			if v, ok := os.LookupEnv(key); ok {
				fields[i].secretVersion = v // TODO: Support types other than string
			}
		}
		return nil
	}
}
