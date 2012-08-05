// Package redis provides access to the rell redis instance.
package redis

import (
	"flag"
	"github.com/daaku/go.redis"
)

var memo = redis.NewClient("", 0, "", 50)

func init() {
	flag.StringVar(&memo.Addr, "rell.redis", "127.0.0.1:6379", "Redis addr.")
}

// Get the shared Memcache client instance.
func Client() *redis.Client {
	return memo
}
