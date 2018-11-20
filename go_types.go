package lua

import (
	"fmt"
	R "reflect"

	lua "github.com/yuin/gopher-lua"
)

type goTypes struct {
	types map[R.Type]*Type
}

func (m *goTypes) Define(x *Type) {

	_, ok := m.types[x.Type]

	if ok {
		panic(
			fmt.Sprintf("type [%s:%s] already exists", x.Type.Name(), x.UUID))
	}

	m.types[x.Type] = x
}

func (m *goTypes) Lookup(x R.Type) (t *Type, ok bool) {

	t, ok = m.types[x]

	return t, ok

}

func loadTypes(vm *VM) (types *goTypes) {

	var (
		name = "GoTypes8087a566454b4d77a83d96b58dba5980"
	)

	lv := vm.GetGlobal(name)

	if lv.Type() == lua.LTNil {
		types = &goTypes{
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

		types, ok = ud.Value.(*goTypes)

		if !ok {
			panic("lua go type value must *golangTypes")
		}

		return
	}
}

func typeLookup(vm *VM, t R.Type) (*Type, bool) {
	return loadTypes(vm).Lookup(t)
}

func typeAttach(vm *VM, t *Type) {
	loadTypes(vm).Define(t)
}
