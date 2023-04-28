package aws

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/jack-mcveigh/secretly"
)

const secretKeyFormat = "%s_%s"

var (
	errSecretNotFound             = errors.New("secret not found")
	errSecretAccessedMoreThanOnce = errors.New("secret accessed more than once")

	testSecretContent = []byte("secret content")
)

type secretInfo struct {
	name    string
	version string
}

type stubClient struct {
	secrets map[string][]byte

	accessed                   bool
	failIfAccessedMoreThanOnce bool
}

func newStubClientWithSecrets() *stubClient {

	c := &stubClient{
		secrets: make(map[string][]byte),
	}

	c.secrets[fmt.Sprintf(secretKeyFormat, "fake-secret", AWSCURRENT)] = testSecretContent
	c.secrets[fmt.Sprintf(secretKeyFormat, "fake-secret", "1")] = testSecretContent

	return c
}

func (c *stubClient) GetSecretValueWithContext(ctx context.Context, input *secretsmanager.GetSecretValueInput, opts ...request.Option) (*secretsmanager.GetSecretValueOutput, error) {
	if c.failIfAccessedMoreThanOnce && c.accessed {
		return nil, errSecretAccessedMoreThanOnce
	}
	c.accessed = true

	var version string
	if input.VersionStage != nil {
		version = *input.VersionStage
	} else {
		version = *input.VersionId
	}

	key := fmt.Sprintf(secretKeyFormat, *input.SecretId, version)
	fmt.Println(key)

	if b, ok := c.secrets[key]; ok {
		v := string(b)
		resp := &secretsmanager.GetSecretValueOutput{
			SecretString: &v,
		}
		return resp, nil
	}
	return nil, errSecretNotFound
}

func (c *stubClient) Close() error { return nil }

func TestProcess(t *testing.T) {

}

func TestGetSecretVersion(t *testing.T) {
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
				version: AWSCURRENT,
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
			smc := newStubClientWithSecrets()
			c := Client{client: smc, secretCache: secretly.NewSecretCache()}

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
					version: AWSCURRENT,
				},
				{
					name:    "fake-secret",
					version: AWSCURRENT,
				},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			smc := newStubClientWithSecrets()
			smc.failIfAccessedMoreThanOnce = true

			c := Client{
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
