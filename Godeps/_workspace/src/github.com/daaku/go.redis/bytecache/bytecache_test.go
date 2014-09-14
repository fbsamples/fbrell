package bytecache_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/daaku/go.redis/bytecache"
	"github.com/daaku/go.redis/redistest"
)

func TestSetGet(t *testing.T) {
	server, client := redistest.NewServerClient(t)
	defer server.Close()
	cache := bytecache.New(client)
	const key = "key"
	expected := []byte("data")
	if err := cache.Store(key, expected, time.Second); err != nil {
		t.Fatal(err)
	}
	actual, err := cache.Get(key)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(actual, expected) {
		t.Fatalf("found %s instead of %s", actual, expected)
	}
}

func TestGetMissing(t *testing.T) {
	server, client := redistest.NewServerClient(t)
	defer server.Close()
	cache := bytecache.New(client)
	const key = "key"
	actual, err := cache.Get(key)
	if err != nil {
		t.Fatal(err)
	}
	if actual != nil {
		t.Fatalf("found %s instead of nil", actual)
	}
}

func TestDeadClient(t *testing.T) {
	server, client := redistest.NewServerClient(t)
	server.Close()
	cache := bytecache.New(client)
	const key = "key"
	_, err := cache.Get(key)
	if err == nil {
		t.Fatal("was expecting error")
	}
}

func TestExpires(t *testing.T) {
	server, client := redistest.NewServerClient(t)
	defer server.Close()
	cache := bytecache.New(client)
	const key = "key"
	expected := []byte("data")
	if err := cache.Store(key, expected, time.Millisecond); err != nil {
		t.Fatal(err)
	}
	time.Sleep(3 * time.Millisecond)
	actual, err := cache.Get(key)
	if err != nil {
		t.Fatal(err)
	}
	if actual != nil {
		t.Fatalf("found %s instead of nil", actual)
	}
}
