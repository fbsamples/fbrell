package redis_test

import (
	"bytes"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/daaku/go.redis/redistest"
)

func error_(t *testing.T, name string, expected, got interface{}, err error) {
	if err != nil {
		t.Errorf("`%s` expected `%v` got `%v`, err(%v)", name, expected, got, err.Error())
	} else {
		t.Errorf("`%s` expected `%v` got `%v`, err(%v)", name, expected, got, err)
	}
}

func TestClient(t *testing.T) {
	server, client := redistest.NewServerClient(t)
	defer server.Close()
	if _, err := client.Call("SET", "foo", "foo"); err != nil {
		t.Fatal(err.Error())
	}
}

func TestTimeout(t *testing.T) {
	server, client := redistest.NewServerClient(t)
	defer server.Close()
	client.Timeout = time.Nanosecond
	_, err := client.Call("SET", "foo", "foo")
	if err == nil {
		t.Fatal("was expecting timeout error but got no error. redis too fast?")
	}
	if !strings.HasSuffix(err.Error(), "i/o timeout") {
		t.Fatalf("was expecting timeout error but got: %s", err)
	}
}

func BenchmarkItoa(b *testing.B) {
	for i := 0; i < b.N; i++ {
		strconv.Itoa(i)
	}
}

func BenchmarkSet(b *testing.B) {
	b.StopTimer()
	server, client := redistest.NewServerClient(b)
	b.StartTimer()
	defer func() {
		b.StopTimer()
		server.Close()
		b.StartTimer()
	}()
	for i := 0; i < b.N; i++ {
		client.Call("SET", "foo", "foo")
	}
}

func BenchmarkAppendUint(b *testing.B) {
	var buf []byte
	buf = make([]byte, 0, 1024*16)

	for i := 0; i < b.N; i++ {
		strconv.AppendUint(buf, uint64(i), 10)
	}
}

func BenchmarkAppendBytes(b *testing.B) {
	var buf []byte
	buf = make([]byte, 0, 1024*16)

	for i := 0; i < b.N; i++ {
		buf = append(buf, '\r')
	}
}

func BenchmarkAppendBuffer(b *testing.B) {
	buf := bytes.NewBuffer(make([]byte, 1024*16))

	for i := 0; i < b.N; i++ {
		buf.WriteByte('\r')
	}
}
