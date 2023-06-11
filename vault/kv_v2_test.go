package vault

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"

	vault "github.com/hashicorp/vault/api"
	"github.com/jack-mcveigh/secretly"
)

type stubKVv2Client struct {
	secrets map[string][]byte

	accessed                   bool
	failIfAccessedMoreThanOnce bool
}

func newStubKVv2ClientWithSecrets() *stubKVv2Client {
	c := &stubKVv2Client{
		secrets: make(map[string][]byte),
	}

	c.secrets[fmt.Sprintf(secretKeyFormat, "fake-secret", 0)] = testSecretContent
	c.secrets[fmt.Sprintf(secretKeyFormat, "fake-secret", 1)] = testSecretContent

	return c
}

func (c *stubKVv2Client) Get(ctx context.Context, secretPath string) (*vault.KVSecret, error) {
	return c.GetVersion(ctx, secretPath, 0)
}

func (c *stubKVv2Client) GetVersion(ctx context.Context, secretPath string, version int) (*vault.KVSecret, error) {
	if c.failIfAccessedMoreThanOnce && c.accessed {
		return nil, errSecretAccessedMoreThanOnce
	}
	c.accessed = true

	key := fmt.Sprintf(secretKeyFormat, secretPath, version)

	if b, ok := c.secrets[key]; ok {
		// In vault, the most basic secret content is a map.
		var secret map[string]interface{}
		err := json.Unmarshal(b, &secret)
		if err != nil {
			return nil, err
		}

		resp := &vault.KVSecret{
			Data: secret,
		}
		return resp, nil
	}
	return nil, errSecretNotFound
}

func TestKVv2GetSecretVersion(t *testing.T) {
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
				version: "0",
			},
			want:    testSecretContent,
			wantErr: nil,
		},
		{
			name: "Success With Numbered Version",
			secretInfo: secretInfo{
				name:    "fake-secret",
				version: "1",
			},
			want:    testSecretContent,
			wantErr: nil,
		},
		{
			name: "Success With Latest Version",
			secretInfo: secretInfo{
				name:    "fake-secret",
				version: "0",
			},
			want:    testSecretContent,
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

			smc := newStubKVv2ClientWithSecrets()
			c := KVv2Client{client: smc, secretCache: secretly.NewSecretCache()}

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

func TestKVv2GetSecretVersionCaching(t *testing.T) {
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
					version: "0",
				},
				{
					name:    "fake-secret",
					version: "0",
				},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			smc := newStubKVv2ClientWithSecrets()
			smc.failIfAccessedMoreThanOnce = true

			c := KVv2Client{
				client:      smc,
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
