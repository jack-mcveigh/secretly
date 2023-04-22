// Package secretly provides a client wrapper
// for popular secret management services.
//
// The secretly package's client interface exposes convenience methods
// for retrieving secrets.
package secretly

import (
	"context"
)

type Client interface {
	// Process resolves the provided specification.
	// ProcessOptions can be provided
	// to add additional processing for the fields,
	// like reading version info from the env or a file.
	Process(spec any, opts ...ProcessOption) error

	// GetSecret retrieves the latest secret version for name
	// from the secret management service.
	GetSecret(ctx context.Context, name string) ([]byte, error)

	// GetSecretVersion retrieves the specific secret version for name
	// from the secret management service.
	GetSecretVersion(ctx context.Context, name, version string) ([]byte, error)

	// Close releases resources consumed by the client.
	Close() error
}
