package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"

	"gopkg.in/yaml.v3"
)

type secretConfig struct {
	Version string `json:"version" yaml:"version"`
}

var RegexMatchCapitals = regexp.MustCompile("([a-z0-9])([A-Z])")

func ProcessWithVersionsFromConfig(filePath string, spec interface{}) ([]field, error) {
	fields, err := Process(spec)
	if err != nil {
		return nil, err
	}

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

func ProcessWithVersionsFromEnv(prefix string, spec interface{}) ([]field, error) {
	_, err := Process(spec)
	if err != nil {
		return nil, err
	}
	return nil, errors.New("not implemented")
}

func Process(spec interface{}, opts ...ProcessOption) ([]field, error) {
	// ensure spec is a struct pointer
	sValue := reflect.ValueOf(spec)
	if sValue.Kind() != reflect.Ptr {
		return nil, ErrInvalidSpecification
	}
	sValue = sValue.Elem()
	if sValue.Kind() != reflect.Struct {
		return nil, ErrInvalidSpecification
	}
	sType := sValue.Type()

	// spec is a struct pointer, iterate over its fields
	fields := make([]field, 0, sValue.NumField())
	for i := 0; i < sValue.NumField(); i++ {
		f, fStructField := sValue.Field(i), sType.Field(i)

		// Get the ignored value, setting it to false if not explicitly set
		ignored, ok, err := parseOptionalStructTagKey[bool](fStructField, TagIgnored)
		if err != nil {
			return nil, StructTagError{
				name: fStructField.Name,
				key:  TagIgnored,
				err:  err,
			}
		}
		if !ok {
			ignored = false
		}

		if ignored || !f.CanSet() {
			continue
		}

		field, err := NewField(f, fStructField)
		if err != nil {
			return nil, err
		}
		fields = append(fields, field)
	}

	for _, opt := range opts {
		var err error
		fields, err = opt(fields)
		if err != nil {
			return nil, err
		}
	}

	return fields, nil
}

func parseOptionalStructTagKey[T any](sf reflect.StructField, key string) (T, bool, error) {
	var (
		raw string
		v   T
		ok  bool
		err error
	)

	if raw, ok = sf.Tag.Lookup(key); ok { // If key present
		switch any(v).(type) {
		case string:
			v = any(raw).(T)
		case int:
			i, err := strconv.Atoi(raw)
			if err != nil {
				break
			}
			v = any(i).(T)
		case bool:
			b, err := strconv.ParseBool(raw)
			if err != nil {
				break
			}
			v = any(b).(T)
		}

		if err != nil {
			return v, false, fmt.Errorf("%w: %w", ErrInvalidStructTagValue, err)
		}
	}
	return v, ok, nil
}
