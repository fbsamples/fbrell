package subcache_test

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/daaku/go.subcache"
)

type fixedCache struct {
	store func(key string, value []byte, timeout time.Duration) error
	get   func(key string) ([]byte, error)
}

func (c *fixedCache) Store(key string, value []byte, timeout time.Duration) error {
	return c.store(key, value, timeout)
}

func (c *fixedCache) Get(key string) ([]byte, error) {
	return c.get(key)
}

type proxyLogger struct {
	p func(v ...interface{})
}

func (l proxyLogger) Print(v ...interface{}) {
	l.p(v...)
}

func TestEmptyPrefix(t *testing.T) {
	t.Parallel()
	const errStr = "subcache: empty prefix"
	sc := &subcache.Client{}
	err := sc.Store("f", nil, time.Microsecond)
	if err.Error() != errStr {
		t.Fatalf(`expected "%s" got "%s"`, errStr, err)
	}
	_, err = sc.Get("f")
	if err.Error() != errStr {
		t.Fatalf(`expected "%s" got "%s"`, errStr, err)
	}
}

func TestGetNil(t *testing.T) {
	t.Parallel()
	prefix := "foo"
	plainKey := "bar"
	fullKey := prefix + ":" + plainKey
	c := &fixedCache{
		store: func(key string, value []byte, timeout time.Duration) error {
			return nil
		},
		get: func(key string) ([]byte, error) {
			if key != fullKey {
				t.Fatalf("expected key %s got %s", fullKey, key)
			}
			time.Sleep(2 * time.Millisecond)
			return nil, nil
		},
	}

	sc := &subcache.Client{
		Prefix:    prefix,
		ByteCache: c,
		Stats: func(s *subcache.Stats) {
			if s.Op != subcache.OpGet {
				t.Fatalf("expected op %s got %s", subcache.OpGet, s.Op)
			}
			if s.Key != fullKey {
				t.Fatalf("expected key %s got %s", fullKey, s.Key)
			}
			if s.Value != nil {
				t.Fatal("expected nil value")
			}
			if s.Duration.Nanoseconds() == 0 {
				t.Fatal("got zero duration value")
			}
			if s.Error != nil {
				t.Fatal(s.Error)
			}
		},
	}

	value, err := sc.Get(plainKey)
	if err != nil {
		t.Fatal(err)
	}
	if value != nil {
		t.Fatal("expected nil value")
	}
}

func TestStore(t *testing.T) {
	t.Parallel()
	prefix := "foo"
	plainKey := "bar"
	fullKey := prefix + ":" + plainKey
	actualVal := []byte("1")
	actualTimeout := time.Millisecond
	c := &fixedCache{
		store: func(key string, value []byte, timeout time.Duration) error {
			if key != fullKey {
				t.Fatalf("expected key %s got %s", fullKey, key)
			}
			if !bytes.Equal(value, actualVal) {
				t.Fatalf("expected value %s got %s", actualVal, value)
			}
			if timeout != actualTimeout {
				t.Fatalf("expected timeout %s got %s", actualTimeout, timeout)
			}
			time.Sleep(2 * time.Millisecond)
			return nil
		},
		get: func(key string) ([]byte, error) {
			return nil, nil
		},
	}

	sc := &subcache.Client{
		Prefix:    prefix,
		ByteCache: c,
		Stats: func(s *subcache.Stats) {
			if s.Op != subcache.OpStore {
				t.Fatalf("expected op %s got %s", subcache.OpStore, s.Op)
			}
			if s.Key != fullKey {
				t.Fatalf("expected key %s got %s", fullKey, s.Key)
			}
			if !bytes.Equal(s.Value, actualVal) {
				t.Fatalf("expected value %s got %s", actualVal, s.Value)
			}
			if s.Duration.Nanoseconds() == 0 {
				t.Fatal("got zero duration value")
			}
			if s.Error != nil {
				t.Fatal(s.Error)
			}
		},
	}

	err := sc.Store(plainKey, actualVal, actualTimeout)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDebugLoggerOnGetSuccess(t *testing.T) {
	t.Parallel()
	prefix := "foo"
	plainKey := "bar"
	expected := fmt.Sprintf(
		`%s subcache get for %s:%s took`,
		prefix,
		prefix,
		plainKey,
	)

	c := &fixedCache{
		get: func(key string) ([]byte, error) {
			return nil, nil
		},
	}
	l := &proxyLogger{
		p: func(v ...interface{}) {
			s := fmt.Sprint(v...)
			if !strings.HasPrefix(s, expected) {
				t.Fatalf(`did not find expected prefix "%s" got "%s"`, expected, s)
			}
		},
	}

	sc := &subcache.Client{
		Prefix:      prefix,
		ByteCache:   c,
		DebugLogger: l,
	}

	sc.Get(plainKey)
}

func TestDebugLoggerOnGetError(t *testing.T) {
	t.Parallel()
	prefix := "foo"
	plainKey := "bar"
	expected := fmt.Sprintf(
		`%s subcache get for %s:%s took`,
		prefix,
		prefix,
		plainKey,
	)
	const errStr = "the error"

	c := &fixedCache{
		get: func(key string) ([]byte, error) {
			return nil, errors.New(errStr)
		},
	}
	l := &proxyLogger{
		p: func(v ...interface{}) {
			s := fmt.Sprint(v...)
			if !strings.HasPrefix(s, expected) {
				t.Fatalf(`did not find expected prefix "%s" got "%s"`, expected, s)
			}
			if !strings.Contains(s, errStr) {
				t.Fatalf(`did not find expected error "%s" got "%s"`, errStr, s)
			}
		},
	}

	sc := &subcache.Client{
		Prefix:      prefix,
		ByteCache:   c,
		DebugLogger: l,
	}

	sc.Get(plainKey)
}

func TestErrorLoggerOnGetError(t *testing.T) {
	t.Parallel()
	prefix := "foo"
	plainKey := "bar"
	expected := fmt.Sprintf(
		`%s subcache get for %s:%s took`,
		prefix,
		prefix,
		plainKey,
	)
	const errStr = "the error"

	c := &fixedCache{
		get: func(key string) ([]byte, error) {
			return nil, errors.New(errStr)
		},
	}
	dl := &proxyLogger{
		p: func(v ...interface{}) {
			t.Fatal("was not expecting debug logger to be called")
		},
	}
	el := &proxyLogger{
		p: func(v ...interface{}) {
			s := fmt.Sprint(v...)
			if !strings.HasPrefix(s, expected) {
				t.Fatalf(`did not find expected prefix "%s" got "%s"`, expected, s)
			}
			if !strings.Contains(s, errStr) {
				t.Fatalf(`did not find expected error "%s" got "%s"`, errStr, s)
			}
		},
	}

	sc := &subcache.Client{
		Prefix:      prefix,
		ByteCache:   c,
		DebugLogger: dl,
		ErrorLogger: el,
	}

	sc.Get(plainKey)
}
