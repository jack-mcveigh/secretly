package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type ProcessOption func([]field) ([]field, error)

func WithVersionsFromConfig(filePath string) ProcessOption {
	return func(fields []field) ([]field, error) {
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

func WithVersionsFromEnv(prefix string) ProcessOption {
	return func(f []field) ([]field, error) {
		return nil, errors.New("not implemented")
	}
}
