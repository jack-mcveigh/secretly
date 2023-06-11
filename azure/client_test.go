package azure

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/keyvault/azsecrets"
	"github.com/jack-mcveigh/secretly"
)

const secretKeyFormat = "%s_%s"

var (
	errSecretNotFound             = errors.New("secret not found")
	errSecretAccessedMoreThanOnce = errors.New("secret accessed more than once")

	testSecretContent = "secret content"
)

type secretInfo struct {
	name    string
	version string
}

type stubClient struct {
	secrets map[string]string

	accessed                   bool
	failIfAccessedMoreThanOnce bool
}

func newStubClientWithSecrets() *stubClient {
	c := &stubClient{
		secrets: make(map[string]string),
	}

	c.secrets[fmt.Sprintf(secretKeyFormat, "fake-secret", "")] = testSecretContent
	c.secrets[fmt.Sprintf(secretKeyFormat, "fake-secret", "1")] = testSecretContent

	return c
}

func (c *stubClient) GetSecret(ctx context.Context, name string, version string, options *azsecrets.GetSecretOptions) (azsecrets.GetSecretResponse, error) {
	resp := azsecrets.GetSecretResponse{}
	if c.failIfAccessedMoreThanOnce && c.accessed {
		return resp, errSecretAccessedMoreThanOnce
	}
	c.accessed = true

	key := fmt.Sprintf(secretKeyFormat, name, version)

	if s, ok := c.secrets[key]; ok {
		resp.Value = &s
		return resp, nil
	}
	return resp, errSecretNotFound
}

func (c *stubClient) Close() error { return nil }

func TestGetSecretVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		secretInfo secretInfo
		want       []byte
		wantErr    error
	}{
		{
			name: "Success With Default Version",
			secretInfo: secretInfo{
				name:    "fake-secret",
				version: "",
			},
			want:    []byte(testSecretContent),
			wantErr: nil,
		},
		{
			name: "Success With Numbered Version",
			secretInfo: secretInfo{
				name:    "fake-secret",
				version: "1",
			},
			want:    []byte(testSecretContent),
			wantErr: nil,
		},
		{
			name: "Success With Latest Version",
			secretInfo: secretInfo{
				name:    "fake-secret",
				version: "",
			},
			want:    []byte(testSecretContent),
			wantErr: nil,
		},
		{
			name: "Secret Does Not Exist",
			secretInfo: secretInfo{
				name:    "fake-secret-that-does-not-exist",
				version: "0",
			},
			want:    nil,
			wantErr: errSecretNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt := tt
			t.Parallel()

			azsc := newStubClientWithSecrets()
			c := Client{client: azsc, secretCache: secretly.NewSecretCache()}

			got, err := c.GetSecretWithVersion(context.Background(), tt.secretInfo.name, tt.secretInfo.version)

			if err != tt.wantErr {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Incorrect error. Want %v, got %v", tt.wantErr, err)
				}
			}

			if !reflect.DeepEqual(tt.want, got) {
				t.Errorf("Incorrect secret content. Want %v, got %v", tt.want, got)
			}
		})
	}
}

func TestGetSecretVersionCaching(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		secretInfos []secretInfo
		wantErr     error
	}{
		{
			name: "Cache Hit",
			secretInfos: []secretInfo{
				{
					name:    "fake-secret",
					version: "",
				},
				{
					name:    "fake-secret",
					version: "",
				},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt := tt
			t.Parallel()

			azsc := newStubClientWithSecrets()
			azsc.failIfAccessedMoreThanOnce = true

			c := Client{
				client:      azsc,
				secretCache: secretly.NewSecretCache(),
			}

			var err error
			for _, secretInfo := range tt.secretInfos {
				_, err = c.GetSecretWithVersion(context.Background(), secretInfo.name, secretInfo.version)
			}

			if err != tt.wantErr {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Incorrect error. Want %v, got %v", tt.wantErr, err)
				}
			}
		})
	}
}
