// Package redis provides access to the rell redis instance.
package redis

import (
	"flag"
	"github.com/nshah/Go-Redis"
	"log"
)

var (
	redisHost = flag.String(
		"rell.redis.host", redis.DefaultRedisHost, "Redist host.")
	redisPort = flag.Int(
		"rell.redis.port", redis.DefaultRedisPort, "Redist port.")
	redisMemo redis.Client
)

// Get the shared Memcache client instance.
func Client() redis.Client {
	if redisMemo == nil {
		spec := redis.DefaultSpec().Host(*redisHost).Port(*redisPort)
		var err error
		redisMemo, err = redis.NewSynchClientWithSpec(spec)
		if err != nil {
			log.Fatalf("Failed to create redis client: %s", err)
		}
	}
	return redisMemo
}
