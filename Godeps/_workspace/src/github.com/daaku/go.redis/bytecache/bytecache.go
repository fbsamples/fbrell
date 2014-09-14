// Package bytecache provides a redis backed bytecache.
package bytecache

import (
	"github.com/daaku/go.redis"
	"time"
)

// Provides a redis backed Cache.
type Cache struct {
	client *redis.Client
}

// Create a new Cache instance with the given client.
func New(client *redis.Client) *Cache {
	return &Cache{client}
}

// Store a value with the given timeout.
func (c *Cache) Store(key string, value []byte, timeout time.Duration) error {
	args := []interface{}{"SET", key, value}
	if timeout != 0 {
		args = append(args, "PX", uint64(timeout.Nanoseconds()/int64(time.Millisecond)))
	}
	_, err := c.client.Call(args...)
	return err
}

// Get a stored value. A missing value will return nil, nil.
func (c *Cache) Get(key string) ([]byte, error) {
	item, err := c.client.Call("GET", key)
	if err != nil {
		return nil, err
	}
	if !item.Nil() {
		return item.Elem.Bytes(), nil
	}
	return nil, nil
}
