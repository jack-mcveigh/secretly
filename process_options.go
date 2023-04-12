package secretly

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jack-mcveigh/secretly/internal"
	"gopkg.in/yaml.v3"
)

type (
	unmarshalFunc func([]byte, any) error

	secretConfig struct {
		Version string `json:"version" yaml:"version"`
	}
)

// WithVersionsFromConfig returns an internal.ProcessOption which overwrites
// the specified/default secret versions with versions provided in the file.
//
// Types of version files are determined by their extensions.
// Accepted version file types are:
//  1. JSON (.json)
//  2. YAML (.yaml,.yml)
func WithVersionsFromConfig(filePath string) internal.ProcessOption {
	return func(fields []internal.Field) error {
		b, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}

		switch ext := filepath.Ext(filePath); ext {
		case ".json":
			err = setVersionsFromConfig(json.Unmarshal, b, fields)
		case ".yaml", ".yml":
			err = setVersionsFromConfig(yaml.Unmarshal, b, fields)
		default:
			err = fmt.Errorf("file type \"%s\" not supported", ext)
		}
		return err
	}
}

// setVersionsFromConfig retrieves version info for the fields
// by applying unmarshal to the bytes, b.
func setVersionsFromConfig(unmarshal unmarshalFunc, b []byte, fields []internal.Field) error {
	secretConfigMap := make(map[string]secretConfig, len(fields))

	err := unmarshal(b, &secretConfigMap)
	if err != nil {
		return err
	}

	for i, f := range fields {
		if sc, ok := secretConfigMap[f.Name()]; ok {
			fields[i].SecretVersion = sc.Version
		}
	}

	return nil
}

// WithVersionsFromEnv returns an internal.ProcessOption which overwrites
// the specified/default secret versions with versions from the environment.
// Environment variables are to be named with the following logic:
//
//	if prefix
//		uppercase( prefix + "_" + field.Name() )
//	else
//		uppercase( field.Name() )
func WithVersionsFromEnv(prefix string) internal.ProcessOption {
	return func(fields []internal.Field) error {
		if prefix != "" {
			prefix += "_"
		}

		for i, field := range fields {
			name := strings.ReplaceAll(field.Name(), "-", "_")
			key := strings.ToUpper(prefix + name)
			if v, ok := os.LookupEnv(key); ok {
				fields[i].SecretVersion = v // TODO: Support types other than string
			}
		}
		return nil
	}
}
