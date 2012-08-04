// Package redis provides access to the rell redis instance.
package redis

import (
	"flag"
	"fmt"
	"github.com/simonz05/godis"
)

var (
	redisHost = flag.String("rell.redis.host", "127.0.0.1", "Redist host.")
	redisPort = flag.Int("rell.redis.port", 6379, "Redist port.")
	redisMemo *godis.Client
)

// Get the shared Memcache client instance.
func Client() *godis.Client {
	if redisMemo == nil {
		redisMemo = godis.New(fmt.Sprintf("tcp:%s:%d", *redisHost, *redisPort), 0, "")
	}
	return redisMemo
}
