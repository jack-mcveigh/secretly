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
	fields := make([]Field, 0, sValue.NumField())
	for i := 0; i < sValue.NumField(); i++ {
		f, fStructField := sValue.Field(i), sType.Field(i)

		// Get the ignored value, setting it to false if not explicitly set
		ignored, ok, err := parseOptionalStructTagKey[bool](fStructField, TagIgnored)
		if err != nil {
			return nil, StructTagError{
				Name: fStructField.Name,
				Key:  TagIgnored,
				Err:  err,
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
		err := opt(fields)
		if err != nil {
			return nil, err
		}
	}

	return fields, nil
}
