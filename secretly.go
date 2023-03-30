package secretly

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
)

const (
	DefaultType    = "text"
	DefaultVersion = "0"
	TagIgnored     = "ignored"
	TagKeyName     = "key_name"
	TagSecretName  = "secret_name"
	TagSplitWords  = "split_words"
	TagType        = "type"
	TagVersion     = "version"
)

var (
	ErrInvalidSpecification           = errors.New("specification must be a struct pointer")
	ErrInvalidStructTagValue          = errors.New("invalid struct tag key value")
	ErrInvalidSecretType              = errors.New("invalid secret type")
	ErrSecretTypeDoesNotSupportTagKey = errors.New("secret type does not support this tag key")

	RegexMatchCapitals = regexp.MustCompile("([a-z0-9])([A-Z])")
)

type StructTagError struct {
	name string
	key  string
	err  error
}

func (e StructTagError) Error() string { return fmt.Sprintf("%s: %s: %s", e.name, e.key, e.err) }

func (e StructTagError) Is(target error) bool { return e.err == target }

func (e StructTagError) Unwrap() error { return e.err }

type Client interface {
	Process(spec interface{}) error
	ProcessFromConfig(fileNamePath string, spec interface{}) error
	ProcessFromEnv(prefix string, spec interface{}) error
	GetSecret(ctx context.Context, name, version string) ([]byte, error)
	Close() error
}

type field struct {
	secretType    string
	secretName    string
	secretVersion string
	mapKeyName    string // NOTE: Only used when secretType is "map"
	value         reflect.Value
}

func newField(fValue reflect.Value, fStructField reflect.StructField) (field, error) {
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

	var newField field

	// Get the split_words value, setting it to false if not explicitly set
	splitWordsEnabled, ok, err := parseOptionalStructTagKey[bool](fStructField, TagSplitWords)
	if err != nil {
		return field{}, StructTagError{
			name: fStructField.Name,
			key:  TagSplitWords,
			err:  err,
		}
	}
	if !ok {
		splitWordsEnabled = false
	}

	// Get the type value, setting it to the default, "text", if not explicitly set.
	// Also perform validation to ensure only valid types are provided
	newField.secretType, ok, err = parseOptionalStructTagKey[string](fStructField, TagType)
	if err != nil {
		return field{}, StructTagError{
			name: fStructField.Name,
			key:  TagType,
			err:  err,
		}
	}
	if !ok {
		newField.secretType = DefaultType
	}
	switch newField.secretType {
	case "text", "map":
	default:
		return field{}, StructTagError{
			name: fStructField.Name,
			key:  TagType,
			err:  ErrInvalidSecretType,
		}
	}

	// Get the secret_name value, setting it to the field's name if not explicitly set.
	// Split the words if the default value was used and split_words was set to true
	newField.secretName, ok, err = parseOptionalStructTagKey[string](fStructField, TagSecretName)
	if err != nil {
		return field{}, StructTagError{
			name: fStructField.Name,
			key:  TagSecretName,
			err:  err,
		}
	}
	if !ok {
		newField.secretName = fStructField.Name
		if splitWordsEnabled {
			newField.secretName = splitWords(newField.secretName)
		}
	}

	// Get the key_name value, if the type is "map", and setting it to the field's name
	// if not explicitly set. Split the words if the default value was used and
	// split_words was set to true
	switch newField.secretType {
	case "map":
		newField.mapKeyName, ok, err = parseOptionalStructTagKey[string](fStructField, TagKeyName)
		if err != nil {
			return field{}, StructTagError{
				name: fStructField.Name,
				key:  TagKeyName,
				err:  err,
			}
		}
		if !ok {
			newField.mapKeyName = fStructField.Name
			if splitWordsEnabled {
				newField.mapKeyName = splitWords(newField.mapKeyName)
			}
		}
	default:
		if _, ok = fStructField.Tag.Lookup(TagKeyName); ok {
			return field{}, StructTagError{
				name: fStructField.Name,
				key:  TagKeyName,
				err:  ErrSecretTypeDoesNotSupportTagKey,
			}
		}
	}

	// Get the version value, setting it to the default, "default", if not explicitly
	// set. Split the words if the default value was used and split_words was set to true
	newField.secretVersion, ok, err = parseOptionalStructTagKey[string](fStructField, TagVersion)
	if err != nil {
		return field{}, StructTagError{
			name: fStructField.Name,
			key:  TagVersion,
			err:  err,
		}
	}
	if !ok {
		newField.secretVersion = DefaultVersion
	}

	return newField, nil
}

func parseSpecification(spec interface{}) ([]field, error) {
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

		field, err := newField(f, fStructField)
		if err != nil {
			return nil, err
		}
		fields = append(fields, field)
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

func splitWords(s string) string {
	return RegexMatchCapitals.ReplaceAllString(s, "${1}_${2}")
}
