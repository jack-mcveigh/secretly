package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

// Default values for optional field tags
const (
	DefaultType    = "text"
	DefaultVersion = "0"
)

// All valid tags
const (
	TagIgnored    = "ignored"
	TagKeyName    = "key_name"
	TagSecretName = "secret_name"
	TagSplitWords = "split_words"
	TagType       = "type"
	TagVersion    = "version"
)

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
	// reduce pointer to value/struct pointer. Initialize underlying struct if needed
	// TODO: maybe remove struct handling, might not be needed for this implementation
	for fValue.Kind() == reflect.Ptr {
		if fValue.IsNil() {
			if fValue.Type().Elem().Kind() != reflect.Struct {
				// value other than struct
				break
			}
			// value is a struct, initialize it
			fValue.Set(reflect.New(fValue.Type().Elem()))
		}
		fValue = fValue.Elem()
	}

	var (
		newField Field
		ok       bool
		err      error
	)

	// Set the reference to the field's reflection
	newField.Value = fValue

	// Get the split_words value, setting it to false if not explicitly set
	newField.SplitWords, ok, err = parseOptionalStructTagKey[bool](fStructField, TagSplitWords)
	if err != nil {
		return Field{}, StructTagError{
			Name: fStructField.Name,
			Key:  TagSplitWords,
			Err:  err,
		}
	}
	if !ok {
		newField.SplitWords = false
	}

	// Get the type value, setting it to the default, "text", if not explicitly set.
	// Also perform validation to ensure only valid types are provided
	newField.SecretType, ok, err = parseOptionalStructTagKey[string](fStructField, TagType)
	if err != nil {
		return Field{}, StructTagError{
			Name: fStructField.Name,
			Key:  TagType,
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
			Key:  TagType,
			Err:  ErrInvalidSecretType,
		}
	}

	// Get the secret_name value, setting it to the field's name if not explicitly set.
	// Split the words if the default value was used and split_words was set to true
	newField.SecretName, ok, err = parseOptionalStructTagKey[string](fStructField, TagSecretName)
	if err != nil {
		return Field{}, StructTagError{
			Name: fStructField.Name,
			Key:  TagSecretName,
			Err:  err,
		}
	}
	if !ok {
		newField.SecretName = fStructField.Name
		if newField.SplitWords {
			newField.SecretName = splitWords(newField.SecretName)
		}
	}

	// Get the key_name value, if the type is "json" or "yaml", and setting it to the field's name
	// if not explicitly set. Split the words if the default value was used and
	// split_words was set to true
	switch newField.SecretType {
	case "json", "yaml":
		newField.MapKeyName, ok, err = parseOptionalStructTagKey[string](fStructField, TagKeyName)
		if err != nil {
			return Field{}, StructTagError{
				Name: fStructField.Name,
				Key:  TagKeyName,
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
		if _, ok = fStructField.Tag.Lookup(TagKeyName); ok {
			return Field{}, StructTagError{
				Name: fStructField.Name,
				Key:  TagKeyName,
				Err:  ErrSecretTypeDoesNotSupportTagKey,
			}
		}
	}

	// Get the version value, setting it to the default, "default", if not explicitly
	// set. Split the words if the default value was used and split_words was set to true
	newField.SecretVersion, ok, err = parseOptionalStructTagKey[string](fStructField, TagVersion)
	if err != nil {
		return Field{}, StructTagError{
			Name: fStructField.Name,
			Key:  TagVersion,
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
// of secret name and key name are split with an underscore
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

	value := string(b)

	typ := f.Value.Type()
	switch f.Value.Kind() {
	case reflect.String:
		f.Value.SetString(value)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var (
			v   int64
			err error
		)

		if f.Value.Kind() == reflect.Int64 && typ.PkgPath() == "time" && typ.Name() == "Duration" {
			var d time.Duration
			d, err = time.ParseDuration(value)
			v = int64(d)
		} else {
			v, err = strconv.ParseInt(value, 0, typ.Bits())
		}
		if err != nil {
			t := fmt.Sprintf("int%d", typ.Bits())
			return fmt.Errorf(ErrFailedConvertFormat, f.SecretName, f.MapKeyName, t, err)
		}

		f.Value.SetInt(v)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := strconv.ParseUint(value, 0, typ.Bits())
		if err != nil {
			t := fmt.Sprintf("uint%d", typ.Bits())
			return fmt.Errorf(ErrFailedConvertFormat, f.SecretName, f.MapKeyName, t, err)
		}
		f.Value.SetUint(v)

	case reflect.Bool:
		v, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf(ErrFailedConvertFormat, f.SecretName, f.MapKeyName, "bool", err)
		}
		f.Value.SetBool(v)

	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(value, typ.Bits())
		if err != nil {
			t := fmt.Sprintf("float%d", typ.Bits())
			return fmt.Errorf(ErrFailedConvertFormat, f.SecretName, f.MapKeyName, t, err)
		}
		f.Value.SetFloat(v)
	}

	return nil
}

// setJSON sets the field's underlying value,
// handling the input as a "json" secret.
func (f *Field) setJSON(b []byte) error {
	var secretMap map[string]string

	err := json.Unmarshal(b, &secretMap)
	if err != nil {
		return errors.New("secret is not valid json")
	}

	if value, ok := secretMap[f.MapKeyName]; ok {
		return f.setText([]byte(value))
	}

	return fmt.Errorf("the json secret, \"%s\" does not contain key \"%s\"", f.SecretName, f.MapKeyName)
}

// setYAML sets the field's underlying value,
// handling the input as a "yaml" secret
func (f *Field) setYAML(b []byte) error {
	var secretMap map[string]string

	err := yaml.Unmarshal(b, &secretMap)
	if err != nil {
		return errors.New("secret is not valid yaml")
	}

	if value, ok := secretMap[f.MapKeyName]; ok {
		return f.setText([]byte(value))
	}

	return fmt.Errorf("the yaml secret, \"%s\" does not contain key \"%s\"", f.SecretName, f.MapKeyName)
}

// parseOptionalStructTagKey parses the provided key's value from the struct field,
// returning the value as the type T, a bool indicating if the key was present, and an
// error if the key's value was not a valid T
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

// splitWords converts the camelCase/PascalCase string, s, to snake_case
func splitWords(s string) string {
	return RegexMatchCapitals.ReplaceAllString(s, "${1}_${2}")
}
