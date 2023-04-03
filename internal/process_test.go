package internal

import (
	"errors"
	"reflect"
	"testing"
)

type CorrectSpecification struct {
	Text           string
	TextSplitWords string `split_words:"true"`
	TextSecretName string `secret_name:"a_secret"`
	TextVersion    string `version:"latest"`
	// NOTE: split_words doesn't do anything when secret_name is provided
	TextAll string `secret_name:"a_secret" version:"latest" split_words:"true"`

	Map           string `type:"map"`
	MapSplitWords string `type:"map" split_words:"true"`
	MapSecretName string `type:"map" secret_name:"a_secret"`
	MapKeyName    string `type:"map" key_name:"a_key"`
	MapVersion    string `type:"map" version:"1"`
	// NOTE: split_words doesn't do anything when secret_name and key_name is provided
	MapAll string `type:"map" secret_name:"a_secret" key_name:"a_key" version:"latest" split_words:"true"`

	Ignored string `ignored:"true"`
	ignored string
}

type TextWithKeyNameSpecification struct {
	TextKeyName string `secret_name:"a_secret" key_name:"a_key"`
}

func TestParsingCorrectSpecification(t *testing.T) {
	want := correctSpecificationFields

	spec := CorrectSpecification{ignored: "testing"}
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
	{
		SecretType:    "map",
		SecretName:    "Map",
		SecretVersion: DefaultVersion,
		MapKeyName:    "Map",
		SplitWords:    false,
	},
	{
		SecretType:    "map",
		SecretName:    "Map_Split_Words",
		SecretVersion: DefaultVersion,
		MapKeyName:    "Map_Split_Words",
		SplitWords:    true,
	},
	{
		SecretType:    "map",
		SecretName:    "a_secret",
		SecretVersion: DefaultVersion,
		MapKeyName:    "MapSecretName",
		SplitWords:    false,
	},
	{
		SecretType:    "map",
		SecretName:    "MapKeyName",
		SecretVersion: DefaultVersion,
		MapKeyName:    "a_key",
		SplitWords:    false,
	},
	{
		SecretType:    "map",
		SecretName:    "MapVersion",
		SecretVersion: "1",
		MapKeyName:    "MapVersion",
		SplitWords:    false,
	},
	{
		SecretType:    "map",
		SecretName:    "a_secret",
		SecretVersion: "latest",
		MapKeyName:    "a_key",
		SplitWords:    true,
	},
}
