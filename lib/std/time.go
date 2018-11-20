package std

import (
	"time"

	"github.com/mz-eco/lua"
)

func init() {
	module.Define(&lua.Type{
		UUID: "7ba15f02-09c1-4b8d-9c34-7dcd29599dd4",
		Name: "Time",
		Type: lua.GoValue((*time.Time)(nil)),
	})
}
