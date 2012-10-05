// Package redis provides access to the rell redis instance.
package redis

import (
	"github.com/daaku/go.redis"
)

var (
	Client    = redis.ClientFlag("rell.redis")
	ByteCache = &bytecache{}
)

type bytecache struct{}

func (c *bytecache) Store(key string, value []byte) error {
	_, err := Client.Call("SET", key, value)
	return err
}

func (c *bytecache) Get(key string) ([]byte, error) {
	item, err := Client.Call("GET", key)
	if err != nil {
		return nil, err
	} else if !item.Nil() {
		return item.Elem.Bytes(), nil
	}
	return nil, nil
}
