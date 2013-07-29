package empcheck

import (
	"bytes"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/daaku/go.fbapi"
	"github.com/daaku/go.fbapp"
	"github.com/daaku/go.subcache"

	"github.com/daaku/rell/service"
)

var (
	app    = fbapp.Flag("empcheck")
	fields = fbapi.ParamFields("is_employee")
	cache  = &subcache.Client{
		Prefix:      "is_employee",
		ByteCache:   service.ByteCache,
		ErrorLogger: service.Logger,
		Stats:       service.SubCacheStats,
	}
	cacheTimeout = 24 * 90 * time.Hour
)

var (
	yes = []byte("1")
	no  = []byte("0")
)

type user struct {
	IsEmployee bool `json:"is_employee"`
}

// Check if the user is a Facebook Employee. This only available by
// special permission granted to an application by Facebook.
func IsEmployee(id uint64) bool {
	ids := strconv.FormatUint(id, 10)

	is, _ := cache.Get(ids)
	if is != nil {
		if bytes.Equal(is, yes) {
			return true
		}
		if bytes.Equal(is, no) {
			return false
		}
		service.Logger.Printf("invalid cached result for IsEmployee %s = %s", ids, is)
	}

	values, err := fbapi.ParamValues(app, fields)
	if err != nil {
		service.Logger.Printf("Ignoring error in IsEmployee ParamValues: %s", err)
		return false
	}

	var user user
	u := url.URL{Path: ids, RawQuery: values.Encode()}
	req := http.Request{Method: "GET", URL: &u}
	_, err = service.FbApiClient.Do(&req, &user)
	if err != nil {
		if apiErr, ok := err.(*fbapi.Error); ok {
			if apiErr.Code == 100 { // common error with test users
				return false
			}
		}
		service.Logger.Printf("Ignoring error in IsEmployee FbApiClient.Do: %s", err)
		return false
	}

	v := no
	if user.IsEmployee {
		v = yes
	}
	cache.Store(ids, v, cacheTimeout)
	return user.IsEmployee
}
