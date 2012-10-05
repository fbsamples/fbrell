// Package service provides configured service instances.
package service

import (
	"github.com/daaku/go.redis"
	"github.com/daaku/go.stats/stathatbackend"
	"time"
)

var (
	Redis     = redis.ClientFlag("rell.redis")
	Stats     = stathatbackend.EZKeyFlag("rell.stats")
	ByteCache = &bytecache{Redis}
)

type bytecache struct {
	Redis *redis.Client
}

func (c *bytecache) Store(key string, value []byte, timeout time.Duration) error {
	_, err := c.Redis.Call("SET", key, value)
	return err
}

func (c *bytecache) Get(key string) ([]byte, error) {
	item, err := c.Redis.Call("GET", key)
	if err != nil {
		return nil, err
	} else if !item.Nil() {
		return item.Elem.Bytes(), nil
	}
	return nil, nil
}
