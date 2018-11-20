package types

import (
	"time"

	"github.com/mz-eco/lua"
)

type TimeMembers struct {
	lua.TypeMembers
}

func (m *TimeMembers) Now(cc int) time.Time {
	return time.Now()
}

func Load(vm *lua.VM) {

	lua.Define(vm, &lua.Type{
		UUID:    "7ba15f02-09c1-4b8d-9c34-7dcd29599dd4",
		Name:    "Time",
		Type:    lua.GoValue((*time.Time)(nil)),
		Members: &TimeMembers{},
	})
}
