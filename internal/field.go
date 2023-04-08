package internal

import (
	"errors"
	"reflect"
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

type Field struct {
	SecretType    string
	SecretName    string
	SecretVersion string
	MapKeyName    string // NOTE: Only used when secretType is "map"
	SplitWords    bool
	Value         reflect.Value
}

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
	case "text", "map":
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

	// Get the key_name value, if the type is "map", and setting it to the field's name
	// if not explicitly set. Split the words if the default value was used and
	// split_words was set to true
	switch newField.SecretType {
	case "map":
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

func (f *Field) Name() string {
	name := f.SecretName

	if f.SecretType == "map" {
		var delimiter string
		if f.SplitWords {
			delimiter = "_"
		}
		name += delimiter + f.MapKeyName
	}
	return name
}

func (f *Field) Set(b []byte) error {
	switch f.SecretType {
	case "text":
		return f.setText(b)
	case "map":
		return f.setMap(b)
	default:
		return ErrInvalidSecretType
	}
}

func (f *Field) setText(b []byte) error {
	return errors.New("not implemented")
}

func (f *Field) setMap(b []byte) error {
	return errors.New("not implemented")
}

func splitWords(s string) string {
	return RegexMatchCapitals.ReplaceAllString(s, "${1}_${2}")
}
