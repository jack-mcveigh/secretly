package secretly

import "context"

type Client interface {
	Process(spec interface{}) error
	ProcessWithVersionsFromConfig(fileNamePath string, spec interface{}) error
	ProcessWithVersionsFromEnv(prefix string, spec interface{}) error
	GetSecret(ctx context.Context, name string) ([]byte, error)
	GetSecretVersion(ctx context.Context, name, version string) ([]byte, error)
	Close() error
}
