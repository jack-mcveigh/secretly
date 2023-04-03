package secretly

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

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
	return func(f []internal.Field) ([]internal.Field, error) {
		return nil, errors.New("not implemented")
	}
}
