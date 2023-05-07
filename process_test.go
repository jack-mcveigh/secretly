package secretly

import (
	"errors"
	"reflect"
	"testing"
)

// exported to allow testing embedding support
type CorrectSubSpecification struct {
	Text string
}

type correctSpecification struct {
	Text           string
	TextSplitWords string `split_words:"true"`
	TextSecretName string `name:"a_secret"`
	TextVersion    string `version:"latest"`
	// NOTE: split_words doesn't do anything when name is provided.
	// It only modifies the default name and key, the struct field name.
	TextAll string `name:"a_secret" version:"latest" split_words:"true"`

	Json           string `type:"json"`
	JsonSplitWords string `type:"json" split_words:"true"`
	JsonSecretName string `type:"json" name:"a_secret"`
	JsonKeyName    string `type:"json" key:"a_key"`
	JsonVersion    string `type:"json" version:"1"`
	// NOTE: split_words doesn't do anything when name and key is provided.
	// It only modifies the default name and key, the struct field name.
	JsonAll string `type:"json" name:"a_secret" key:"a_key" version:"latest" split_words:"true"`

	Yaml           string `type:"yaml"`
	YamlSplitWords string `type:"yaml" split_words:"true"`
	YamlSecretName string `type:"yaml" name:"a_secret"`
	YamlKeyName    string `type:"yaml" key:"a_key"`
	YamlVersion    string `type:"yaml" version:"1"`
	// NOTE: split_words doesn't do anything when name and key is provided.
	// It only modifies the default name and key, the struct field name.
	YamlAll string `type:"yaml" name:"a_secret" key:"a_key" version:"latest" split_words:"true"`

	Pointer *string

	ComposedSpecification CorrectSubSpecification

	ComposedSpecificationPtr *CorrectSubSpecification

	CorrectSubSpecification // test embedding

	Ignored                      string `ignored:"true"`
	ignored                      string
	IgnoredComposedSpecification CorrectSubSpecification `ignored:"true"`
	ignoredComposedSpecification CorrectSubSpecification
}

type TextWithKeyNameSpecification struct {
	TextKeyName string `name:"a_secret" key:"a_key"`
}

func TestParsingCorrectSpecification(t *testing.T) {
	want := correctSpecificationFields

	spec := correctSpecification{ignored: "testing", ignoredComposedSpecification: CorrectSubSpecification{}}
	got, err := processSpec(&spec)
	if err != nil {
		t.Errorf("Incorrect error. Want %v, got %v", nil, err)
	}

	// Ensure reflect.Value is not already the zero value.
	// If it isn't set it to the zero value
	for i, field := range got {
		v := reflect.Value{}
		if field.Value == v {
			t.Errorf("Incorrect field.Value. Got the zero value, should be something else")
		}
		got[i].Value = v
	}

	if !reflect.DeepEqual(want, got) {
		t.Errorf("Incorrect fields. Want %v, got %v", want, got)
	}
}

func TestParsingTextWithKeyNameSpecification(t *testing.T) {
	spec := TextWithKeyNameSpecification{}
	_, err := processSpec(&spec)
	if err != nil {
		if !errors.Is(err, ErrSecretTypeDoesNotSupportKey) {
			t.Errorf("Incorrect error. Want %v, got %v", ErrSecretTypeDoesNotSupportKey, err)
		}
	}
}

func TestParsingNonPointerSpecification(t *testing.T) {
	spec := correctSpecification{}

	_, err := processSpec(spec)
	if err != nil {
		if !errors.Is(err, ErrInvalidSpecification) {
			t.Errorf("Incorrect error. Want %v, got %v", ErrInvalidSpecification, err)
		}
	}
}

var correctSpecificationFields = []Field{
	// text
	{
		SecretType:    DefaultType,
		SecretName:    "Text",
		SecretVersion: DefaultVersion,
		MapKeyName:    "",
		SplitWords:    false,
	},
	{
		SecretType:    DefaultType,
		SecretName:    "Text_Split_Words",
		SecretVersion: DefaultVersion,
		MapKeyName:    "",
		SplitWords:    true,
	},
	{
		SecretType:    DefaultType,
		SecretName:    "a_secret",
		SecretVersion: DefaultVersion,
		MapKeyName:    "",
		SplitWords:    false,
	},
	{
		SecretType:    DefaultType,
		SecretName:    "TextVersion",
		SecretVersion: "latest",
		MapKeyName:    "",
		SplitWords:    false,
	},
	{
		SecretType:    DefaultType,
		SecretName:    "a_secret",
		SecretVersion: "latest",
		MapKeyName:    "",
		SplitWords:    true,
	},
	// json
	{
		SecretType:    "json",
		SecretName:    "Json",
		SecretVersion: DefaultVersion,
		MapKeyName:    "Json",
		SplitWords:    false,
	},
	{
		SecretType:    "json",
		SecretName:    "Json_Split_Words",
		SecretVersion: DefaultVersion,
		MapKeyName:    "Json_Split_Words",
		SplitWords:    true,
	},
	{
		SecretType:    "json",
		SecretName:    "a_secret",
		SecretVersion: DefaultVersion,
		MapKeyName:    "JsonSecretName",
		SplitWords:    false,
	},
	{
		SecretType:    "json",
		SecretName:    "JsonKeyName",
		SecretVersion: DefaultVersion,
		MapKeyName:    "a_key",
		SplitWords:    false,
	},
	{
		SecretType:    "json",
		SecretName:    "JsonVersion",
		SecretVersion: "1",
		MapKeyName:    "JsonVersion",
		SplitWords:    false,
	},
	{
		SecretType:    "json",
		SecretName:    "a_secret",
		SecretVersion: "latest",
		MapKeyName:    "a_key",
		SplitWords:    true,
	},
	// yaml
	{
		SecretType:    "yaml",
		SecretName:    "Yaml",
		SecretVersion: DefaultVersion,
		MapKeyName:    "Yaml",
		SplitWords:    false,
	},
	{
		SecretType:    "yaml",
		SecretName:    "Yaml_Split_Words",
		SecretVersion: DefaultVersion,
		MapKeyName:    "Yaml_Split_Words",
		SplitWords:    true,
	},
	{
		SecretType:    "yaml",
		SecretName:    "a_secret",
		SecretVersion: DefaultVersion,
		MapKeyName:    "YamlSecretName",
		SplitWords:    false,
	},
	{
		SecretType:    "yaml",
		SecretName:    "YamlKeyName",
		SecretVersion: DefaultVersion,
		MapKeyName:    "a_key",
		SplitWords:    false,
	},
	{
		SecretType:    "yaml",
		SecretName:    "YamlVersion",
		SecretVersion: "1",
		MapKeyName:    "YamlVersion",
		SplitWords:    false,
	},
	{
		SecretType:    "yaml",
		SecretName:    "a_secret",
		SecretVersion: "latest",
		MapKeyName:    "a_key",
		SplitWords:    true,
	},
	{
		SecretType:    "text",
		SecretName:    "Pointer",
		SecretVersion: "0",
		MapKeyName:    "",
		SplitWords:    false,
	},
	{
		SecretType:    "text",
		SecretName:    "Text",
		SecretVersion: "0",
		MapKeyName:    "",
		SplitWords:    false,
	},
	{
		SecretType:    "text",
		SecretName:    "Text",
		SecretVersion: "0",
		MapKeyName:    "",
		SplitWords:    false,
	},
	{
		SecretType:    "text",
		SecretName:    "Text",
		SecretVersion: "0",
		MapKeyName:    "",
		SplitWords:    false,
	},
}
