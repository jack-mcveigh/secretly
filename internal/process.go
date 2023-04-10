package internal

import (
	"reflect"
	"regexp"
)

type ProcessOption func([]Field) error

var RegexMatchCapitals = regexp.MustCompile("([a-z0-9])([A-Z])")

// Process interprets the provided specification, returning a slice of fields
// referencing the specification's fields. "ProcessOptions" can be provided to add
// additional processing to the fields, like reading version info from the env or a file
func Process(spec any, opts ...ProcessOption) ([]Field, error) {
	// ensure spec is a struct pointer
	specValue := reflect.ValueOf(spec)
	if specValue.Kind() != reflect.Ptr {
		return nil, ErrInvalidSpecification
	}
	specValue = specValue.Elem()
	if specValue.Kind() != reflect.Struct {
		return nil, ErrInvalidSpecification
	}

	specType := specValue.Type()
	fields, err := process(specValue, specType)
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		err := opt(fields)
		if err != nil {
			return nil, err
		}
	}

	return fields, nil
}

// process recursively processes each field.
func process(specValue reflect.Value, specType reflect.Type) ([]Field, error) {
	fields := make([]Field, 0, specValue.NumField())
	for i := 0; i < specValue.NumField(); i++ {
		f, fStructField := specValue.Field(i), specType.Field(i)

		// Get the ignored value, setting it to false if not explicitly set
		ignored, _, err := parseOptionalStructTagKey[bool](fStructField, TagIgnored)
		if err != nil {
			return nil, StructTagError{
				Name: fStructField.Name,
				Key:  TagIgnored,
				Err:  err,
			}
		}

		if ignored || !f.CanSet() {
			continue
		}

		switch fStructField.Type.Kind() {
		case reflect.Interface | reflect.Array | reflect.Slice | reflect.Map:
			// ignore these types
		case reflect.Struct:
			fs, err := process(f, fStructField.Type)
			if err != nil {
				return nil, err
			}
			fields = append(fields, fs...)
		default:
			field, err := NewField(f, fStructField)
			if err != nil {
				return nil, err
			}
			fields = append(fields, field)
		}
	}
	return fields, nil
}
