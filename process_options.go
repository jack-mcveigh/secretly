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

func WithVersionsFromEnv(prefix string) internal.ProcessOption {
	return func(fields []internal.Field) error {
		for i, field := range fields {
			if prefix != "" {
				prefix += "_"
			}

			key := strings.ToUpper(prefix + field.Name())
			if v, ok := os.LookupEnv(key); ok {
				fields[i].SecretVersion = v // TODO: Support types other than string
			}
		}
		return nil
	}
}
