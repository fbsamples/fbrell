package redis

import (
	"github.com/daaku/go.redis/bufin"
	"net"
	"time"
)

// Represents a Connection to the server and abstracts the read/write
// via a connection.
type Conn interface {
	// Write accepts any redis command and arbitrary list of arguments.
	//
	//     Write("SET", "counter", 1)
	//     Write("INCR", "counter")
	//
	// Write might return a net.Conn.Write error
	Write(args ...interface{}) error

	// Read a single reply from the connection. If there is no reply waiting
	// this method will block.
	Read() (*Reply, error)

	// Close the Connection.
	Close() error

	// Returns the underlying net.Conn. This is useful for example to set
	// set a r/w deadline on the connection.
	//
	//      conn.Sock().SetDeadline(t)
	Sock() net.Conn
}

type connection struct {
	rbuf *bufin.Reader
	conn net.Conn
}

// Dial expects a network address, protocol and a dial timeout:
//
//     Dial("127.0.0.1:6379", "tcp", time.Second)
//
// Or for a unix domain socket:
//
//     Dial("/path/to/redis.sock", "unix", time.Second)
func Dial(addr, proto string, timeout time.Duration) (Conn, error) {
	conn, err := net.DialTimeout(proto, addr, timeout)
	if err != nil {
		return nil, err
	}
	c := &connection{bufin.NewReader(conn), conn}
	return c, nil
}

func (c *connection) Read() (*Reply, error) {
	reply := parse(c.rbuf)
	if reply.Err != nil {
		return nil, reply.Err
	}
	return reply, nil
}

func (c *connection) Write(args ...interface{}) error {
	_, err := c.conn.Write(format(args...))
	if err != nil {
		return err
	}
	return nil
}

func (c *connection) Close() error {
	return c.conn.Close()
}

func (c *connection) Sock() net.Conn {
	return c.conn
}
