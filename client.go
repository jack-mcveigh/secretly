package secretly

import "context"

type Client interface {
	Process(spec any) error
	GetSecret(ctx context.Context, name string) ([]byte, error)
	GetSecretVersion(ctx context.Context, name, version string) ([]byte, error)
	Close() error
}
