package secretly

import (
	"context"

	"github.com/jack-mcveigh/secretly/internal"
)

type Client interface {
	// Process resolves the provided specification. "internal.ProcessOptions" can be
	// provided to add additional processing to the fields, like reading version info
	// from the env or a file
	Process(spec any, opts ...internal.ProcessOption) error

	// GetSecret retrieves the latest version of the secret from the
	// secret management service
	GetSecret(ctx context.Context, name string) ([]byte, error)

	// GetSecret retrieves the a specific version of the secret from the
	// secret management service
	GetSecretVersion(ctx context.Context, name, version string) ([]byte, error)

	// Close releases resources consumed by the client
	Close() error
}
