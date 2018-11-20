package lua

import (
	"crypto/md5"
	"errors"
	"fmt"
	R "reflect"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

type Type struct {
	UUID string
	Name string
	Type R.Type
	name string
}

func (m *Type) checkValue(vm *VM) (R.Value, error) {

	ud := vm.CheckUserData(1)

	if ud == nil {
		return R.Value{}, errors.New(
			fmt.Sprintf(
				"load [%s:%s] from vm check lua type error", m.Name, m.UUID))
	}

	var (
		ov = R.ValueOf(ud.Value)
		ot = ov.Type()
	)

	if ot != m.Type {
		return R.Value{}, errors.New(
			fmt.Sprintf(
				"load [%s:%s] from argument error, want <%s> given <%s>",
				m.Name, m.UUID, m.Type, ot,
			))
	}

	return ov, nil
}

func (m *Type) GetName() string {
	return strings.ToUpper(
		fmt.Sprintf("GoMeta%x", md5.Sum([]byte(m.UUID))))
}

type TypeLoader func() *Type

func setter(vm *VM, x *Type) GFunction {
	i := &Invoker{
		Name: "__newindex",
		GoFunc: func(c *Call) int {

			var (
				o, err = x.checkValue(c.vm)
			)

			if err != nil {
				c.ArgError(1, err.Error())
				return 0
			}

			if len(c.Args) != 2 {
				c.ArgError(1, "setter argument error.")
				return 0
			}

			v := c.Args[0]

			name := lua.LVAsString(v)

			if len(name) == 0 {
				c.ArgError(1, "element not found.")
				return 0
			}

			field := o.Elem().FieldByName(name)

			if !field.IsValid() {
				c.ArgError(1, "element %s not found.", name)
				return 0
			}

			arg := c.Args[1]
			opts := &asOptions{
				vm:       c.vm,
				field:    name,
				skipFunc: true,
				base:     Nil,
				optional: false,
				raise:    false,
			}

			if goValue(c.vm, field, arg, opts) && opts.Ok() {
				return 0
			} else {
				c.ArgError(
					2,
					"%s",
					errTypeConvert{name, arg.Type(), field.Type()},
				)
				return 0
			}

			return 0
		},
	}

	return VMGFunction(i)
}

func getter(vm *VM, x *Type) GFunction {

	members := memberFunctions(x.Type, func(v R.Value, m int, i *Invoker) {

		i.Caller = func(vm *VM) (R.Value, error) {

			var (
				ov, err = x.checkValue(vm)
			)

			if err != nil {
				return R.Value{}, err
			}

			return ov.MethodByName(i.Name), nil
		}

	})

	i := &Invoker{
		Name: "__index",
		GoFunc: func(c *Call) int {

			var (
				o, err = x.checkValue(c.vm)
			)

			if err != nil {
				c.ArgError(1, err.Error())
				return 0
			}

			if len(c.Args) != 1 {
				c.ArgError(1, "setter argument error.")
				return 0
			}

			v := c.Args[0]

			name := lua.LVAsString(v)

			if len(name) == 0 {
				c.ArgError(1, "element not found.")
				return 0
			}

			field := o.Elem().FieldByName(name)

			if !field.IsValid() {

				member, ok := members[name]

				if !ok {
					c.ArgError(1, "element %s not found.", name)
					return 0
				}
				return member(c.vm)
			}

			c.vm.Push(luaValue(c.vm, field))

			return 1
		},
	}

	return VMGFunction(i)

}

func Define(vm *VM, x *Type) {

	x.name = x.GetName()
	tbl := vm.NewTypeMetatable(x.name)

	vm.SetGlobal(x.name, tbl)

	vm.SetField(
		tbl,
		"__newindex",
		vm.NewFunction(
			setter(vm, x)))

	vm.SetField(
		tbl,
		"__index",
		vm.NewFunction(
			getter(vm, x)),
	)

	typeAttach(vm, x)
}
