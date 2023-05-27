package secretly

import (
	"context"
	"reflect"
	"regexp"
)

var regexMatchCapitals = regexp.MustCompile("([a-z0-9])([A-Z])")

// Process interprets the provided specification,
// resolving the described secrets
// with the provided secret management Client.
func Process(c Client, spec any, opts ...ProcessOption) error {
	fields, err := processSpec(spec, opts...)
	if err != nil {
		return err
	}

	for _, f := range fields {
		b, err := c.GetSecretWithVersion(context.Background(), f.SecretName, f.SecretVersion)
		if err != nil {
			return err
		}

		err = f.Set(b)
		if err != nil {
			return err
		}
	}
	return nil
}

// processSpec interprets the provided specification,
// returning a slice of fields referencing the specification's fields.
// opts can be provided to add additional processing to the fields,
// like reading version info from the env or a file.
//
// spec must be a pointer to a struct,
// otherwise [ErrInvalidSpecification] is returned.
func processSpec(spec any, opts ...ProcessOption) ([]Field, error) {
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
	fields, err := processStruct(specValue, specType)
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

// processStruct recursively processes the struct, specValue,
// returning a slice of its fields.
func processStruct(specValue reflect.Value, specType reflect.Type) ([]Field, error) {
	fields := make([]Field, 0, specValue.NumField())
	for i := 0; i < specValue.NumField(); i++ {
		f, fStructField := specValue.Field(i), specType.Field(i)

		// Get the ignored value, setting it to false if not explicitly set
		ignored, _, err := parseOptionalStructTagKey[bool](fStructField, tagIgnored)
		if err != nil {
			return nil, StructTagError{
				Name: fStructField.Name,
				Key:  tagIgnored,
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
			fs, err := processStruct(f, fStructField.Type)
			if err != nil {
				return nil, err
			}
			fields = append(fields, fs...)
		case reflect.Pointer:
			for f.Kind() == reflect.Ptr {
				if f.IsNil() {
					if f.Type().Elem().Kind() != reflect.Struct {
						// value other than struct
						break
					}
					// value is a struct, initialize it
					f.Set(reflect.New(f.Type().Elem()))
				}
				f = f.Elem()
			}

			if f.Kind() == reflect.Struct {
				subFields, err := processStruct(f, f.Type())
				if err != nil {
					return nil, err
				}

				fields = append(fields, subFields...)
				continue
			}

			fallthrough
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
