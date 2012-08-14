// Package redis provides access to the rell redis instance.
package redis

import (
	"github.com/daaku/go.redis"
)

var Client = redis.ClientFlag("rell.redis")
