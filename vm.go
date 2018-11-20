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

func VMFunction(vm *VM, name string, v interface{}) error {

	return As(
		vm,
		vm.GetGlobal(name),
		v,
	)
}

func VMValue(vm *VM, name string, v interface{}) error {

	return As(
		vm,
		vm.GetGlobal(name),
		v)
}
