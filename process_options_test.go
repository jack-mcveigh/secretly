package secretly

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/jack-mcveigh/secretly/internal"
	"gopkg.in/yaml.v3"
)

type TestingSpecification struct {
	TextSecret string `split_words:"true"`
	MapSecret  string `type:"map" key_name:"Key" split_words:"true"`
}

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
			want:          correctSpecificationFields,
			wantErr:       false,
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
			want:          correctSpecificationFields,
			wantErr:       false,
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
	}
}

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

var (
	correctSpecificationFields = []internal.Field{
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
	}

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
