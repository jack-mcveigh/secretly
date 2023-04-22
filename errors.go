package secretly

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidSpecification        = errors.New("specification must be a struct pointer")
	ErrInvalidStructTagValue       = errors.New("invalid struct tag key value")
	ErrInvalidSecretType           = errors.New("invalid secret type")
	ErrInvalidSecretVersion        = errors.New("invalid secret version")
	ErrSecretTypeDoesNotSupportKey = errors.New("secret type does not support \"key\"")
)

// StructTagError describes an error resulting from an issue with a struct tag.
type StructTagError struct {
	Name string
	Key  string
	Err  error
}

func (e StructTagError) Error() string { return fmt.Sprintf("%s: %s: %s", e.Name, e.Key, e.Err) }

func (e StructTagError) Unwrap() error { return e.Err }
