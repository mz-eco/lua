package lua

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/yuin/gopher-lua"

	R "reflect"
)

type asOptions struct {
	vm       *VM
	field    string
	skipFunc bool
	base     Value
	error    error
	optional bool
	raise    bool
}

func (m *asOptions) Save() asOptions {
	return asOptions{
		field:    m.field,
		skipFunc: m.skipFunc,
		base:     m.base,
		optional: m.optional,
		raise:    m.raise,
	}
}

func (m *asOptions) Load(opts asOptions) {
	m.field = opts.field
	m.skipFunc = opts.skipFunc
	m.base = opts.base
	m.optional = opts.optional
	m.raise = opts.raise
}

func (m *asOptions) Ok() bool {
	return m.error == nil
}

func (m *asOptions) GetError(err error) error {

	if m.error != nil {
		return m.error
	}

	return err

}

func (m *asOptions) NewError(format string, v ...interface{}) bool {
	return m.Error(
		errors.New(fmt.Sprintf(
			format,
			v...)))
}

func (m *asOptions) Error(err error) bool {

	m.error = err

	if m.error != nil && m.raise {
		m.vm.RaiseError(err.Error())
	}
	return true
}

func goValuePod(vm *VM, v R.Value, src Value, opts *asOptions) (ok bool) {

	var (
		lt        = src.Type()
		kind      = v.Kind()
		t         = v.Type()
		typeError = errTypeConvert{opts.field, lt, v.Type()}
		number    = func(kind R.Kind) error {

			var (
				i64     int64
				ui64    uint64
				float   float64
				boolean bool
				err     error
			)

			switch lt {
			case lua.LTNil:
				if !opts.optional {
					opts.Error(typeError)
				}
				return nil
			case lua.LTNumber:
				var (
					n = lua.LVAsNumber(src)
				)

				i64 = int64(n)
				ui64 = uint64(n)
				float = float64(n)
				boolean = n != 0
			case lua.LTString:

				var (
					s = lua.LVAsString(src)
				)

				if len(s) == 0 {
					if opts.optional {
						break
					}
				}

				switch kind {
				case R.Int:
					i64, err = strconv.ParseInt(s, 10, 64)
				case R.Uint:
					ui64, err = strconv.ParseUint(s, 10, 64)
				case R.Float64:
					float, err = strconv.ParseFloat(s, 64)
				case R.Bool:
					boolean, err = strconv.ParseBool(s)
				}

				if err != nil {
					opts.Error(err)
				}
			}

			switch kind {
			case R.Int:
				v.SetInt(i64)
			case R.Uint:
				v.SetUint(ui64)
			case R.Float64:
				v.SetFloat(float)
			case R.Bool:
				v.SetBool(boolean)
			default:
				panic("error use")
			}

			return nil
		}
	)

	ok = true

	switch kind {
	case R.Int, R.Int8, R.Int16, R.Int32, R.Int64:
		number(R.Int)
	case R.Uint, R.Uint16, R.Uint32, R.Uint64:
		number(R.Uint)
	case R.String:
		switch lt {
		case lua.LTString, lua.LTNumber:
			v.SetString(lua.LVAsString(src))
		default:
			opts.Error(typeError)
		}
	case R.Bool:
		number(R.Bool)
	case R.Float32, R.Float64:
		number(R.Float64)
	case R.Interface:
		if t != typeInterface {
			opts.Error(typeError)
			return true
		}

		var (
			i interface{}
		)

		switch lt {
		case lua.LTNil:
			i = nil
		case lua.LTNumber:
			i = float64(lua.LVAsNumber(src))
		case lua.LTString:
			i = lua.LVAsString(src)
		case lua.LTBool:
			i = lua.LVAsBool(src)
		case lua.LTTable:
			opts.skipFunc = true
			var (
				values = make(map[interface{}]interface{})
			)

			if goValue(vm, R.ValueOf(values), src, opts) && opts.Ok() {
				i = values
			}

		default:
			if !opts.skipFunc {
				opts.Error(typeError)
				return true
			}
			return false
		}

		v.Set(R.ValueOf(i))
	case R.Slice:

		if lt != lua.LTTable {
			opts.Error(typeError)
			break
		}

		var (
			tbl   = src.(*Table)
			et    = t.Elem()
			slice = R.MakeSlice(t, 0, 0)
			osf   = opts.skipFunc
		)
		opts.skipFunc = true

		tbl.ForEach(func(_ lua.LValue, value lua.LValue) {

			if opts.error != nil {
				return
			}

			var (
				ptr = R.New(et)
			)

			if goValue(vm, ptr.Elem(), value, opts) {

				if opts.error != nil {
					return
				}

				slice = R.Append(slice, ptr.Elem())
			}
		})
		opts.skipFunc = osf

		v.Set(slice)
	case R.Map:

		if lt != lua.LTTable {
			opts.Error(typeError)
			return
		}

		var (
			tbl = src.(*Table)
			tk  = t.Key()
			tv  = t.Elem()
			vk  = R.New(tk)
			vv  = R.New(tv)
			m   = R.MakeMap(t)
			osf = opts.skipFunc
		)

		opts.skipFunc = true
		tbl.ForEach(func(key lua.LValue, value lua.LValue) {

			if !opts.Ok() {
				return
			}

			if goValue(vm, vk.Elem(), key, opts) && opts.Ok() {
				if goValue(vm, vv.Elem(), value, opts) && opts.Ok() {
					m.SetMapIndex(vk.Elem(), vv.Elem())
				}
			}
		})
		opts.skipFunc = osf

		v.Set(m)
	case R.Ptr:
		var (
			ev = R.New(t.Elem())
		)

		ok = goValue(vm, ev.Elem(), src, opts)

		if ok && opts.Ok() {
			v.Set(ev)
		}

	default:
		ok = false
	}
	return ok
}

var (
	typeInterface = R.TypeOf((*interface{})(nil)).Elem()
	typeError     = R.TypeOf((*error)(nil)).Elem()
)

func makeFunc(vm *VM, v R.Value, src, self Value) R.Value {

	return R.MakeFunc(
		v.Type(),
		func(args []R.Value) (results []R.Value) {

			var (
				i    = make([]Value, 0)
				t    = v.Type()
				fn   = src.(*Function)
				no   = t.NumOut()
				ret  = no - 1
				opts = &asOptions{
					field:    t.Name(),
					skipFunc: true,
					base:     src,
					raise:    false,
					optional: false,
				}
				encoder = NewEncoder(vm, FlagSkipMethod)
				value   Value
				err     error
			)

			results = make([]R.Value, no)

			for index := 0; index < no; index++ {
				results[index] = R.Zero(t.Out(index))
			}

			if self != Nil {
				i = append(i, self)
			}

			for index, arg := range args {

				value, err = encoder.Encode(arg)

				if err != nil {
					vm.ArgError(index, err.Error())
					break
				}

				i = append(i, value)
			}

			if err != nil {
				results[ret] = R.ValueOf((*error)(&err)).Elem()
				return
			}

			err = vm.CallByParam(
				lua.P{
					Fn:      fn,
					NRet:    ret,
					Protect: true,
				}, i...)

			if err != nil {
				results[ret] = R.ValueOf((*error)(&err)).Elem()
				return
			}

			for index := 0; index < ret; index++ {

				var (
					ot = t.Out(index)
					ov = R.New(ot)
					lv = vm.Get(-1 * (ret - index))
				)

				if lv.Type() == lua.LTTable {
					opts.base = lv
				} else {
					opts.base = Nil
				}

				if !goValue(vm, ov.Elem(), lv, opts) {
					results[ret] = R.ValueOf(
						opts.GetError(
							errTypeConvert{
								fmt.Sprintf("%s(Argument #%d)", t.Name(), index),
								src.Type(),
								ot,
							}))

					return
				}

				results[index] = ov.Elem()

			}

			return
		})
}

func goValueFunction(vm *VM, v R.Value, src Value, opts *asOptions) (ok bool) {

	ok = true
	var (
		t  = v.Type()
		lt = src.Type()
		k  = v.Kind()
	)

	if k != R.Func || lt != lua.LTFunction {
		return false
	}

	if t.NumOut() == 0 || t.Out(t.NumOut()-1) != typeError {
		return opts.NewError("bind lua function, the go function last out argument must bu error")
	}

	v.Set(makeFunc(vm, v, src, opts.base))

	return

}

func goValueObject(vm *VM, v R.Value, src Value, opts *asOptions) (ok bool) {

	var (
		t    = v.Type()
		kind = v.Kind()
		lt   = src.Type()
	)

	ok = true

	if kind != R.Struct || lt != lua.LTTable {
		return false
	}

	n := v.NumField()

	for index := 0; index < n; index++ {

		var (
			vf  = v.Field(index)
			tf  = t.Field(index)
			tag = makeTags(tf)
		)

		opts.optional = tag.option
		opts.field = tag.name
		opts.base = src

		if tag.skip {
			continue
		}

		lf := vm.GetField(src, tag.name)

		if !(goValue(vm, vf, lf, opts) && opts.Ok()) {
			return
		}

	}

	return

}

func goValueBuiltin(vm *VM, v R.Value, src Value, opts *asOptions) (ok bool) {

	if !v.CanInterface() {
		return false
	}

	var (
		i         = v.Interface()
		typeError = errTypeConvert{opts.field, src.Type(), v.Type()}
		x         interface{}
		is        = false
	)

	ok = true

	switch i.(type) {
	case *Table:
		x, is = src.(*Table)
	case Number:
		x, is = src.(Number)
	case String:
		x, is = src.(String)
	case Bool:
		x, is = src.(Bool)
	case lua.LChannel:
		x, is = src.(lua.LChannel)
	case *lua.LUserData:
		x, is = src.(*lua.LUserData)
	case *Function:
		x, is = src.(*Function)
	default:
		return false
	}

	if !is {
		opts.Error(typeError)
	} else {
		v.Set(R.ValueOf(x))
	}
	return
}

func goValue(vm *VM, v R.Value, src Value, opts *asOptions) bool {

	var (
		stack = opts.Save()
	)

	defer func() {
		opts.Load(stack)
	}()

	var (
		ok = true
		lt = src.Type()
	)

	switch lt {
	case lua.LTFunction:
		if opts.skipFunc {
			return false
		}
	case lua.LTChannel, lua.LTThread:
		return false
	}

	switch {
	case goValueBuiltin(vm, v, src, opts):
	case goValuePod(vm, v, src, opts):
	case goValueFunction(vm, v, src, opts):
	case goValueObject(vm, v, src, opts):

	default:
		ok = false
	}

	return ok
}

func As(vm *VM, src Value, value interface{}) error {

	var (
		v    = R.ValueOf(value)
		opts = &asOptions{
			vm:       vm,
			field:    "",
			skipFunc: false,
			base:     Nil,
			optional: false,
			raise:    true,
		}
	)

	if v.Kind() != R.Ptr {
		return errors.New("go value parser must give a point")
	}

	if goValue(vm, v.Elem(), src, opts) {
		if !opts.Ok() {
			return opts.error
		}

		return nil
	}

	return errTypeConvert{
		"", src.Type(), v.Type(),
	}
}
