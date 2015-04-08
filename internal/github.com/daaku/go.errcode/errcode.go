// Package errcode provides for errors with codes. This is handly for
// HTTP errors among other things.
package errcode

import (
	"fmt"
)

// An Error and a Code.
type Error interface {
	error
	Code() int
}

type errT struct {
	code int
	err  error
}

func (e errT) Error() string {
	return e.err.Error()
}

// Return the associated code.
func (e errT) Code() int {
	return e.code
}

// Create a new Coded Error.
func New(code int, f string, args ...interface{}) Error {
	return errT{
		code: code,
		err:  fmt.Errorf(f, args...),
	}
}

// Add a Code to an existing Error.
func Add(code int, err error) Error {
	return errT{code: code, err: err}
}

// Get a Code from an existing error if it is an Error, else return the
// provided default code.
func Get(err error, code int) int {
	if e, ok := err.(Error); ok {
		return e.Code()
	}
	return code
}
