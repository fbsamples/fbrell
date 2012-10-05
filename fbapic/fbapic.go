// Provides cached FB API calls.
package fbapic

import (
	"encoding/json"
	"fmt"
	"github.com/daaku/go.fbapi"
	"github.com/daaku/go.stats"
	"log"
	"net/url"
	"time"
)

type Backend interface {
	Store(key string, value []byte) error
	Get(key string) ([]byte, error)
}

// Configure a Cached API accessor instance. You'll typically define
// one per type of cached call. An instance can be shared across
// goroutines.
type Cache struct {
	Backend Backend       // provides the storage implementation
	Prefix  string        // cache key prefix
	Timeout time.Duration // per value timeout
}

// Make a GET Graph API request.
func (c *Cache) Get(result interface{}, path string, values ...fbapi.Values) error {
	key := fmt.Sprintf("%s:%s", c.Prefix, path)
	raw, err := c.Backend.Get(key)
	if err != nil {
		stats.Inc("fbapic backend.Get error")
		stats.Inc("fbapic backend.Get error " + c.Prefix)
		return fmt.Errorf("fbapic error in backend.Get: %s", err)
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
		taken := float64(time.Since(start).Nanoseconds())
		stats.Record("fbapic graph api time", taken)
		stats.Record("fbapic graph api time "+c.Prefix, taken)

		err = c.Backend.Store(key, raw)
		if err != nil {
			log.Printf("fbapic error in cache.Set: %s", err)
		}
	} else {
		stats.Inc("fbapic cache hit")
		stats.Inc("fbapic cache hit " + c.Prefix)
	}

	err = json.Unmarshal(raw, result)
	if err != nil {
		return fmt.Errorf(
			"Request for path %s with response %s failed with "+
				"json.Unmarshal error %s.", path, string(raw), err)
	}
	return nil
}
