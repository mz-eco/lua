package lua

import (
	"github.com/yuin/gopher-lua"
)

type LGFunction func(vm *VM) int

func New() *VM {
	return lua.NewState(lua.Options{
		IncludeGoStackTrace: true,
	})
}
