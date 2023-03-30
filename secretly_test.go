package secretly

import (
	"errors"
	"fmt"
	"testing"
)

type TestingSpecification struct {
	ExampleField string `type:"map" split_words:"true" secret_name:"a_secret" key_name:"a_key" version:"latest"`
}

func TestParseSpecification(t *testing.T) {
	spec := TestingSpecification{}
	f, err := parseSpecification(&spec)
	fmt.Println(f)
	if err != nil {
		t.Errorf("Incorrect error. Want %v, got %v", nil, err)
	}
}

func TestParseSpecificationNonPointer(t *testing.T) {
	spec := TestingSpecification{}

	_, err := parseSpecification(spec)
	if err != nil {
		if !errors.Is(err, ErrInvalidSpecification) {
			t.Errorf("Incorrect error. Want %v, got %v", ErrInvalidSpecification, err)
		}
	}
}
