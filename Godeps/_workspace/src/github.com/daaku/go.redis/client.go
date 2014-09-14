// Package redis implements a client for Redis.
package redis

import (
	"errors"
	"strings"
	"time"
)

var errPoolSizeNotSpecified = errors.New("redis client pool size not specified")

type Stats interface {
	Inc(name string)
	Record(name string, value float64)
}

// Client implements a Redis connection which is what you should typically use
// instead of the lower level Conn interface. It implements a fixed size
// connection pool and supports per-call timeout.
type Client struct {
	Addr     string        // Server Address like "127.0.0.1:6379" or "/run/redis.sock"
	Proto    string        // Server Protocol like "tcp" or "unix"
	PoolSize uint          // Connection Pool Size, must be specified
	Timeout  time.Duration // Timeout per call
	Stats    Stats         // For Stats collection
	pool     chan Conn
}

func (c *Client) inc(name string) {
	if c.Stats != nil {
		c.Stats.Inc(name)
	}
}

func (c *Client) record(name string, value float64) {
	if c.Stats != nil {
		c.Stats.Record(name, value)
	}
}

// Call is the canonical way of talking to Redis. It accepts any
// Redis command and a arbitrary number of arguments.
func (c *Client) Call(args ...interface{}) (reply *Reply, err error) {
	start := time.Now()
	conn, err := c.connect()
	c.record(
		"redis connection acquire", float64(time.Since(start).Nanoseconds()))
	defer func() {
		c.record(
			"redis connection release", float64(time.Since(start).Nanoseconds()))
		if err != nil && c.shouldClose(err) {
			c.inc("redis connection error close")
			conn.Close()
			conn = nil
		}
		c.pool <- conn
	}()
	if err != nil {
		c.inc("redis connection accquire error")
		return nil, err
	}
	err = conn.Sock().SetDeadline(start.Add(c.Timeout))
	if err != nil {
		c.inc("redis connection set deadline error")
		return nil, err
	}
	err = conn.Write(args...)
	c.record("redis connection write", float64(time.Since(start).Nanoseconds()))
	if err != nil {
		c.inc("redis connection write error")
		return nil, err
	}
	reply, err = conn.Read()
	c.record("redis connection read", float64(time.Since(start).Nanoseconds()))
	if err != nil {
		c.inc("redis connection read error")
	}
	return reply, err
}

// Pop a connection from the pool or create a fresh one.
func (c *Client) connect() (conn Conn, err error) {
	if c.pool == nil {
		if c.PoolSize == 0 {
			return nil, errPoolSizeNotSpecified
		}
		c.pool = make(chan Conn, c.PoolSize)
		go func() {
			var i uint
			for i = 0; i < c.PoolSize; i++ {
				c.pool <- nil
			}
		}()
	}
	conn = <-c.pool
	if conn == nil {
		c.inc("redis connection new")
		conn, err = Dial(c.Addr, c.Proto, c.Timeout)
		if err != nil {
			return nil, err
		}
	}
	return conn, err
}

// Check if an error deserves closing the connection.
func (c *Client) shouldClose(err error) bool {
	if strings.HasSuffix(err.Error(), "broken pipe") {
		return true
	}
	return false
}
