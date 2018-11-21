package lua

import (
	"fmt"
	R "reflect"
)

type Call struct {
	vm   *VM
	name string
	Args []Value
}

func (m *Call) ArgError(n int, format string, v ...interface{}) {
	m.vm.ArgError(n, fmt.Sprintf(format, v...))
}

func (m *Call) Push(values ...Value) int {

	for _, value := range values {
		m.vm.Push(value)
	}

	return len(values)
}

func (m *Call) oValues(src []R.Value) int {

	var (
		dec    = NewEncoder(m.vm, true)
		values = make([]Value, 0)
	)

	for index := 0; index < len(src); index++ {

		value, err := dec.Encode(src[index])

		if err != nil {
			m.vm.RaiseError(
				"return #%d %s", index, err)
			return 0
		}

		values = append(values, value)
	}

	return m.Push(values...)
}

func (m *Call) Return(args ...interface{}) int {

	var (
		dec    = NewEncoder(m.vm, true)
		values = make([]Value, 0)
	)

	for index, arg := range args {
		value, err := dec.Encode(arg)

		if err != nil {
			m.vm.RaiseError(
				"return #%d %s", index, err)
			return 0
		}

		values = append(values, value)
	}

	return m.Push(values...)
}

func newCall(vm *VM, name string) *Call {

	c := &Call{
		vm:   vm,
		name: name,
		Args: make([]Value, 0),
	}

	for index := 1; index < vm.GetTop(); index++ {
		c.Args = append(c.Args, vm.Get(index+1))
	}

	return c

}
