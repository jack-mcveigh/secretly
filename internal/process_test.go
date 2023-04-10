package internal

import (
	"errors"
	"reflect"
	"testing"
)

type CorrectSubSpecification struct {
	Text string
}

type CorrectSpecification struct {
	Text           string
	TextSplitWords string `split_words:"true"`
	TextSecretName string `secret_name:"a_secret"`
	TextVersion    string `version:"latest"`
	// NOTE: split_words doesn't do anything when secret_name is provided
	TextAll string `secret_name:"a_secret" version:"latest" split_words:"true"`

	Json           string `type:"json"`
	JsonSplitWords string `type:"json" split_words:"true"`
	JsonSecretName string `type:"json" secret_name:"a_secret"`
	JsonKeyName    string `type:"json" key_name:"a_key"`
	JsonVersion    string `type:"json" version:"1"`
	// NOTE: split_words doesn't do anything when secret_name and key_name is provided
	JsonAll string `type:"json" secret_name:"a_secret" key_name:"a_key" version:"latest" split_words:"true"`

	Yaml           string `type:"yaml"`
	YamlSplitWords string `type:"yaml" split_words:"true"`
	YamlSecretName string `type:"yaml" secret_name:"a_secret"`
	YamlKeyName    string `type:"yaml" key_name:"a_key"`
	YamlVersion    string `type:"yaml" version:"1"`
	// NOTE: split_words doesn't do anything when secret_name and key_name is provided
	YamlAll string `type:"yaml" secret_name:"a_secret" key_name:"a_key" version:"latest" split_words:"true"`

	ComposedSpecification CorrectSubSpecification

	CorrectSubSpecification // test embedding

	Ignored                      string `ignored:"true"`
	ignored                      string
	IgnoredComposedSpecification CorrectSubSpecification `ignored:"true"`
	ignoredComposedSpecification CorrectSubSpecification
}

type TextWithKeyNameSpecification struct {
	TextKeyName string `secret_name:"a_secret" key_name:"a_key"`
}

func TestParsingCorrectSpecification(t *testing.T) {
	want := correctSpecificationFields

	spec := CorrectSpecification{ignored: "testing", ignoredComposedSpecification: CorrectSubSpecification{}}
	got, err := Process(&spec)
	if err != nil {
		t.Errorf("Incorrect error. Want %v, got %v", nil, err)
	}

	// Don't check reflect.Value for equality
	for i := range got {
		got[i].Value = reflect.Value{}
	}

	if !reflect.DeepEqual(want, got) {
		t.Errorf("Incorrect fields. Want %v, got %v", want, got)
	}
}

func TestParsingTextWithKeyNameSpecification(t *testing.T) {
	spec := TextWithKeyNameSpecification{}
	_, err := Process(&spec)
	if err != nil {
		if !errors.Is(err, ErrSecretTypeDoesNotSupportTagKey) {
			t.Errorf("Incorrect error. Want %v, got %v", ErrSecretTypeDoesNotSupportTagKey, err)
		}
	}
}

func TestParsingNonPointerSpecification(t *testing.T) {
	spec := CorrectSpecification{}

	_, err := Process(spec)
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
