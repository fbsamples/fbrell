// Package service provides configured service instances.
package service

import (
	"fmt"
	"log"
	"os"

	"github.com/daaku/go.fbapi"
	"github.com/daaku/go.httpcontrol"
	"github.com/daaku/go.redis"
	"github.com/daaku/go.redis/bytecache"
	"github.com/daaku/go.redis/bytestore"
	"github.com/daaku/go.static"
	"github.com/daaku/go.stats/stathat"
	"github.com/daaku/go.subcache"
	"github.com/daaku/go.xsrf"
)

var (
	Xsrf          = xsrf.ProviderFlag("xsrf")
	Static        = static.HandlerFlag("rell.static")
	Stats         = stathat.ClientFlag("rell.stats")
	Redis         = redis.ClientFlag("rell.redis")
	ByteCache     = bytecache.New(Redis)
	ByteStore     = bytestore.New(Redis)
	HttpTransport = httpcontrol.TransportFlag("rell.transport")
	FbApiClient   = fbapi.ClientFlag("rell.fbapi")
	Logger        = log.New(os.Stderr, "", log.LstdFlags)
)

func init() {
	Stats.Transport = HttpTransport
	FbApiClient.Transport = HttpTransport
	Redis.Stats = Stats
}

func SubCacheStats(s *subcache.Stats) {
	var message string
	switch s.Op {
	default:
		Logger.Printf("unknown subcache.Stats.Op %s", s.Op)
		return
	case subcache.OpGet:
		if s.Error == nil {
			if s.Value == nil {
				message = fmt.Sprintf("%s subcache miss", s.Client.Prefix)
			} else {
				message = fmt.Sprintf("%s subcache hit", s.Client.Prefix)
			}
		} else {
			message = fmt.Sprintf("%s subcache get error", s.Client.Prefix)
		}
	case subcache.OpStore:
		if s.Error == nil {
			message = fmt.Sprintf("%s subcache store", s.Client.Prefix)
		} else {
			message = fmt.Sprintf("%s subcache store error", s.Client.Prefix)
		}
	}
	Stats.Inc(message)
	Stats.Record(message+" time", float64(s.Duration.Nanoseconds()))
}
