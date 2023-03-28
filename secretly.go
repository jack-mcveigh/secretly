package secretly

import "context"

type Client interface {
	Process(spec interface{}) error
	ProcessFromConfig(fileNamePath string, spec interface{}) error
	ProcessFromEnv(prefix string, spec interface{}) error
	GetSecret(ctx context.Context, name, version string) ([]byte, error)
	Close() error
}
