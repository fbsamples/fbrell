package bytestore_test

import (
	"bytes"
	"testing"

	"github.com/daaku/go.redis/bytestore"
	"github.com/daaku/go.redis/redistest"
)

func TestSetGet(t *testing.T) {
	server, client := redistest.NewServerClient(t)
	defer server.Close()
	store := bytestore.New(client)
	const key = "key"
	expected := []byte("data")
	if err := store.Store(key, expected); err != nil {
		t.Fatal(err)
	}
	actual, err := store.Get(key)
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
	store := bytestore.New(client)
	const key = "key"
	actual, err := store.Get(key)
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
	store := bytestore.New(client)
	const key = "key"
	_, err := store.Get(key)
	if err == nil {
		t.Fatal("was expecting error")
	}
}
