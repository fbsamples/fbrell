// Provides cached FB API calls.
package fbapic

import (
	"encoding/json"
	"fmt"
	"github.com/daaku/go.fbapi"
	"github.com/daaku/go.stats"
	"github.com/daaku/rell/redis"
	"log"
	"net/url"
	"time"
)

// Configure a Cached API accessor instance. You'll typically define
// one per type of cached call. An instance can be shared across
// goroutines.
type Cache struct {
	Prefix  string        // cache key prefix
	Timeout time.Duration // per value timeout
}

// Make a GET Graph API request.
func (c *Cache) Get(result interface{}, path string, values ...fbapi.Values) error {
	var err error
	var raw []byte

	key := fmt.Sprintf("%s:%s", c.Prefix, path)
	item, err := redis.Client.Call("GET", key)
	if err != nil {
		log.Printf("fbapic error in redis.Get: %s", err)
	} else if !item.Nil() {
		raw = item.Elem.Bytes()
		stats.Inc("fbapic cache hit")
		stats.Inc("fbapic cache hit " + c.Prefix)
	}

	if raw == nil {
		stats.Inc("fbapic cache miss")
		stats.Inc("fbapic cache miss " + c.Prefix)
		final := url.Values{}
		for _, v := range values {
			v.Set(final)
		}
		start := time.Now()
		raw, err = fbapi.GetRaw(path, final)
		if err != nil {
			stats.Inc("fbapic graph api error")
			stats.Inc("fbapic graph api error " + c.Prefix)
			return err
		}
		stats.Record("fbapic graph api time", float64(time.Since(start).Nanoseconds()))
		stats.Record("fbapic graph api time "+c.Prefix, float64(time.Since(start).Nanoseconds()))
	}

	err = json.Unmarshal(raw, result)
	if err != nil {
		return fmt.Errorf(
			"Request for path %s with response %s failed with "+
				"json.Unmarshal error %s.", path, string(raw), err)
	}
	_, err = redis.Client.Call("SET", key, raw)
	if err != nil {
		log.Printf("fbapic error in cache.Set: %s", err)
	}
	return nil
}
