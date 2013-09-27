package empcheck

import (
	"bytes"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/daaku/go.fbapi"
	"github.com/daaku/go.fbapp"
)

var (
	fields = fbapi.ParamFields("is_employee")
	yes    = []byte("1")
	no     = []byte("0")
)

type user struct {
	IsEmployee bool `json:"is_employee"`
}

type Logger interface {
	Printf(format string, v ...interface{})
}

type Cache interface {
	Get(key string) ([]byte, error)
	Store(key string, val []byte, timeout time.Duration) error
}

type Checker struct {
	FbApiClient  *fbapi.Client
	App          fbapp.App
	Logger       Logger
	Cache        Cache
	CacheTimeout time.Duration
}

// Check if the user is a Facebook Employee. This only available by
// special permission granted to an application by Facebook.
func (c *Checker) Check(id uint64) bool {
	ids := strconv.FormatUint(id, 10)

	is, _ := c.Cache.Get(ids)
	if is != nil {
		if bytes.Equal(is, yes) {
			return true
		}
		if bytes.Equal(is, no) {
			return false
		}
		c.Logger.Printf("invalid cached result for IsEmployee %s = %s", ids, is)
	}

	values, err := fbapi.ParamValues(c.App, fields)
	if err != nil {
		c.Logger.Printf("Ignoring error in IsEmployee ParamValues: %s", err)
		return false
	}

	var user user
	u := url.URL{Path: ids, RawQuery: values.Encode()}
	req := http.Request{Method: "GET", URL: &u}
	_, err = c.FbApiClient.Do(&req, &user)
	if err != nil {
		if apiErr, ok := err.(*fbapi.Error); ok {
			if apiErr.Code == 100 { // common error with test users
				return false
			}
		}
		c.Logger.Printf("Ignoring error in IsEmployee FbApiClient.Do: %s", err)
		return false
	}

	v := no
	if user.IsEmployee {
		v = yes
	}
	c.Cache.Store(ids, v, c.CacheTimeout)
	return user.IsEmployee
}
