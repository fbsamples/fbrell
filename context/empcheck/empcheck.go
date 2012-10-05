package empcheck

import (
	"fmt"
	"github.com/daaku/go.fbapi"
	"github.com/daaku/go.fbapp"
	"github.com/daaku/rell/fbapic"
	"github.com/daaku/rell/service"
	"log"
)

var (
	app    = fbapp.Flag("empcheck")
	fields = fbapi.Fields([]string{"is_employee"})
	cached = &fbapic.Cache{
		Prefix:    "is_employee",
		ByteCache: service.ByteCache,
		Stats:     service.Stats,
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
		log.Printf("Ignoring error in IsEmployee check: %s", err)
	}
	return user.IsEmployee
}
