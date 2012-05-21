package empcheck

import (
	"fmt"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/nshah/go.fbapi"
	"github.com/nshah/go.fbapp"
	"github.com/nshah/rell/cache"
	"log"
)

const (
	yes = byte('1')
	no  = byte('0')
)

var (
	app      = fbapp.Flag("empcheck")
	fields   = fbapi.Fields([]string{"is_employee"})
	yesSlice = []byte{yes}
	noSlice  = []byte{no}
)

type user struct {
	IsEmployee bool `json:"is_employee"`
}

// Check if the user is a Facebook Employee. This only available by
// special permission granted to an application by Facebook.
func IsEmployee(id uint64) bool {
	key := fmt.Sprintf("is_employee:%d", id)
	item, err := cache.Client().Get(key)
	if err != nil {
		if err != memcache.ErrCacheMiss {
			log.Printf("Unknown cache error: %s", err)
		}
	} else {
		if item.Value[0] == yes {
			return true
		} else if item.Value[0] == no {
			return false
		} else {
			log.Printf("Unknown cache key/value: %s=%s", key, item.Value)
		}
	}

	user := &user{}
	err = fbapi.Get(user, fmt.Sprintf("/%d", id), app, fields)
	if err != nil {
		log.Printf("Ignoring error in IsEmployee check: %s", err)
		return false
	}
	value := noSlice
	if user.IsEmployee {
		value = yesSlice
	}
	err = cache.Client().Set(&memcache.Item{Key: key, Value: value})
	if err != nil {
		log.Printf("Error in cache.Set: %s", err)
	}
	return user.IsEmployee
}
