package std

import "github.com/mz-eco/lua"

type Members struct {
	lua.ModuleMembers
}

var (
	module = &lua.Module{
		Name:    "std",
		Members: &Members{},
	}
)

func Loader() *lua.Module {
	return module
}
