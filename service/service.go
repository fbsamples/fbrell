// Package service provides configured service instances.
package service

import (
	"github.com/daaku/go.redis"
	"github.com/daaku/go.redis/bytecache"
	"github.com/daaku/go.redis/bytestore"
	"github.com/daaku/go.static"
	"github.com/daaku/go.stats/stathatbackend"
)

var (
	Static    = static.HandlerFlag("rell.static")
	Stats     = stathatbackend.EZKeyFlag("rell.stats")
	Redis     = redis.ClientFlag("rell.redis")
	ByteCache = bytecache.New(Redis)
	ByteStore = bytestore.New(Redis)
)

func init() {
	Redis.Stats = Stats
}
