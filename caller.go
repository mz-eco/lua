package lua

import (
	"fmt"
	R "reflect"
)

type Call struct {
	vm   *VM
	Args []Value
}

func typedCaller(_ *Call) int { return 0 }

var (
	typeCall   = R.TypeOf((*Call)(nil))
	typeCaller = R.TypeOf(typedCaller)
)

func (m *Call) ArgError(n int, format string, v ...interface{}) {

	m.vm.ArgError(n, fmt.Sprintf(format, v...))
}

func newCall(vm *VM) *Call {

	c := &Call{
		vm:   vm,
		Args: make([]Value, 0),
	}

	for index := 1; index < vm.GetTop(); index++ {
		c.Args = append(c.Args, vm.Get(index+1))
	}

	return c

}
