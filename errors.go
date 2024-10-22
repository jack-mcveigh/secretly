package secretly

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidSpecification        = errors.New("invalid specification")
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

func (e StructTagError) Error() string {
	return fmt.Sprintf("field %q: key %q: %s", e.Name, e.Key, e.Err)
}

func (e StructTagError) Unwrap() error { return e.Err }
