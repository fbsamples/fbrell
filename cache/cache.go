// Package cache provides a simple cache wrapper.
package cache

import (
	"flag"
	"github.com/bradfitz/gomemcache/memcache"
)

var (
	cacheServer = flag.String(
		"rell.memcache",
		"127.0.0.1:11211",
		"Memcache server for caching purposes.")
	cacheMemo *memcache.Client
)

// Get the shared Memcache client instance.
func Client() *memcache.Client {
	if cacheMemo == nil {
		cacheMemo = memcache.New(*cacheServer)
	}
	return cacheMemo
}
