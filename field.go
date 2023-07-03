package secretly

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	yaml "gopkg.in/yaml.v3"
)

// Default values for optional field tags.
const (
	// Defaults.
	DefaultType    = "text"
	DefaultVersion = "0"

	// Tags.
	tagIgnored    = "ignored"
	tagKey        = "key"
	tagName       = "name"
	tagSplitWords = "split_words"
	tagType       = "type"
	tagVersion    = "version"
)

var (
	regexGatherWords = regexp.MustCompile("([^A-Z]+|[A-Z]+[^A-Z]+|[A-Z]+)")
	regexAcronym     = regexp.MustCompile("([A-Z]+)([A-Z][^A-Z]+)")

	ErrInvalidJSONSecret = errors.New("secret is not valid json")
	ErrInvalidYAMLSecret = errors.New("secret is not valid yaml")
	ErrSecretMissingKey  = errors.New("secret is missing provided key")
)

// Field represents a field in a struct,
// exposing its secretly tag values
// and reference to the underlying value.
type Field struct {
	SecretType    string
	SecretName    string
	SecretVersion string
	MapKeyName    string // NOTE: Only used when secretType is "json" or "yaml"
	SplitWords    bool
	Value         reflect.Value
}

// NewField constructs a field referencing the provided reflect.Value with the tags from
// the reflect.StructField applied
func NewField(fValue reflect.Value, fStructField reflect.StructField) (Field, error) {
	var (
		newField Field
		ok       bool
		err      error
	)

	// Set the reference to the field's reflection
	newField.Value = fValue

	// Get the split_words value, setting it to false if not explicitly set
	newField.SplitWords, ok, err = parseOptionalStructTagKey[bool](fStructField, tagSplitWords)
	if err != nil {
		return Field{}, StructTagError{
			Name: fStructField.Name,
			Key:  tagSplitWords,
			Err:  err,
		}
	}

	if !ok {
		newField.SplitWords = false
	}

	// Get the type value, setting it to the default, "text", if not explicitly set.
	// Also perform validation to ensure only valid types are provided
	newField.SecretType, ok, err = parseOptionalStructTagKey[string](fStructField, tagType)
	if err != nil {
		return Field{}, StructTagError{
			Name: fStructField.Name,
			Key:  tagType,
			Err:  err,
		}
	}

	if !ok {
		newField.SecretType = DefaultType
	}

	switch newField.SecretType {
	case "text", "json", "yaml":
	default:
		return Field{}, StructTagError{
			Name: fStructField.Name,
			Key:  tagType,
			Err:  ErrInvalidSecretType,
		}
	}

	// Get the name value, setting it to the field's name if not explicitly set.
	// Split the words if the default value was used and split_words was set to true
	newField.SecretName, ok, err = parseOptionalStructTagKey[string](fStructField, tagName)
	if err != nil {
		return Field{}, StructTagError{
			Name: fStructField.Name,
			Key:  tagName,
			Err:  err,
		}
	}

	if !ok {
		newField.SecretName = fStructField.Name
		if newField.SplitWords {
			newField.SecretName = splitWords(newField.SecretName)
		}
	}

	// Get the key value, if the type is "json" or "yaml", and setting it to the field's name
	// if not explicitly set. Split the words if the default value was used and
	// split_words was set to true
	switch newField.SecretType {
	case "json", "yaml":
		newField.MapKeyName, ok, err = parseOptionalStructTagKey[string](fStructField, tagKey)
		if err != nil {
			return Field{}, StructTagError{
				Name: fStructField.Name,
				Key:  tagKey,
				Err:  err,
			}
		}
		if !ok {
			newField.MapKeyName = fStructField.Name
			if newField.SplitWords {
				newField.MapKeyName = splitWords(newField.MapKeyName)
			}
		}
	default:
		if _, ok = fStructField.Tag.Lookup(tagKey); ok {
			return Field{}, StructTagError{
				Name: fStructField.Name,
				Key:  tagKey,
				Err:  ErrSecretTypeDoesNotSupportKey,
			}
		}
	}

	// Get the version value, setting it to the default, "default", if not explicitly
	// set. Split the words if the default value was used and split_words was set to true
	newField.SecretVersion, ok, err = parseOptionalStructTagKey[string](fStructField, tagVersion)
	if err != nil {
		return Field{}, StructTagError{
			Name: fStructField.Name,
			Key:  tagVersion,
			Err:  err,
		}
	}
	if !ok {
		newField.SecretVersion = DefaultVersion
	}

	return newField, nil
}

// Name returns the resolved name of the field. If the secret type is "json" or "yaml",
// the secret name and key name are combined. If "split_words" is true, the combination
// of secret name and key name are transformed into uppercase, snake case.
func (f *Field) Name() string {
	name := f.SecretName

	switch f.SecretType {
	case "json", "yaml":
		var delimiter string
		if f.SplitWords {
			delimiter = "_"
		}

		name += delimiter + f.MapKeyName
	}

	return name
}

// Set sets the field's reflect.Value with b.
func (f *Field) Set(b []byte) error {
	switch f.SecretType {
	case "text":
		return f.setText(b)
	case "json":
		return f.setJSON(b)
	case "yaml":
		return f.setYAML(b)
	default:
		return ErrInvalidSecretType
	}
}

// setText sets the field's underlying value,
// handling the input as a "text" secret.
func (f *Field) setText(b []byte) error {
	const ErrFailedConvertFormat = "failed to convert secret \"%s's\" key, \"%s\" to %s: %w"

	byteString := string(b)

	valueType := f.Value.Type()

	switch f.Value.Kind() {
	case reflect.String:
		f.Value.SetString(byteString)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var (
			value int64
			err   error
		)

		if f.Value.Kind() == reflect.Int64 && valueType.PkgPath() == "time" && valueType.Name() == "Duration" {
			var d time.Duration
			d, err = time.ParseDuration(byteString)
			value = int64(d)
		} else {
			value, err = strconv.ParseInt(byteString, 0, valueType.Bits())
		}
		if err != nil {
			t := fmt.Sprintf("int%d", valueType.Bits())

			return fmt.Errorf(ErrFailedConvertFormat, f.SecretName, f.MapKeyName, t, err)
		}

		f.Value.SetInt(value)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		value, err := strconv.ParseUint(byteString, 0, valueType.Bits())
		if err != nil {
			t := fmt.Sprintf("uint%d", valueType.Bits())

			return fmt.Errorf(ErrFailedConvertFormat, f.SecretName, f.MapKeyName, t, err)
		}

		f.Value.SetUint(value)

	case reflect.Bool:
		value, err := strconv.ParseBool(byteString)
		if err != nil {
			return fmt.Errorf(ErrFailedConvertFormat, f.SecretName, f.MapKeyName, "bool", err)
		}

		f.Value.SetBool(value)

	case reflect.Float32, reflect.Float64:
		value, err := strconv.ParseFloat(byteString, valueType.Bits())
		if err != nil {
			t := fmt.Sprintf("float%d", valueType.Bits())

			return fmt.Errorf(ErrFailedConvertFormat, f.SecretName, f.MapKeyName, t, err)
		}

		f.Value.SetFloat(value)
	}

	return nil
}

// setJSON sets the field's underlying value,
// handling the input as a "json" secret.
func (f *Field) setJSON(b []byte) error {
	var secretMap map[string]string

	err := json.Unmarshal(b, &secretMap)
	if err != nil {
		return ErrInvalidJSONSecret
	}

	if value, ok := secretMap[f.MapKeyName]; ok {
		return f.setText([]byte(value))
	}

	return fmt.Errorf("%w: secret \"%s\" missing \"%s\"", ErrSecretMissingKey, f.SecretName, f.MapKeyName)
}

// setYAML sets the field's underlying value,
// handling the input as a "yaml" secret
func (f *Field) setYAML(b []byte) error {
	var secretMap map[string]string

	err := yaml.Unmarshal(b, &secretMap)
	if err != nil {
		return ErrInvalidYAMLSecret
	}

	if value, ok := secretMap[f.MapKeyName]; ok {
		return f.setText([]byte(value))
	}

	return fmt.Errorf("%w: secret \"%s\" missing \"%s\"", ErrSecretMissingKey, f.SecretName, f.MapKeyName)
}

// parseOptionalStructTagKey parses the provided key's value from the struct field,
// returning the value as the type T, a bool indicating if the key was present, and an
// error if the key's value was not a valid T
func parseOptionalStructTagKey[T any](structField reflect.StructField, key string) (value T, ok bool, err error) {
	var raw string

	if raw, ok = structField.Tag.Lookup(key); ok { // If key present
		switch any(value).(type) {
		case string:
			value = any(raw).(T)
		case int:
			i, err := strconv.Atoi(raw)
			if err != nil {
				break
			}

			value = any(i).(T)
		case bool:
			b, err := strconv.ParseBool(raw)
			if err != nil {
				break
			}

			value = any(b).(T)
		}

		if err != nil {
			return value, false, fmt.Errorf("invalid struct tag key value: %w", err)
		}
	}

	return value, ok, nil
}

// splitWords converts the camelCase/PascalCase string, s, to snake_case
func splitWords(s string) string {
	const minAcronymLength = 3

	words := regexGatherWords.FindAllStringSubmatch(s, -1)
	if len(words) == 0 {
		return s
	}

	var name []string
	for _, words := range words {
		if m := regexAcronym.FindStringSubmatch(words[0]); len(m) == minAcronymLength {
			name = append(name, m[1], m[2])
		} else {
			name = append(name, words[0])
		}
	}

	return strings.Join(name, "_")
}
