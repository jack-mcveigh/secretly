package internal

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidSpecification           = errors.New("specification must be a struct pointer")
	ErrInvalidStructTagValue          = errors.New("invalid struct tag key value")
	ErrInvalidSecretType              = errors.New("invalid secret type")
	ErrSecretTypeDoesNotSupportTagKey = errors.New("secret type does not support this tag key")
)

type StructTagError struct {
	name string
	key  string
	err  error
}

func (e StructTagError) Error() string { return fmt.Sprintf("%s: %s: %s", e.name, e.key, e.err) }

func (e StructTagError) Is(target error) bool { return e.err == target }

func (e StructTagError) Unwrap() error { return e.err }
