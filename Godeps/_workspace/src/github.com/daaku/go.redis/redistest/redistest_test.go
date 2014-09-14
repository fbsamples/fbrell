package redistest_test

import (
	"testing"

	"github.com/daaku/go.redis/redistest"
)

func TestClient(t *testing.T) {
	server, client := redistest.NewServerClient(t)
	defer server.Close()
	if _, err := client.Call("SET", "foo", "foo"); err != nil {
		t.Fatal(err.Error())
	}
}
