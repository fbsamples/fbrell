package empcheck

import (
	"fmt"
	"log"

	"github.com/daaku/go.fbapi"
	"github.com/daaku/go.fbapi/fbapic"
	"github.com/daaku/go.fbapp"

	"github.com/daaku/rell/service"
)

var (
	app    = fbapp.Flag("empcheck")
	fields = fbapi.Fields([]string{"is_employee"})
	cached = &fbapic.Cache{
		Prefix:    "is_employee",
		ByteCache: service.ByteCache,
		Stats:     service.Stats,
		Client:    service.FbApiClient,
	}
)

type user struct {
	IsEmployee bool `json:"is_employee"`
}

// Check if the user is a Facebook Employee. This only available by
// special permission granted to an application by Facebook.
func IsEmployee(id uint64) bool {
	user := &user{}
	err := cached.Get(user, fmt.Sprintf("/%d", id), app, fields)
	if err != nil {
		if apiErr, ok := err.(*fbapi.Error); ok {
			if apiErr.Code == 100 { // common error with test users
				return false
			}
		}
		log.Printf("Ignoring error in IsEmployee check: %s", err)
	}
	return user.IsEmployee
}
