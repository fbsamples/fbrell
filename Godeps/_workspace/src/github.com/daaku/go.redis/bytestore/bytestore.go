// Package bytestore provides a redis backed bytestore.
package bytestore

import (
	"github.com/daaku/go.redis"
)

// Provides a redis backed Store.
type Store struct {
	client *redis.Client
}

// Create a new Store instance with the given client.
func New(client *redis.Client) *Store {
	return &Store{client}
}

// Store a value.
func (c *Store) Store(key string, value []byte) error {
	_, err := c.client.Call("SET", key, value)
	return err
}

// Get a stored value. A missing value will return nil, nil.
func (c *Store) Get(key string) ([]byte, error) {
	item, err := c.client.Call("GET", key)
	if err != nil {
		return nil, err
	} else if !item.Nil() {
		return item.Elem.Bytes(), nil
	}
	return nil, nil
}
