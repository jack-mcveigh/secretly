package gcp

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/googleapis/gax-go/v2"
	"github.com/jack-mcveigh/secretly"
)

const testProjectId = "test-project"

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

	c.secrets[fmt.Sprintf(secretVersionsFormat, testProjectId, "fake-secret", "latest")] = testSecretContent
	c.secrets[fmt.Sprintf(secretVersionsFormat, testProjectId, "fake-secret", "1")] = testSecretContent

	return c
}

func (c *stubClient) AccessSecretVersion(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error) {
	if c.failIfAccessedMoreThanOnce && c.accessed {
		return nil, errSecretAccessedMoreThanOnce
	}
	c.accessed = true

	if b, ok := c.secrets[req.Name]; ok {
		resp := &secretmanagerpb.AccessSecretVersionResponse{
			Payload: &secretmanagerpb.SecretPayload{
				Data: b,
			},
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
				version: "latest",
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
			c := client{client: smc, projectID: testProjectId, secretCache: secretly.NewSecretCache()}

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
					version: "latest",
				},
				{
					name:    "fake-secret",
					version: "latest",
				},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			smc := newStubClientWithSecrets()
			smc.failIfAccessedMoreThanOnce = true

			c := client{
				client:      smc,
				projectID:   testProjectId,
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
