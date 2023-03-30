package secretly

import (
	"errors"
	"testing"
)

type CorrectSpecification struct {
	Text           string
	TextSplitWords string `split_words:"true"`
	TextSecretName string `secret_name:"a_secret"`
	TextVersion    string `version:"latest"`
	TextAll        string `secret_name:"a_secret" version:"latest" split_words:"true"`

	Map           string `type:"map"`
	MapSplitWords string `type:"map" split_words:"true"`
	MapSecretName string `type:"map" secret_name:"a_secret"`
	MapKeyName    string `type:"map" key_name:"a_key"`
	MapVersion    string `type:"map" version:"1"`
	MapAll        string `type:"map" secret_name:"a_secret" key_name:"a_key" version:"latest" split_words:"true"`
}

type TextWithKeyNameSpecification struct {
	TextKeyName string `secret_name:"a_secret" key_name:"a_key"`
}

func TestParsingCorrectSpecification(t *testing.T) {
	spec := CorrectSpecification{}
	_, err := parseSpecification(&spec)
	if err != nil {
		t.Errorf("Incorrect error. Want %v, got %v", nil, err)
	}
}

func TestParsingTextWithKeyNameSpecification(t *testing.T) {
	spec := TextWithKeyNameSpecification{}
	_, err := parseSpecification(&spec)
	if err != nil {
		if !errors.Is(err, ErrSecretTypeDoesNotSupportTagKey) {
			t.Errorf("Incorrect error. Want %v, got %v", ErrSecretTypeDoesNotSupportTagKey, err)
		}
	}
}

func TestParsingNonPointerSpecification(t *testing.T) {
	spec := CorrectSpecification{}

	_, err := parseSpecification(spec)
	if err != nil {
		if !errors.Is(err, ErrInvalidSpecification) {
			t.Errorf("Incorrect error. Want %v, got %v", ErrInvalidSpecification, err)
		}
	}
}
