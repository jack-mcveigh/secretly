package secretly

import (
	"context"
	"fmt"
	"reflect"
)

// Process interprets the provided specification,
// resolving the described secrets
// with the provided secret management Client.
func Process(client Client, spec any, opts ...ProcessOption) error {
	fields, err := processSpec(spec, opts...)
	if err != nil {
		return fmt.Errorf("Process: %w", err)
	}

	for _, field := range fields {
		b, err := client.GetSecretWithVersion(context.Background(), field.SecretName, field.SecretVersion)
		if err != nil {
			return fmt.Errorf("Process: %w", err)
		}

		err = field.Set(b)
		if err != nil {
			return fmt.Errorf("Process: %w", err)
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
		fValue, fStructField := specValue.Field(i), specType.Field(i)

		// Get the ignored value, setting it to false if not explicitly set
		ignored, _, err := parseOptionalStructTagKey[bool](fStructField, tagIgnored)
		if err != nil {
			return nil, StructTagError{
				Name: fStructField.Name,
				Key:  tagIgnored,
				Err:  err,
			}
		}

		if ignored || !fValue.CanSet() {
			continue
		}

		switch fStructField.Type.Kind() {
		case reflect.Interface | reflect.Array | reflect.Slice | reflect.Map:
			// ignore these types
		case reflect.Struct:
			fs, err := processStruct(fValue, fStructField.Type)
			if err != nil {
				return nil, err
			}
			fields = append(fields, fs...)
		case reflect.Pointer:
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

			if fValue.Kind() == reflect.Struct {
				subFields, err := processStruct(fValue, fValue.Type())
				if err != nil {
					return nil, err
				}

				fields = append(fields, subFields...)

				continue
			}

			fallthrough
		default:
			field, err := NewField(fValue, fStructField)
			if err != nil {
				return nil, err
			}

			fields = append(fields, field)
		}
	}
	return fields, nil
}
