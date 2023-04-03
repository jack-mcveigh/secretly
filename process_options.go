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

type secretConfig struct {
	Version string `json:"version" yaml:"version"`
}

func WithVersionsFromConfig(filePath string) internal.ProcessOption {
	return func(fields []internal.Field) ([]internal.Field, error) {
		b, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}

		secretConfigMap := make(map[string]secretConfig, len(fields))

		switch ext := filepath.Ext(filePath); ext {
		case ".json":
			err = json.Unmarshal(b, &secretConfigMap)
		case ".yaml", ".yml":
			err = yaml.Unmarshal(b, &secretConfigMap)
		default:
			return nil, fmt.Errorf("file type \"%s\" not supported", ext)
		}
		if err != nil {
			return nil, err
		}

		for i, f := range fields {
			fmt.Println(f.Name())
			if sc, ok := secretConfigMap[f.Name()]; ok {
				fields[i].SecretVersion = sc.Version
			}
		}

		return fields, nil
	}
}

func WithVersionsFromEnv(prefix string) internal.ProcessOption {
	return func(fields []internal.Field) ([]internal.Field, error) {
		for i, field := range fields {
			if prefix != "" {
				prefix += "_"
			}

			key := strings.ToUpper(prefix + field.Name())
			if v, ok := os.LookupEnv(key); ok {
				fields[i].SecretVersion = v // TODO: Support types other than string
			}
		}
		return fields, nil
	}
}
