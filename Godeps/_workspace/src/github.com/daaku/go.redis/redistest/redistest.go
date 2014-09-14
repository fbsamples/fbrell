// Package redistest provides test redis server support. It provides a real
// in-memory redis server.
package redistest

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/ParsePlatform/go.freeport"
	"github.com/daaku/go.redis"
)

type Server struct {
	Command *exec.Cmd
	Port    int
	T       Fatalf
}

type Fatalf interface {
	Fatalf(format string, args ...interface{})
}

func NewServer(t Fatalf) *Server {
	port, err := freeport.Get()
	if err != nil {
		t.Fatalf("failed to find freeport: %s", err)
	}
	s := &Server{
		Port: port,
		T:    t,
	}
	err = s.Start()
	if err != nil {
		t.Fatalf("failed to find freeport: %s", err)
	}
	return s
}

func NewServerClient(t Fatalf) (*Server, *redis.Client) {
	server := NewServer(t)
	client := &redis.Client{
		Proto:    server.Proto(),
		Addr:     server.Addr(),
		PoolSize: 10,
		Timeout:  time.Millisecond * 100,
	}
	for {
		_, err := client.Call("PING")
		if err == nil {
			break
		}
		if strings.HasSuffix(err.Error(), "connection refused") {
			continue
		}
		t.Fatalf("err %s", err)
	}
	return server, client
}

func (s *Server) Proto() string {
	return "tcp"
}

func (s *Server) Addr() string {
	return fmt.Sprintf("127.0.0.1:%d", s.Port)
}

func (s *Server) Start() error {
	s.Command = exec.Command("redis-server", "-")
	in, err := s.Command.StdinPipe()
	if err != nil {
		return err
	}
	stderr, err := s.Command.StderrPipe()
	if err != nil {
		return err
	}
	go io.Copy(os.Stderr, stderr)
	_, err = fmt.Fprintf(in, "port %d", s.Port)
	if err != nil {
		return err
	}
	in.Close()
	err = s.Command.Start()
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) Close() error {
	return s.Command.Process.Kill()
}
