package secretly

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

var errGetSecret = errors.New("get secret error")

func getSecretFromMapManager(secrets map[string]map[string]string, err error) GetSecretFunc {
	return func(ctx context.Context, name, version string) ([]byte, error) {
		return []byte(secrets[name][version]), err
	}
}

func TestProcess(t *testing.T) {
	type SubSpecification struct {
		SubField string
	}

	type specification struct {
		Field                       string
		IgnoredField                string `ignored:"true"`
		unexportedField             string
		SplitField                  string `split_words:"true"`
		VersionedField              string `version:"1"`
		VersionedField2             string `version:"latest"`
		JSONField                   string `type:"json" key:"Field1"`
		YAMLField                   string `type:"json" key:"Field2"`
		SubSpecificationNotEmbedded SubSpecification
		SubSpecification
	}
	_ = specification{unexportedField: ""} // Just to hide the staticcheck warning...

	var secretsMap = map[string]map[string]string{
		"Field": {
			"0": "field secret",
		},
		"IgnoredField": {
			"0": "ignored field secret",
		},
		"unexportedField": {
			"0": "unexported field secret",
		},
		"Split_Field": {
			"0": "split field secret",
		},
		"VersionedField": {
			"1": "versioned field secret",
		},
		"VersionedField2": {
			"latest": "versioned field secret 2",
		},
		"JSONField": {
			"0": `{
					"Field1": "json field's field 1 secret",
					"Field2": "json field's field 2 secret"
				}`,
		},
		"YAMLField": {
			"0": `{
					"Field1": "yaml field's field 1 secret",
					"Field2": "yaml field's field 2 secret"
				}`,
		},
		"SubField": {
			"0": "sub-specification sub-field secret",
		},
	}

	tests := []struct {
		name      string
		getSecret GetSecretFunc
		wantSpec  specification
		wantErr   error
	}{
		{
			name:      "Simple",
			getSecret: getSecretFromMapManager(secretsMap, nil),
			wantSpec: specification{
				Field:           "field secret",
				IgnoredField:    "",
				SplitField:      "split field secret",
				VersionedField:  "versioned field secret",
				VersionedField2: "versioned field secret 2",
				JSONField:       "json field's field 1 secret",
				YAMLField:       "yaml field's field 2 secret",
				SubSpecificationNotEmbedded: SubSpecification{
					SubField: "sub-specification sub-field secret",
				},
				SubSpecification: SubSpecification{
					SubField: "sub-specification sub-field secret",
				},
			},
			wantErr: nil,
		},
		{
			name:      "GetSecret Error",
			getSecret: getSecretFromMapManager(nil, errGetSecret),
			wantSpec:  specification{},
			wantErr:   errGetSecret,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var spec specification
			err := Process(context.Background(), &spec, tt.getSecret)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Incorrect error. Want %+v, got %+v", tt.wantErr, err)
			}

			if !reflect.DeepEqual(tt.wantSpec, spec) {
				t.Fatalf("Incorrect specification. Want %v, got %v", tt.wantSpec, spec)
			}
		})
	}
}

func TestProcessWithPatch(t *testing.T) {
	type SubSpecification struct {
		SubField string
	}

	type specification struct {
		Field                       string
		IgnoredField                string `ignored:"true"`
		unexportedField             string
		SplitField                  string `split_words:"true"`
		VersionedField              string `version:"1"`
		VersionedField2             string `version:"latest"`
		JSONField                   string `type:"json" key:"Field1"`
		YAMLField                   string `type:"json" key:"Field2"`
		SubSpecificationNotEmbedded SubSpecification
		SubSpecification
	}
	_ = specification{unexportedField: ""} // Just to hide the staticcheck warning...

	var secretsMap = map[string]map[string]string{
		"Field": {
			"0": "field secret",
		},
		"IgnoredField": {
			"0": "ignored field secret",
		},
		"unexportedField": {
			"0": "unexported field secret",
		},
		"Split_Field": {
			"0": "split field secret",
		},
		"VersionedField": {
			"1": "versioned field secret",
		},
		"VersionedField2": {
			"latest": "versioned field secret 2",
		},
		"JSONField": {
			"0": `{
					"Field1": "json field's field 1 secret",
					"Field2": "json field's field 2 secret"
				}`,
		},
		"YAMLField": {
			"0": `{
					"Field1": "yaml field's field 1 secret",
					"Field2": "yaml field's field 2 secret"
				}`,
		},
		"SubField": {
			"0": "sub-specification sub-field secret",
		},
	}

	tests := []struct {
		name      string
		getSecret GetSecretFunc
		wantSpec  specification
		wantErr   error
	}{
		{
			name:      "Simple",
			getSecret: getSecretFromMapManager(secretsMap, nil),
			wantSpec: specification{
				Field:           "field secret",
				IgnoredField:    "",
				SplitField:      "split field secret",
				VersionedField:  "versioned field secret",
				VersionedField2: "versioned field secret 2",
				JSONField:       "json field's field 1 secret",
				YAMLField:       "yaml field's field 2 secret",
				SubSpecificationNotEmbedded: SubSpecification{
					SubField: "sub-specification sub-field secret",
				},
				SubSpecification: SubSpecification{
					SubField: "sub-specification sub-field secret",
				},
			},
			wantErr: nil,
		},
		{
			name:      "GetSecret Error",
			getSecret: getSecretFromMapManager(nil, errGetSecret),
			wantSpec:  specification{},
			wantErr:   errGetSecret,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var spec specification
			err := Process(context.Background(), &spec, tt.getSecret)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Incorrect error. Want %+v, got %+v", tt.wantErr, err)
			}

			if !reflect.DeepEqual(tt.wantSpec, spec) {
				t.Fatalf("Incorrect specification. Want %v, got %v", tt.wantSpec, spec)
			}
		})
	}
}
