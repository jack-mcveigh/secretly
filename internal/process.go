package internal

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
)

type ProcessOption func([]Field) error

var RegexMatchCapitals = regexp.MustCompile("([a-z0-9])([A-Z])")

func Process(spec interface{}, opts ...ProcessOption) ([]Field, error) {
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
		err := opt(fields)
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
