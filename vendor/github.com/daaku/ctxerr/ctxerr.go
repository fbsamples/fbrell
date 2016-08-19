// Package ctxerr is a hidden system designed to optionally augment errors
// with caller information.
//
// This allows for varying degrees of stack trace information of where the
// error occurred. It does so in an invisible manner, which means your return
// types are still plain errors, and the errors are only augmented when the
// caller wants them to be augmented. It does so by leveraging context.Context.
// This is nice for various reasons. One is cleanliness -- those who don't wish
// to these use augmented errors don't do so. The cost when turned off is 1
// context.Context Value lookup. Since capturing this information is often
// expensive, this also allows for sampling if you have enough volume.
// Similarly this enables a much nicer development process where stack traces
// can be enabled all the time.
package ctxerr

import (
	"context"
	"fmt"

	"github.com/facebookgo/stack"
)

// StackMode defines the modes for capturing stack traces.
// The default mode prevents all wrapping all together.
type StackMode int

const (
	// StackModeNone is the default and will result in no stack being included.
	StackModeNone StackMode = iota

	// StackModeSingleFrame triggers inclusion of a single frame, that is only
	// the immediate callers information.
	StackModeSingleFrame

	// StackModeSingleStack triggers inclusion of a single set of stack frames,
	// that is an entire chain of callers.
	StackModeSingleStack

	// StackModeMultiStack triggers inclusion of multiple set of stack frames,
	// that is many chains of callers. This is useful when an errors is being
	// passed between goroutines.
	StackModeMultiStack
)

// StringMode defines the behaviour of the Error() method with respect to stack
// traces. The default doesn't include any information at all. If the error
// isn't configured with an appropriate StackMode then it does nothing.
type StringMode int

const (
	// StringModeNone prevents any information from being included in the error
	// string. That is, the underlying error string is returned as is.
	StringModeNone StringMode = iota

	// StringModeAll includes all trace information.
	StringModeAll
)

// Config defines the per-context configuration for wrapping errors.
type Config struct {
	StackMode  StackMode
	StringMode StringMode
}

type configKeyT int

var configKey = configKeyT(1)

// ContextConfig returns the config from the context.
func ContextConfig(ctx context.Context) Config {
	ac, _ := ctx.Value(configKey).(Config)
	return ac
}

// WithConfig returns a new context with the specified Config.
func WithConfig(ctx context.Context, c Config) context.Context {
	return context.WithValue(ctx, configKey, c)
}

type singleFrameError struct {
	underlying error
	config     Config
	frame      stack.Frame
}

func (e *singleFrameError) Underlying() error { return e.underlying }

func (e *singleFrameError) Error() string {
	if e.config.StringMode == StringModeAll {
		return e.rich()
	}
	return e.underlying.Error()
}

func (e *singleFrameError) rich() string {
	return fmt.Sprintf("%s %s", e.frame, e.underlying)
}

type singleStackError struct {
	underlying error
	config     Config
	stack      stack.Stack
}

func (e *singleStackError) Underlying() error { return e.underlying }

func (e *singleStackError) Error() string {
	if e.config.StringMode == StringModeAll {
		return e.rich()
	}
	return e.underlying.Error()
}

func (e *singleStackError) rich() string {
	return fmt.Sprintf("%s\n%s", e.underlying, e.stack)
}

type multiStackError struct {
	underlying error
	config     Config
	multiStack *stack.Multi
}

func (e *multiStackError) Underlying() error { return e.underlying }

func (e *multiStackError) Error() string {
	if e.config.StringMode == StringModeAll {
		return e.rich()
	}
	return e.underlying.Error()
}

func (e *multiStackError) rich() string {
	return fmt.Sprintf("%s\n%s", e.underlying, e.multiStack)
}

// Wrap calls WrapSkip with skip=1.
func Wrap(ctx context.Context, err error) error {
	return WrapSkip(ctx, err, 1)
}

// WrapSkip may wrap the error and return an augmented error depending on the
// configuration in the context. The defaults result in the error being
// returned as is. It also handles nils correctly, returning a nil in that
// case. If the given error is already wrapped, it returns it as-is.
func WrapSkip(ctx context.Context, err error, skip int) error {
	switch err := err.(type) {
	case nil:
		return nil
	case *singleFrameError:
		return err
	case *singleStackError:
		return err
	case *multiStackError:
		err.multiStack.AddCallers(skip + 1)
		return err
	}

	config := ContextConfig(ctx)
	switch config.StackMode {
	case StackModeSingleFrame:
		return &singleFrameError{
			config:     config,
			underlying: err,
			frame:      stack.Caller(skip + 1),
		}
	case StackModeSingleStack:
		return &singleStackError{
			config:     config,
			underlying: err,
			stack:      stack.Callers(skip + 1),
		}
	case StackModeMultiStack:
		return &multiStackError{
			config:     config,
			underlying: err,
			multiStack: stack.CallersMulti(skip + 1),
		}
	}
	return err
}

// StackFrame returns the best possible stack.Frame from the error or nil if
// one isn't found.
func StackFrame(err error) *stack.Frame {
	switch err := err.(type) {
	case *singleFrameError:
		return &err.frame
	case *singleStackError:
		return &err.stack[0]
	case *multiStackError:
		return &err.multiStack.Stacks()[0][0]
	}
	return nil
}

// Stack returns the best possible stack.Stack from the error or nil if one
// isn't found.
func Stack(err error) stack.Stack {
	switch err := err.(type) {
	case *singleFrameError:
		return stack.Stack{err.frame}
	case *singleStackError:
		return err.stack
	case *multiStackError:
		return err.multiStack.Stacks()[0]
	}
	return nil
}

// MultiStack returns the best possible stack.Multi from the error or nil if
// one isn't found.
func MultiStack(err error) *stack.Multi {
	switch err := err.(type) {
	case *singleFrameError:
		var m stack.Multi
		m.Add(stack.Stack{err.frame})
		return &m
	case *singleStackError:
		var m stack.Multi
		m.Add(err.stack)
		return &m
	case *multiStackError:
		return err.multiStack
	}
	return nil
}

// RichString returns an error string with stack information if available. If
// none is available, it is the same as using the Error method on the error. It
// does so ignoring the context configuration. This is useful for example when
// internally logging an error.
func RichString(err error) string {
	type rich interface {
		rich() string
	}

	if re, ok := err.(rich); ok {
		return re.rich()
	}

	return err.Error()
}
