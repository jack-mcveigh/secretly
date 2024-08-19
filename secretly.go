// Package secretly provides a client wrapper
// for popular secret management services.
//
// The secretly package's client interface exposes convenience methods
// for retrieving secrets.
package secretly

import (
	"context"
	"fmt"
	"reflect"
)

// GetSecretFunc gets the secret from the secret manager.
// If your secret manager does not accept versioning,
// just ignore the version parameter.
type GetSecretFunc func(ctx context.Context, name, version string) ([]byte, error)

// Process interprets the provided specification,
// resolving the described secrets
// with the provided secret management Client.
func Process(ctx context.Context, spec any, getSecret GetSecretFunc, opts ...ProcessOption) error {
	fields, err := processSpec(spec)
	if err != nil {
		return fmt.Errorf("processing: %w", err)
	}

	for _, opt := range opts {
		err := opt(fields)
		if err != nil {
			return err
		}
	}

	for _, field := range fields {
		b, err := getSecret(ctx, field.SecretName(), field.secretVersion)
		if err != nil {
			return fmt.Errorf("getting secret: secret %q version %q: %w", field.SecretName(), field.secretVersion, err)
		}

		err = field.Set(b)
		if err != nil {
			return fmt.Errorf("setting field: %s: %w", field.Name(), err)
		}
	}

	return nil
}

// Process interprets the provided specification,
// resolving the described secrets
// with the provided secret management Client.
func MustProcess(ctx context.Context, spec any, getSecret GetSecretFunc, opts ...ProcessOption) {
	if err := Process(ctx, spec, getSecret, opts...); err != nil {
		panic(err)
	}
}

// processSpec interprets the provided specification,
// returning a slice of fields referencing the specification's fields.
// opts can be provided to add additional processing to the fields,
// like reading version info from the env or a file.
//
// spec must be a pointer to a struct,
// otherwise [ErrInvalidSpecification] is returned.
func processSpec(spec any) (fields, error) {
	// ensure spec is a struct pointer
	specValue := reflect.ValueOf(spec)
	if specValue.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("%w: not a pointer to a struct", ErrInvalidSpecification)
	}

	specValue = specValue.Elem()
	if specValue.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%w: not a pointer to a struct", ErrInvalidSpecification)
	}

	specType := specValue.Type()

	fields, err := processStruct(specValue, specType)
	if err != nil {
		return nil, fmt.Errorf("processing: %w", err)
	}

	return fields, nil
}

// processStruct recursively processes the struct, specValue,
// returning a slice of its fields.
func processStruct(specValue reflect.Value, specType reflect.Type) (fields, error) {
	fields := make(fields, 0, specValue.NumField())

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
			for fValue.Kind() == reflect.Pointer {
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
			field, err := newField(fValue, fStructField)
			if err != nil {
				return nil, err
			}

			fields = append(fields, field)
		}
	}
	return fields, nil
}
