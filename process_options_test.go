package secretly

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/jack-mcveigh/secretly/internal"
	"gopkg.in/yaml.v3"
)

func TestSetVersionsFromConfig(t *testing.T) {
	tests := []struct {
		name          string
		unmarshalFunc unmarshalFunc
		content       []byte
		want          []internal.Field
		wantErr       bool
	}{
		{
			name:          "Valid JSON",
			unmarshalFunc: json.Unmarshal,
			content:       validJsonBytes,
			want: []internal.Field{
				{
					SecretType:    internal.DefaultType,
					SecretName:    "Text_Secret",
					SecretVersion: "1",
					MapKeyName:    "",
					SplitWords:    true,
				},
				{
					SecretType:    "map",
					SecretName:    "Map_Secret",
					SecretVersion: "latest",
					MapKeyName:    "Key",
					SplitWords:    true,
				},
			},
			wantErr: false,
		},
		{
			name:          "Invalid JSON",
			unmarshalFunc: json.Unmarshal,
			content:       invalidJsonBytes,
			want:          newDefaultSpecificationFields(),
			wantErr:       true,
		},
		{
			name:          "Valid YAML",
			unmarshalFunc: yaml.Unmarshal,
			content:       validYamlBytes,
			want: []internal.Field{
				{
					SecretType:    internal.DefaultType,
					SecretName:    "Text_Secret",
					SecretVersion: "1",
					MapKeyName:    "",
					SplitWords:    true,
				},
				{
					SecretType:    "map",
					SecretName:    "Map_Secret",
					SecretVersion: "latest",
					MapKeyName:    "Key",
					SplitWords:    true,
				},
			},
			wantErr: false,
		},
		{
			name:          "Invalid YAML",
			unmarshalFunc: yaml.Unmarshal,
			content:       invalidYamlBytes,
			want:          newDefaultSpecificationFields(),
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields := newDefaultSpecificationFields()
			err := setVersionsFromConfig(tt.unmarshalFunc, tt.content, fields)

			// If the test is set up with an invalid input, we don't care what the error is,
			// only that there is an error. The same is true for the opposite scenario
			if tt.wantErr {
				if err == nil {
					t.Errorf("Incorrect error. Want an error but did not get one")
				}
			} else {
				if err != nil {
					t.Errorf("Incorrect error. Do not want an error but got %v", err)
				}
			}

			// Don't check reflect.Value for equality
			for i := range fields {
				fields[i].Value = reflect.Value{}
			}

			if !reflect.DeepEqual(tt.want, fields) {
				t.Errorf("Incorrect fields. Want %v, got %v", tt.want, fields)
			}
		})
	}
}

func TestWithVersionsFromEnv(t *testing.T) {
	tests := []struct {
		name      string
		prefix    string
		envVarMap map[string]string
		want      []internal.Field
		wantErr   error
	}{
		{
			name:   "All Env Vars Exist",
			prefix: "TEST",
			envVarMap: map[string]string{
				"TEST_TEXT_SECRET":    "1",
				"TEST_MAP_SECRET_KEY": "latest",
			},
			want: []internal.Field{
				{
					SecretType:    internal.DefaultType,
					SecretName:    "Text_Secret",
					SecretVersion: "1",
					MapKeyName:    "",
					SplitWords:    true,
				},
				{
					SecretType:    "map",
					SecretName:    "Map_Secret",
					SecretVersion: "latest",
					MapKeyName:    "Key",
					SplitWords:    true,
				},
			},
			wantErr: nil,
		},
		{
			name:   "One Env Var Missing",
			prefix: "TEST",
			envVarMap: map[string]string{
				"TEST_MAP_SECRET_KEY": "latest",
			},
			want: []internal.Field{
				{
					SecretType:    internal.DefaultType,
					SecretName:    "Text_Secret",
					SecretVersion: "0",
					MapKeyName:    "",
					SplitWords:    true,
				},
				{
					SecretType:    "map",
					SecretName:    "Map_Secret",
					SecretVersion: "latest",
					MapKeyName:    "Key",
					SplitWords:    true,
				},
			},
			wantErr: nil,
		},
		{
			name:      "All Env Vars Missing",
			prefix:    "TEST",
			envVarMap: map[string]string{},
			want: []internal.Field{
				{
					SecretType:    internal.DefaultType,
					SecretName:    "Text_Secret",
					SecretVersion: "0",
					MapKeyName:    "",
					SplitWords:    true,
				},
				{
					SecretType:    "map",
					SecretName:    "Map_Secret",
					SecretVersion: "0",
					MapKeyName:    "Key",
					SplitWords:    true,
				},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		// setup env vars.
		for k, v := range tt.envVarMap {
			err := os.Setenv(k, v)
			if err != nil {
				t.Fatalf("Failed to set env. var. \"%s\" =\"%s\": %v", k, v, err)
			}
		}

		t.Run(tt.name, func(t *testing.T) {
			spec := TestingSpecification{}
			fields, err := internal.Process(&spec, WithVersionsFromEnv(tt.prefix))

			if err != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Incorrect error. Want %v, got %v", tt.wantErr, err)
				}
			}

			// Don't check reflect.Value for equality
			for i := range fields {
				fields[i].Value = reflect.Value{}
			}

			if !reflect.DeepEqual(tt.want, fields) {
				t.Errorf("Incorrect fields. Want %v, got %v", tt.want, fields)
			}
		})

		// teardown env. vars.
		for k := range tt.envVarMap {
			err := os.Unsetenv(k)
			if err != nil {
				t.Fatalf("Failed to unset env. var. \"%s\": %v", k, err)
			}
		}
	}
}

type TestingSpecification struct {
	TextSecret string `split_words:"true"`
	MapSecret  string `type:"map" key_name:"Key" split_words:"true"`
}

var (
	validJsonBytes = []byte(`
{
	"Text_Secret": {
		"version": "1"
	},
	"Map_Secret_Key": {
		"version": "latest"
	}
}`)

	invalidJsonBytes = []byte(`
{
	NOT VALID JSON
	"Text_Secret": {
		"version": "1"
	},
	"Map_Secret_Key": {
		"version": "latest"
	}
}`)

	validYamlBytes = []byte(`
Text_Secret:
    version: 1
Map_Secret_Key:
    version: latest
`)

	invalidYamlBytes = []byte(`
NOT VALID YAML
Text_Secret:
    version: 1
Map_Secret_Key:
    version: latest
`)
)

func newDefaultSpecificationFields() []internal.Field {
	return []internal.Field{
		{
			SecretType:    internal.DefaultType,
			SecretName:    "Text_Secret",
			SecretVersion: "latest",
			MapKeyName:    "",
			SplitWords:    true,
		},
		{
			SecretType:    "map",
			SecretName:    "Map_Secret",
			SecretVersion: "latest",
			MapKeyName:    "Key",
			SplitWords:    true,
		},
	}
}