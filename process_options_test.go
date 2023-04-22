package secretly

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

type TestingSpecification struct {
	TextSecret string `split_words:"true"`
	JsonSecret string `type:"json" key:"Key" split_words:"true"`
	YamlSecret string `type:"yaml" key:"Key"`
}

func newTestingSpecificationFields() []Field {
	return []Field{
		{
			SecretType:    DefaultType,
			SecretName:    "Text_Secret",
			SecretVersion: "latest",
			MapKeyName:    "",
			SplitWords:    true,
		},
		{
			SecretType:    "json",
			SecretName:    "Json_Secret",
			SecretVersion: "latest",
			MapKeyName:    "Key",
			SplitWords:    true,
		},
		{
			SecretType:    "yaml",
			SecretName:    "YamlSecret",
			SecretVersion: "latest",
			MapKeyName:    "Key",
			SplitWords:    false,
		},
	}
}

func TestApplyConfig(t *testing.T) {
	tests := []struct {
		name          string
		unmarshalFunc unmarshalFunc
		content       []byte
		want          []Field
		wantErr       bool
	}{
		{
			name:          "Only Versions JSON",
			unmarshalFunc: json.Unmarshal,
			content:       onlyVersionsJsonBytes,
			want: []Field{
				{
					SecretType:    DefaultType,
					SecretName:    "Text_Secret",
					SecretVersion: "1",
					MapKeyName:    "",
					SplitWords:    true,
				},
				{
					SecretType:    "json",
					SecretName:    "Json_Secret",
					SecretVersion: "latest",
					MapKeyName:    "Key",
					SplitWords:    true,
				},
				{
					SecretType:    "yaml",
					SecretName:    "YamlSecret",
					SecretVersion: "latest",
					MapKeyName:    "Key",
					SplitWords:    false,
				},
			},
			wantErr: false,
		},
		{
			name:          "All Fields JSON",
			unmarshalFunc: json.Unmarshal,
			content:       allFieldsJsonBytes,
			want: []Field{
				{
					SecretType:    DefaultType,
					SecretName:    "Text_Secret_Overwritten",
					SecretVersion: "1",
					MapKeyName:    "",
					SplitWords:    true,
				},
				{
					SecretType:    "json",
					SecretName:    "Json_Secret_Overwritten",
					SecretVersion: "latest",
					MapKeyName:    "Key_Overwritten",
					SplitWords:    true,
				},
				{
					SecretType:    "yaml",
					SecretName:    "YamlSecret_Overwritten",
					SecretVersion: "latest",
					MapKeyName:    "Key_Overwritten",
					SplitWords:    true,
				},
			},
			wantErr: false,
		},
		{
			name:          "Invalid JSON",
			unmarshalFunc: json.Unmarshal,
			content:       invalidJsonBytes,
			want:          newTestingSpecificationFields(),
			wantErr:       true,
		},
		{
			name:          "Only Versions YAML",
			unmarshalFunc: yaml.Unmarshal,
			content:       onlyVersionsYamlBytes,
			want: []Field{
				{
					SecretType:    DefaultType,
					SecretName:    "Text_Secret",
					SecretVersion: "1",
					MapKeyName:    "",
					SplitWords:    true,
				},
				{
					SecretType:    "json",
					SecretName:    "Json_Secret",
					SecretVersion: "latest",
					MapKeyName:    "Key",
					SplitWords:    true,
				},
				{
					SecretType:    "yaml",
					SecretName:    "YamlSecret",
					SecretVersion: "latest",
					MapKeyName:    "Key",
					SplitWords:    false,
				},
			},
			wantErr: false,
		},
		{
			name:          "All Fields Yaml",
			unmarshalFunc: yaml.Unmarshal,
			content:       allFieldsYamlBytes,
			want: []Field{
				{
					SecretType:    DefaultType,
					SecretName:    "Text_Secret_Overwritten",
					SecretVersion: "1",
					MapKeyName:    "",
					SplitWords:    true,
				},
				{
					SecretType:    "json",
					SecretName:    "Json_Secret_Overwritten",
					SecretVersion: "latest",
					MapKeyName:    "Key_Overwritten",
					SplitWords:    true,
				},
				{
					SecretType:    "yaml",
					SecretName:    "YamlSecret_Overwritten",
					SecretVersion: "latest",
					MapKeyName:    "Key_Overwritten",
					SplitWords:    true,
				},
			},
			wantErr: false,
		},
		{
			name:          "Invalid YAML",
			unmarshalFunc: yaml.Unmarshal,
			content:       invalidYamlBytes,
			want:          newTestingSpecificationFields(),
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields := newTestingSpecificationFields()
			err := setFieldsWithConfig(tt.unmarshalFunc, tt.content, fields)

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
		want      []Field
		wantErr   error
	}{
		{
			name:   "All Env Vars Exist",
			prefix: "TEST",
			envVarMap: map[string]string{
				"TEST_TEXT_SECRET_VERSION":     "1",
				"TEST_JSON_SECRET_KEY_VERSION": "latest",
				"TEST_YAMLSECRETKEY_VERSION":   "latest",
			},
			want: []Field{
				{
					SecretType:    DefaultType,
					SecretName:    "Text_Secret",
					SecretVersion: "1",
					MapKeyName:    "",
					SplitWords:    true,
				},
				{
					SecretType:    "json",
					SecretName:    "Json_Secret",
					SecretVersion: "latest",
					MapKeyName:    "Key",
					SplitWords:    true,
				},
				{
					SecretType:    "yaml",
					SecretName:    "YamlSecret",
					SecretVersion: "latest",
					MapKeyName:    "Key",
					SplitWords:    false,
				},
			},
			wantErr: nil,
		},
		{
			name:   "One Env Var Missing",
			prefix: "TEST",
			envVarMap: map[string]string{
				"TEST_JSON_SECRET_KEY_VERSION": "latest",
				"TEST_YAMLSECRETKEY_VERSION":   "latest",
			},
			want: []Field{
				{
					SecretType:    DefaultType,
					SecretName:    "Text_Secret",
					SecretVersion: "0",
					MapKeyName:    "",
					SplitWords:    true,
				},
				{
					SecretType:    "json",
					SecretName:    "Json_Secret",
					SecretVersion: "latest",
					MapKeyName:    "Key",
					SplitWords:    true,
				},
				{
					SecretType:    "yaml",
					SecretName:    "YamlSecret",
					SecretVersion: "latest",
					MapKeyName:    "Key",
					SplitWords:    false,
				},
			},
			wantErr: nil,
		},
		{
			name:      "All Env Vars Missing",
			prefix:    "TEST",
			envVarMap: map[string]string{},
			want: []Field{
				{
					SecretType:    DefaultType,
					SecretName:    "Text_Secret",
					SecretVersion: "0",
					MapKeyName:    "",
					SplitWords:    true,
				},
				{
					SecretType:    "json",
					SecretName:    "Json_Secret",
					SecretVersion: "0",
					MapKeyName:    "Key",
					SplitWords:    true,
				},
				{
					SecretType:    "yaml",
					SecretName:    "YamlSecret",
					SecretVersion: "0",
					MapKeyName:    "Key",
					SplitWords:    false,
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
			fields, err := Process(&spec, WithVersionsFromEnv(tt.prefix))

			if err != tt.wantErr {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Incorrect error. Want %v, got %v", tt.wantErr, err)
				}
			}

			// Don't check reflect.Value for equality
			for i, field := range fields {
				v := reflect.Value{}
				if field.Value == v {
					t.Errorf("Incorrect field.Value. Got the zero value, should be something else")
				}
				fields[i].Value = v
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

var (
	onlyVersionsJsonBytes = []byte(`
{
	"Text_Secret": {
		"version": "1"
	},
	"Json_Secret_Key": {
		"version": "latest"
	},
	"YamlSecretKey": {
		"version": "latest"
	}
}
`)

	allFieldsJsonBytes = []byte(`
{
	"Text_Secret": {
		"name": "Text_Secret_Overwritten",
		"version": "1",
		"split_words": true
	},
	"Json_Secret_Key": {
		"name": "Json_Secret_Overwritten",
		"key": "Key_Overwritten",
		"version": "latest",
		"split_words": true
	},
	"YamlSecretKey": {
		"name": "YamlSecret_Overwritten",
		"key": "Key_Overwritten",
		"version": "latest",
		"split_words": true
	}
}
`)

	invalidJsonBytes = []byte(`NOT VALID JSON`)

	onlyVersionsYamlBytes = []byte(`
Text_Secret:
    version: 1
Json_Secret_Key:
    version: latest
YamlSecretKey:
    version: latest
`)

	allFieldsYamlBytes = []byte(`
Text_Secret:
    name: Text_Secret_Overwritten
    version: 1
    split_words: true
Json_Secret_Key:
    name: Json_Secret_Overwritten
    key: Key_Overwritten
    version: latest
    split_words: true
YamlSecretKey:
    name: YamlSecret_Overwritten
    key: Key_Overwritten
    version: latest
    split_words: true
`)

	invalidYamlBytes = []byte(`NOT VALID YAML`)
)
