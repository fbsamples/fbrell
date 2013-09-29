// Package collector provides rell specific stats collection routines.
package collector

import (
	"fmt"
	"log"

	"github.com/daaku/go.stats"
	"github.com/daaku/go.subcache"
)

type Collector struct {
	Logger *log.Logger
	Stats  stats.Backend
}

func (c *Collector) SubCacheStats(s *subcache.Stats) {
	var message string
	switch s.Op {
	default:
		c.Logger.Printf("unknown subcache.Stats.Op %s", s.Op)
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
	c.Stats.Count(message, 1)
	c.Stats.Record(message+" time", float64(s.Duration.Nanoseconds()))
}
