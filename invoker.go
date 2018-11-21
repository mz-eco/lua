package lua

import (
	R "reflect"
)

type Invoker struct {
	Name    string
	GoFunc  interface{}
	Caller  func(vm *VM) (R.Value, error)
	CheckI  func(i []R.Type)
	CheckO  func(o []R.Type)
	Protect bool

	ft       R.Type
	ret      int
	hasError bool
	caller   bool
}

func (m *Invoker) iTypes() (types []R.Type) {

	types = make([]R.Type, 0)

	for index := 0; index < m.ft.NumIn(); index++ {
		types = append(types, m.ft.In(index))
	}

	return
}

func (m *Invoker) oTypes() (types []R.Type) {

	types = make([]R.Type, 0)

	for index := 0; index < m.ft.NumIn(); index++ {
		types = append(types, m.ft.In(index))
	}

	return
}

func (m *Invoker) iValues(vm *VM) (values []R.Value, index int, err error) {

	values = make([]R.Value, 0)

	var (
		opts = &asOptions{
			vm:       vm,
			field:    "",
			skipFunc: false,
			base:     Nil,
			optional: false,
			raise:    false,
		}
	)

	for index := 1; index < m.ft.NumIn(); index++ {

		var (
			lv = vm.Get(index + 1)
			it = m.ft.In(index)
		)

		if lv == Nil {
			values = append(values, R.Zero(it))
		} else {
			el := R.New(it)

			if goValue(vm, el.Elem(), lv, opts) && opts.Ok() {
				values = append(values, el.Elem())
			} else {
				return nil, index, opts.GetError(
					errTypeConvert{
						m.Name,
						lv.Type(),
						it,
					})
			}
		}
	}

	return values, 0, nil
}

func (m *Invoker) Invoke(vm *VM) int {

	var (
		fn, err = m.Caller(vm)
	)

	if err != nil {
		vm.ArgError(1, err.Error())
		return 0
	}

	call := newCall(vm, m.Name)

	if m.caller {
		caller := fn.Interface().(func(call *Call) int)
		return caller(call)
	}

	i, index, err := m.iValues(vm)

	if err != nil {
		vm.ArgError(index, err.Error())
		return 0
	}

	o := fn.Call(i)

	if m.hasError {
		err := o[m.ret]

		if !err.IsNil() {
			vm.RaiseError(
				"%s", err.Interface(),
			)
			return 0
		}
	}

	return call.oValues(o[0:m.ret])

}

func VMGFunction(i *Invoker) GFunction {

	if i.GoFunc == nil {
		panic("GoFunc must bu set")
	}

	switch fn := i.GoFunc.(type) {
	case R.Value:
		i.ft = fn.Type()
		if i.Caller == nil {
			i.Caller = func(vm *VM) (R.Value, error) {
				return fn, nil
			}
		}
	case R.Type:
		i.ft = fn

		if i.Caller == nil {
			panic("type caller must give Caller func")
		}
	default:
		i.ft = R.TypeOf(i.GoFunc)

		if i.Caller == nil {
			i.Caller = func(vm *VM) (R.Value, error) {
				return R.ValueOf(i.GoFunc), nil
			}
		}
	}

	if i.ft == typeCaller {
		i.caller = true
		return func(vm *VM) int {
			return i.Invoke(vm)
		}
	} else {
		i.caller = false
	}

	if i.ft.Kind() != R.Func {
		panic("GoFunc must type of func")
	}

	if i.CheckI != nil {
		i.CheckI(i.iTypes())
	}

	if i.CheckO != nil {
		i.CheckO(i.oTypes())
	}

	i.ret = i.ft.NumOut()
	i.hasError = false

	if i.ret > 0 {
		if i.ft.Out(i.ret - 1).Implements(typeError) {
			if !i.Protect {
				i.ret = i.ret - 1
				i.hasError = true
			}
		}
	}

	return func(vm *VM) int {
		return i.Invoke(vm)
	}
}

func memberFunctions(value interface{}, cb func(v R.Value, m int, i *Invoker)) (members map[string]GFunction) {

	var (
		x R.Type
		v = R.Value{}
	)

	switch n := value.(type) {
	case R.Type:
		x = n
	case R.Value:
		x = n.Type()
		v = n
	default:
		x = R.TypeOf(value)
		v = R.ValueOf(value)
	}

	if x.Kind() == R.Ptr {
		if x.Elem().Kind() != R.Struct {
			panic("member func value must a struct pointer or struct")
		}
	} else {
		if x.Kind() != R.Struct {
			panic("member func value must a struct pointer or struct")
		}
	}

	members = make(map[string]GFunction)

	for index := 0; index < x.NumMethod(); index++ {
		var (
			m = x.Method(index)
			i = &Invoker{
				Name:    m.Name,
				GoFunc:  m.Type,
				Protect: false,
			}
		)

		if cb != nil {
			cb(v, index, i)
		}

		members[i.Name] = VMGFunction(i)

	}

	return members

}
