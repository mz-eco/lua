package lua

import (
	"crypto/md5"
	"fmt"
	"strings"

	"github.com/yuin/gopher-lua"

	R "reflect"
)

func GoValue(x interface{}) R.Type {
	return R.TypeOf(x).Elem()
}

func GoType(x interface{}) R.Type {
	return R.TypeOf(x)
}

type TypedMembers interface {
	setVM(vm *VM)
}

type TypeMembers struct {
	VM *VM
}

func (m *TypeMembers) setVM(vm *VM) {
	m.VM = vm
}

type Type struct {
	UUID    string
	Name    string
	Type    R.Type
	Members TypedMembers
}

func (m *Type) GetName() string {

	if len(m.Name) != 0 {
		return m.Name
	}

	return strings.ToUpper(
		fmt.Sprintf("GoType%x", md5.Sum([]byte(m.UUID))))
}

type TypeLoader func() *Type

func memberCaller(t R.Type, mt R.Method, x *Type) GFunction {

	var (
		hasError = false
		no       = t.NumOut()
		ret      = no
	)

	if no > 0 {
		hasError = t.Out(no - 1).Implements(typeError)
		if hasError {
			ret = no - 1
		}

	}

	return func(vm *VM) int {

		fmt.Println(vm.GetTop(), vm.Get(0), vm.Get(1), vm.Get(2))
		ud := vm.CheckUserData(1)
		if ud == nil {
			return 0
		}

		if ud.Value == nil {
			vm.TypeError(1, lua.LTUserData)
			return 0
		}

		uv := R.ValueOf(ud.Value)
		ut := uv.Type()

		if ut != x.Type {
			vm.TypeError(1, lua.LTUserData)
			return 0
		}

		var (
			in   = make([]R.Value, 0)
			opts = &asOptions{
				vm:       vm,
				skipFunc: false,
				base:     Nil,
				optional: false,
				raise:    false,
			}
		)

		for index := 1; index < t.NumIn()-1; index++ {

			var (
				it = t.In(index)
				iv = R.New(it)
				lv = vm.Get(index)
			)

			if goValue(vm, iv.Elem(), lv, opts) && opts.Ok() {
				in = append(in, iv.Elem())
			} else {
				vm.ArgError(
					index,
					opts.GetError(
						errTypeConvert{
							"",
							lv.Type(),
							it,
						},
					).Error(),
				)

				return 0
			}

		}

		m := uv.MethodByName(mt.Name)
		o := m.Call(in)

		fmt.Println("xx", hasError)
		if hasError {
			err := o[ret]

			if !err.IsNil() {
				vm.RaiseError(
					fmt.Sprint(err.Interface()))

				return 0
			}
		}

		for index := 0; index < ret; index++ {
			vm.Push(luaValue(vm, o[index]))
		}

		return ret

	}
}

func staticCaller(m R.Value, x *Type) GFunction {

	var (
		t = m.Type()
	)

	return func(vm *VM) int {

		var (
			in   = make([]R.Value, 0)
			opts = &asOptions{
				vm:       vm,
				skipFunc: false,
				base:     Nil,
				optional: false,
				raise:    false,
			}
		)

		for index := 0; index < t.NumIn(); index++ {

			var (
				it = t.In(index)
				iv = R.New(it)
				lv = vm.Get(index + 1)
			)

			if goValue(vm, iv.Elem(), lv, opts) && opts.Ok() {
				in = append(in, iv.Elem())
			} else {
				vm.ArgError(
					index,
					opts.GetError(
						errTypeConvert{
							"",
							lv.Type(),
							it,
						},
					).Error(),
				)

				return 0
			}
		}

		o := m.Call(in)

		if len(o) != 1 {
			vm.RaiseError("call outer fail")
			return 0
		}

		u := vm.NewUserData()
		u.Value = o[0].Interface()

		vm.SetMetatable(u, vm.GetTypeMetatable(x.Name))
		vm.Push(u)

		return 1
	}
}

func members(vm *VM, tbl *Table, x *Type) {

	var (
		t   = x.Type
		n   = t.NumMethod()
		fns = make(map[string]GFunction)
	)

	for index := 0; index < n; index++ {
		var (
			mt = t.Method(index)
			ft = mt.Type
		)

		fns[mt.Name] = memberCaller(ft, mt, x)
	}

	vm.SetField(tbl, "__index",
		vm.SetFuncs(vm.NewTable(), fns),
	)

}

func static(vm *VM, tbl *Table, v R.Value, x *Type) {

	var (
		t = v.Type()
		n = t.NumMethod()
	)

	for index := 0; index < n; index++ {
		var (
			mt = t.Method(index)
			mv = v.Method(index)
			ft = mt.Type
			ol = ft.NumOut()
		)

		if ol != 1 || ft.Out(index) != x.Type {
			panic(
				fmt.Sprintf("type static function must only return type %s", x.Type))
		}

		vm.SetField(
			tbl,
			mt.Name,
			vm.NewFunction(
				staticCaller(mv, x)))

	}

}

type golangTypes struct {
	types map[R.Type]*Type
}

func (m *golangTypes) Define(x *Type) {

	v, ok := m.types[x.Type]

	if !ok {
		panic(
			fmt.Sprintf("type [%s:%s] already exists", x.Type.Name(), x.UUID))
	}
	m.types[x.Type] = v
}

func (m *golangTypes) Lookup(x R.Type) (t *Type, ok bool) {

	t, ok = m.types[x]

	return

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

func Define(vm *VM, x *Type) {

	var (
		name = x.GetName()
		tbl  = vm.NewTypeMetatable(name)
	)

	vm.SetGlobal(name, tbl)

	if x.Members != nil {
		x.Members.setVM(vm)

		static(
			vm,
			tbl,
			R.ValueOf(x.Members),
			x)
	}

	members(vm, tbl, x)
}
