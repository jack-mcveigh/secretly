package vault

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"testing"

	vault "github.com/hashicorp/vault/api"
	"github.com/jack-mcveigh/secretly"
)

const secretKeyFormat = "%s_%d"

var (
	errSecretNotFound             = errors.New("secret not found")
	errSecretAccessedMoreThanOnce = errors.New("secret accessed more than once")

	testSecretContent = []byte(`{"field1":"key1","field2":"key2"}`)
)

type secretInfo struct {
	name    string
	version string
}

type stubKVv1Client struct {
	secrets map[string][]byte

	accessed                   bool
	failIfAccessedMoreThanOnce bool
}

func newStubKVv1ClientWithSecrets() *stubKVv1Client {

	c := &stubKVv1Client{
		secrets: make(map[string][]byte),
	}

	c.secrets[fmt.Sprintf(secretKeyFormat, "fake-secret", 0)] = testSecretContent

	return c
}

func (c *stubKVv1Client) Get(ctx context.Context, secretPath string) (*vault.KVSecret, error) {
	return c.GetVersion(ctx, secretPath, 0)
}

func (c *stubKVv1Client) GetVersion(ctx context.Context, secretPath string, version int) (*vault.KVSecret, error) {
	if c.failIfAccessedMoreThanOnce && c.accessed {
		return nil, errSecretAccessedMoreThanOnce
	}
	c.accessed = true

	if strconv.Itoa(version) != secretly.DefaultVersion {
		return nil, ErrSpecificVersionPassedToKVv1
	}

	// At this point, version == int(secretly.DefaultVersion)
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

func TestKVv1GetSecretVersion(t *testing.T) {
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
			name: "Success With Latest Version",
			secretInfo: secretInfo{
				name:    "fake-secret",
				version: "0",
			},
			want:    testSecretContent,
			wantErr: nil,
		},
		{
			name: "Non-Default Secret Version",
			secretInfo: secretInfo{
				name:    "fake-secret",
				version: "1",
			},
			want:    nil,
			wantErr: ErrSpecificVersionPassedToKVv1,
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
			smc := newStubKVv1ClientWithSecrets()
			c := KVv1Client{client: smc, secretCache: secretly.NewSecretCache()}

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

func TestKVv1GetSecretVersionCaching(t *testing.T) {
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
			smc := newStubKVv1ClientWithSecrets()
			smc.failIfAccessedMoreThanOnce = true

			c := KVv1Client{
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
