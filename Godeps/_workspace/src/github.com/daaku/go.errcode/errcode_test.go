package errcode_test

import (
	"errors"
	"github.com/daaku/rell/Godeps/_workspace/src/github.com/daaku/go.errcode"
	"testing"
)

func TestNew(t *testing.T) {
	const expectedCode = 500
	const expectedMessage = "foo"
	err := errcode.New(expectedCode, expectedMessage)

	actualCode := err.Code()
	if actualCode != expectedCode {
		t.Errorf(
			"did not find expected code %d instead found %d",
			expectedCode,
			actualCode,
		)
	}

	actualMessage := err.Error()
	if actualMessage != expectedMessage {
		t.Errorf(
			"did not find expected message %s instead found %s",
			expectedMessage,
			actualMessage,
		)
	}
}

func TestAdd(t *testing.T) {
	const expectedCode = 500
	const expectedMessage = "foo"
	err := errcode.Add(expectedCode, errors.New(expectedMessage))

	actualCode := err.Code()
	if actualCode != expectedCode {
		t.Errorf(
			"did not find expected code %d instead found %d",
			expectedCode,
			actualCode,
		)
	}

	actualMessage := err.Error()
	if actualMessage != expectedMessage {
		t.Errorf(
			"did not find expected message %s instead found %s",
			expectedMessage,
			actualMessage,
		)
	}
}

func TestGetWithCode(t *testing.T) {
	const expectedCode = 42
	err := errcode.New(expectedCode, "foo")
	actualCode := errcode.Get(err, -1)
	if actualCode != expectedCode {
		t.Errorf(
			"did not find expected code %d instead found %d",
			expectedCode,
			actualCode,
		)
	}
}

func TestGetDefault(t *testing.T) {
	const expectedCode = 42
	err := errors.New("foo")
	actualCode := errcode.Get(err, expectedCode)
	if actualCode != expectedCode {
		t.Errorf(
			"did not find expected code %d instead found %d",
			expectedCode,
			actualCode,
		)
	}
}
