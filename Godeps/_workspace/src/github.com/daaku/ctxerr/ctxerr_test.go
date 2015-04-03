package ctxerr

import (
	"errors"
	"testing"

	"github.com/daaku/rell/Godeps/_workspace/src/github.com/facebookgo/ensure"
	"github.com/daaku/rell/Godeps/_workspace/src/github.com/facebookgo/stack"
	"github.com/daaku/rell/Godeps/_workspace/src/golang.org/x/net/context"
)

func TestContextConfig(t *testing.T) {
	ensure.DeepEqual(t, ContextConfig(context.Background()), Config{})
}

func TestWithConfig(t *testing.T) {
	ac := Config{StackMode: StackModeMultiStack}
	ctx := WithConfig(context.Background(), ac)
	ensure.DeepEqual(t, ContextConfig(ctx), ac)
}

func TestSingleFrameUnderlying(t *testing.T) {
	err := errors.New("")
	ensure.DeepEqual(t, (&singleFrameError{underlying: err}).Underlying(), err)
}

func TestSingleFrameStringModeAll(t *testing.T) {
	err := errors.New("foo")
	ensure.DeepEqual(
		t,
		(&singleFrameError{
			underlying: err,
			config: Config{
				StringMode: StringModeAll,
			},
			frame: stack.Frame{
				File: "f",
				Line: 1,
				Name: "n",
			},
		}).Error(),
		`f:1 n foo`)
}

func TestSingleFrameStringModeNone(t *testing.T) {
	err := errors.New("foo")
	ensure.DeepEqual(
		t,
		(&singleFrameError{
			underlying: err,
			config:     Config{},
			frame: stack.Frame{
				File: "f",
				Line: 1,
				Name: "n",
			},
		}).Error(),
		`foo`)
}

func TestSingleStackUnderlying(t *testing.T) {
	err := errors.New("")
	ensure.DeepEqual(t, (&singleStackError{underlying: err}).Underlying(), err)
}

func TestSingleStackStringModeAll(t *testing.T) {
	err := errors.New("foo")
	ensure.DeepEqual(
		t,
		(&singleStackError{
			underlying: err,
			config: Config{
				StringMode: StringModeAll,
			},
			stack: stack.Stack{
				stack.Frame{
					File: "f",
					Line: 1,
					Name: "n",
				},
			},
		}).Error(),
		"foo\nf:1 n")
}

func TestSingleStackStringModeNone(t *testing.T) {
	err := errors.New("foo")
	ensure.DeepEqual(
		t,
		(&singleStackError{
			underlying: err,
			config:     Config{},
			stack: stack.Stack{
				stack.Frame{
					File: "f",
					Line: 1,
					Name: "n",
				},
			},
		}).Error(),
		`foo`)
}

func TestMultiStackUnderlying(t *testing.T) {
	err := errors.New("")
	ensure.DeepEqual(t, (&multiStackError{underlying: err}).Underlying(), err)
}

func TestMultiStackStringModeAll(t *testing.T) {
	err := errors.New("foo")
	var m stack.Multi
	m.Add(stack.Stack{
		stack.Frame{
			File: "f",
			Line: 1,
			Name: "n",
		},
	})
	ensure.DeepEqual(
		t,
		(&multiStackError{
			underlying: err,
			config: Config{
				StringMode: StringModeAll,
			},
			multiStack: &m,
		}).Error(),
		"foo\nf:1 n")
}

func TestMultiStackStringModeNone(t *testing.T) {
	err := errors.New("foo")
	var m stack.Multi
	m.Add(stack.Stack{
		stack.Frame{
			File: "f",
			Line: 1,
			Name: "n",
		},
	})
	ensure.DeepEqual(
		t,
		(&multiStackError{
			underlying: err,
			config:     Config{},
			multiStack: &m,
		}).Error(),
		`foo`)
}

func TestWrapNil(t *testing.T) {
	ensure.Nil(t, Wrap(context.Background(), nil))
}

func TestWrapSingleFrameError(t *testing.T) {
	err := &singleFrameError{}
	ensure.DeepEqual(t, Wrap(context.Background(), err), err)
}

func TestWrapSingleStackErro(t *testing.T) {
	err := &singleStackError{}
	ensure.DeepEqual(t, Wrap(context.Background(), err), err)
}

func TestWrapExistingMultiStackError(t *testing.T) {
	err := &multiStackError{
		multiStack: new(stack.Multi),
	}
	we := Wrap(context.Background(), err)
	ensure.DeepEqual(t, we, err)
	ensure.DeepEqual(t, len(err.multiStack.Stacks()), 1)
}

func TestWrapNewSingleFrame(t *testing.T) {
	err := errors.New("")
	config := Config{StackMode: StackModeSingleFrame}
	ctx := WithConfig(context.Background(), config)
	we := Wrap(ctx, err).(*singleFrameError)
	ensure.DeepEqual(t, we.config, config)
	ensure.DeepEqual(t, we.underlying, err)
	ensure.DeepEqual(t, we.frame.Name, "TestWrapNewSingleFrame")
}

func TestWrapNewSingleStackError(t *testing.T) {
	err := errors.New("")
	config := Config{StackMode: StackModeSingleStack}
	ctx := WithConfig(context.Background(), config)
	we := Wrap(ctx, err).(*singleStackError)
	ensure.DeepEqual(t, we.config, config)
	ensure.DeepEqual(t, we.underlying, err)
	ensure.DeepEqual(t, we.stack[0].Name, "TestWrapNewSingleStackError")
}

func TestWrapNewMultiStackError(t *testing.T) {
	err := errors.New("")
	config := Config{StackMode: StackModeMultiStack}
	ctx := WithConfig(context.Background(), config)
	we := Wrap(ctx, err).(*multiStackError)
	ensure.DeepEqual(t, we.config, config)
	ensure.DeepEqual(t, we.underlying, err)
	ensure.DeepEqual(t, we.multiStack.Stacks()[0][0].Name, "TestWrapNewMultiStackError")
}

func TestWrapDefaultPassThru(t *testing.T) {
	err := errors.New("")
	ensure.DeepEqual(t, Wrap(context.Background(), err), err)
}

func TestStackFrameSingleFrame(t *testing.T) {
	frame := stack.Frame{Name: "42"}
	ensure.DeepEqual(t, StackFrame(&singleFrameError{frame: frame}), &frame)
}

func TestStackFrameSingleStack(t *testing.T) {
	frame := stack.Frame{Name: "42"}
	stack := stack.Stack{frame}
	ensure.DeepEqual(t, StackFrame(&singleStackError{stack: stack}), &frame)
}

func TestStackFrameMultiStack(t *testing.T) {
	frame := stack.Frame{Name: "42"}
	var multi stack.Multi
	multi.Add(stack.Stack{frame})
	ensure.DeepEqual(t, StackFrame(&multiStackError{multiStack: &multi}), &frame)
}

func TestStackFrameNotFound(t *testing.T) {
	ensure.True(t, StackFrame(errors.New("")) == nil)
}

func TestStackSingleFrame(t *testing.T) {
	frame := stack.Frame{Name: "42"}
	ensure.DeepEqual(t, Stack(&singleFrameError{frame: frame}), stack.Stack{frame})
}

func TestStackSingleStack(t *testing.T) {
	frame := stack.Frame{Name: "42"}
	stack := stack.Stack{frame}
	ensure.DeepEqual(t, Stack(&singleStackError{stack: stack}), stack)
}

func TestStackMultiStack(t *testing.T) {
	frame := stack.Frame{Name: "42"}
	s := stack.Stack{frame}
	var multi stack.Multi
	multi.Add(s)
	ensure.DeepEqual(t, Stack(&multiStackError{multiStack: &multi}), s)
}

func TestStackNotFound(t *testing.T) {
	ensure.True(t, Stack(errors.New("")) == nil)
}

func TestMultiStackSingleFrame(t *testing.T) {
	frame := stack.Frame{Name: "42"}
	var multi stack.Multi
	multi.Add(stack.Stack{frame})
	ensure.DeepEqual(t, MultiStack(&singleFrameError{frame: frame}), &multi)
}

func TestMultiStackSingleStack(t *testing.T) {
	frame := stack.Frame{Name: "42"}
	s := stack.Stack{frame}
	var multi stack.Multi
	multi.Add(s)
	ensure.DeepEqual(t, MultiStack(&singleStackError{stack: s}), &multi)
}

func TestMultiStackMultiStack(t *testing.T) {
	frame := stack.Frame{Name: "42"}
	var multi stack.Multi
	multi.Add(stack.Stack{frame})
	ensure.DeepEqual(t, MultiStack(&multiStackError{multiStack: &multi}), &multi)
}

func TestMultiStackNotFound(t *testing.T) {
	ensure.True(t, MultiStack(errors.New("")) == nil)
}

func TestRichError(t *testing.T) {
	err := errors.New("foo")
	ensure.DeepEqual(
		t,
		RichString(&singleFrameError{
			underlying: err,
			config: Config{
				StringMode: StringModeNone,
			},
			frame: stack.Frame{
				File: "f",
				Line: 1,
				Name: "n",
			},
		}),
		`f:1 n foo`)
}

func TestPoorError(t *testing.T) {
	err := errors.New("foo")
	ensure.DeepEqual(t, RichString(err), `foo`)
}
