package secretly

import (
	"testing"
)

func TestWithDefaultVersion(t *testing.T) {
	const newDefault = "current"

	fs := fields{
		{secretVersion: "0"},
		{secretVersion: "1"},
	}

	err := WithDefaultVersion(newDefault)(fs)
	if err != nil {
		t.Fatalf("Incorrect error. Want %v, Got %v", err, nil)
	}

	if fs[0].secretVersion != newDefault {
		t.Fatalf("Incorrect fields[0].SecretVersion. Want %v, Got %v", newDefault, fs[0].secretVersion)
	}

	if fs[1].secretVersion != "1" {
		t.Fatalf("Incorrect fields[0].SecretVersion. Want %v, Got %v", newDefault, fs[0].secretVersion)
	}
}

func TestWithCache(t *testing.T) {
	fs := fields{{}}

	err := WithCache()(fs)
	if err != nil {
		t.Fatalf("Incorrect error. Want %v, Got %v", err, nil)
	}

	if fs[0].cache == nil {
		t.Fatalf("Incorrect fields[0].Cache. Got %v", fs)
	}
}

func TestWithPatch(t *testing.T) {
	fs := fields{
		field{
			secretType:    DefaultType,
			secretName:    "TopSecret",
			secretVersion: DefaultVersion,
			mapKeyName:    "",
		},
		field{
			secretType:    JSON,
			secretName:    "TopSecret",
			secretVersion: DefaultVersion,
			mapKeyName:    "SpecificSecret",
			splitWords:    true,
		},
	}

	patch := []byte(`{
	"TopSecret": {
		"version": "latest"
	},
	"Top_Secret_Specific_Secret": {
		"version": "1"
	}
	}`)

	err := WithPatch(patch)(fs)
	if err != nil {
		t.Fatalf("Incorrect error. Want %v, Got %v", nil, err)
	}

	if fs[0].secretVersion != "latest" {
		t.Errorf("Incorrect fields[0].SecretVersion. Want %v, got %v", "latest", fs[0].secretVersion)
	}

	if fs[1].secretVersion != "1" {
		t.Errorf("Incorrect fields[1].SecretVersion. Want %v, got %v", "1", fs[1].secretVersion)
	}
}

func TestWithPatchFile(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
	}{
		{
			name:     "JSON Patch File",
			fileName: "testdata/patch_0.json",
		},
		{
			name:     "YAML Patch File",
			fileName: "testdata/patch_1.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := fields{
				field{
					secretType:    DefaultType,
					secretName:    "TopSecret",
					secretVersion: DefaultVersion,
					mapKeyName:    "",
				},
				field{
					secretType:    JSON,
					secretName:    "TopSecret",
					secretVersion: DefaultVersion,
					mapKeyName:    "SpecificSecret",
				},
				field{
					secretType:    YAML,
					secretName:    "AnotherTopSecret",
					secretVersion: DefaultVersion,
					mapKeyName:    "SpecificSecret",
					splitWords:    true,
				},
			}

			err := WithPatchFile(tt.fileName)(fs)
			if err != nil {
				t.Fatalf("Incorrect error. Want %v, Got %v", nil, err)
			}

			if fs[0].secretVersion != "latest" {
				t.Errorf("Incorrect fields[0].SecretVersion. Want %v, got %v", "latest", fs[0].secretVersion)
			}

			if fs[1].secretVersion != "1" {
				t.Errorf("Incorrect fields[1].SecretVersion. Want %v, got %v", "1", fs[1].secretVersion)
			}

			if fs[2].secretVersion != "100" {
				t.Errorf("Incorrect fields[2].SecretVersion. Want %v, got %v", "1", fs[2].secretVersion)
			}
		})
	}
}

func TestWithVersionsFromEnv(t *testing.T) {
	fs := fields{
		field{
			secretType:    DefaultType,
			secretName:    "SECRET",
			secretVersion: DefaultVersion,
			mapKeyName:    "",
		},
		field{
			secretType:    JSON,
			secretName:    "SUPER",
			secretVersion: DefaultVersion,
			mapKeyName:    "SECRET",
		},
	}

	t.Setenv("TEST_SECRET_VERSION", "latest")
	t.Setenv("TEST_SUPERSECRET_VERSION", "1")

	err := WithVersionsFromEnv("TEST")(fs)
	if err != nil {
		t.Fatalf("Incorrect error. Want %v, Got %v", nil, err)
	}

	if fs[0].secretVersion != "latest" {
		t.Errorf("Incorrect fields[0].SecretVersion. Want %v, got %v", "latest", fs[0].secretVersion)
	}

	if fs[1].secretVersion != "1" {
		t.Errorf("Incorrect fields[1].SecretVersion. Want %v, got %v", "latest", fs[1].secretVersion)
	}
}
