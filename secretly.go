// Package secretly provides a client wrapper
// for popular secret management services.
//
// The secretly package's client interface exposes convenience methods
// for retrieving secrets.
package secretly

import "context"

// Config provides configuration
// to change the behavior
// of secretly client wrappers.
type Config struct {
	// DisableCaching disables the secret caching feature.
	// By default, secret caching is enabled.
	// With this set to true,
	// repeated gets to the same secret version will reach out
	// to the secret manager client.
	DisableCaching bool
}

// Client describes a secretly secret manager client wrapper.
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
