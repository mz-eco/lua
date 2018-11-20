package lua

import (
	"crypto/md5"
	"errors"
	"fmt"
	"strings"

	"github.com/yuin/gopher-lua"

	R "reflect"
)

var (
	typeNil = R.TypeOf(nil)
)

func GoType(x interface{}) R.Type {

	var (
		t = R.TypeOf(x)
	)

	if t.Kind() != R.Ptr {
		panic("go type input must bu a point")
	}

	if !t.Implements(typeClass) {
		panic(
			fmt.Sprintf("lua type [%s] must bu implement lua.Typed", t))
	}

	return R.TypeOf(x)
}

type class interface {
	setValue(value Value)
	getValue() Value
}

var (
	typeClass = R.TypeOf((*class)(nil)).Elem()
)

type Typed struct {
	value Value
}

func (m *Typed) setValue(v Value) {
	m.value = v
}

func (m *Typed) getValue() Value {
	return m.value
}

func (m *Typed) class() {

}

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

type golangTypes struct {
	types map[R.Type]*Type
}

func (m *golangTypes) Define(x *Type) {

	_, ok := m.types[x.Type]

	if ok {
		panic(
			fmt.Sprintf("type [%s:%s] already exists", x.Type.Name(), x.UUID))
	}

	m.types[x.Type] = x
}

func (m *golangTypes) Lookup(x R.Type) (t *Type, ok bool) {

	t, ok = m.types[x]

	return t, ok

}

func vmTypeLookup(vm *VM, t R.Type) (*Type, bool) {
	return vmTypes(vm).Lookup(t)
}

func vmTypeDefine(vm *VM, t *Type) {
	vmTypes(vm).Define(t)
}

func vmTypes(vm *VM) (types *golangTypes) {

	var (
		name = "GoTypes8087a566454b4d77a83d96b58dba5980"
	)

	lv := vm.GetGlobal(name)

	if lv.Type() == lua.LTNil {
		types = &golangTypes{
			types: make(map[R.Type]*Type),
		}

		ud := vm.NewUserData()
		ud.Value = types
		vm.SetGlobal(name, ud)

		return
	} else {
		ud, ok := lv.(*lua.LUserData)

		if !ok {
			panic("lua go types table is not a user data")
		}

		types, ok = ud.Value.(*golangTypes)

		if !ok {
			panic("lua go type value must *golangTypes")
		}

		return
	}
}

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

	var (
		name = x.GetName()
		tbl  = vm.NewTypeMetatable(name)
	)

	x.name = name
	vm.SetGlobal(name, tbl)

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

	vmTypeDefine(vm, x)
}
