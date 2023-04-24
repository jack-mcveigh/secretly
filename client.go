package secretly

import (
	"context"
)

type Client interface {
	// Process resolves the provided specification.
	// ProcessOptions can be provided
	// to add additional processing for the fields,
	// like reading version info from the env or a file.
	//
	// (*Client).Process is a convenience
	// for calling secretly.Process with the Client.
	Process(spec any, opts ...ProcessOption) error

	// GetSecret retrieves the latest secret version for name
	// from the secret management service.
	GetSecret(ctx context.Context, name string) ([]byte, error)

	// GetSecretWithVersion retrieves the specific secret version for name
	// from the secret management service.
	GetSecretWithVersion(ctx context.Context, name, version string) ([]byte, error)
}
