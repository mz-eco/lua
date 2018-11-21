package lua

import (
	"errors"
	R "reflect"

	lua "github.com/yuin/gopher-lua"
)

var (
	errNotPointer = errors.New("argument must a pointer")
	errNotType    = errors.New("not support <reflect.Type> to <lua.Value>")
	errChanType   = errors.New("channel tye is not lua.LChannel")
)

type errorFieldNotFound struct {
	field string
}

func (m *errorFieldNotFound) Error() string {
	return m.field
}

type errorClassType struct {
	src R.Type
	to  R.Type
}

func (m *errorClassType) Error() string {
	return ""
}

type errorConvert struct {
	src lua.LValueType
	to  R.Type
}

func (m *errorConvert) Error() string {
	return ""
}

type Decoder struct {
	encoding
	vm         *VM
	skipMethod bool
	base       Value
}

func (m *Decoder) errConvert(src Value, to R.Type) error {

	return m.error(
		&errorConvert{
			src: src.Type(),
			to:  to,
		})
}

func (m *Decoder) errClassType(src R.Type, to R.Type) error {

	return m.error(
		&errorClassType{
			src: src,
			to:  to,
		})
}

func (m *Decoder) fieldNotFound(field string) error {
	return m.error(
		&errorFieldNotFound{
			field: field,
		})
}

func NewDeocder(vm *VM, flags EncodingFlags) *Decoder {

	return &Decoder{
		vm:         vm,
		skipMethod: (flags & FlagSkipMethod) == FlagSkipMethod,
		base:       Nil,
	}
}

func (m *Decoder) Decode(src Value, to interface{}) (err error) {

	var v R.Value

	switch x := to.(type) {
	case R.Value:
		v = x
	case R.Type:
		return m.error(errNotType)
	default:
		v = R.ValueOf(src)
	}

	if v.Kind() != R.Ptr {
		return m.error(errNotPointer)
	}

	return m.decode(src, v)

}

func (m *Decoder) class(src Value, to R.Value) error {

	var (
		t = to.Type()
	)

	if src.Type() != lua.LTUserData {
		return m.errorBuiltin(src, t)
	}

	gt, ok := typeLookup(m.vm, t)

	if !ok {
		return m.errorClass(t)
	}

	if gt.Type != t {
		return m.errClassType(gt.Type, t)
	}

	ud := src.(*lua.LUserData)
	uv := R.ValueOf(ud)
	ut := uv.Type()

	if ut != t {
		return m.errClassType(ut, t)
	}

	to.Set(uv)

	return nil
}

func (m *Decoder) mapping(src Value, to R.Value) error {

	if src.Type() != lua.LTTable {
		return m.errConvert(src, to.Type())
	}

	var (
		ot  = to.Type()
		tbl = src.(*Table)
	)

	for index := 0; index < ot.NumField(); index++ {

		var (
			fv  = to.Field(index)
			ft  = ot.Field(index)
			tag = makeTags(ft)
		)

		if tag.skip {
			continue
		}

		lv := m.vm.GetField(tbl, ft.Name)

		if lv == Nil {
			if tag.option {
				continue
			}

			return m.fieldNotFound(ft.Name)

		}
		err := m.decode(lv, fv, ft.Name)

		if err != nil {
			return err
		}

	}

	return nil

}

func (m *Decoder) bool(src Value, to R.Value) error {

	if src.Type() != lua.LTBool {
		return m.errConvert(src, to.Type())
	}

	to.SetBool(lua.LVAsBool(src))

	return nil
}

func (m *Decoder) checkNil(src Value, to R.Value) bool {

	return src == Nil
}

func (m *Decoder) slice(src Value, to R.Value) error {

	var (
		t      = R.TypeOf(to)
		values = R.MakeSlice(t, 0, 0)
	)

	if src.Type() != lua.LTTable {
		return m.errConvert(src, t)
	}

	var (
		tbl   = src.(*Table)
		err   error
		et    = t.Elem()
		index = 0
	)

	tbl.ForEach(func(_ lua.LValue, value lua.LValue) {

		if err != nil {
			return
		}

		e := R.New(et)

		err = m.decode(value, e.Elem(), index)

		if err != nil {
			return
		}

		values = R.Append(values, e.Elem())

	})

	if err != nil {
		return err
	}

	to.Set(values)

	return nil
}

func (m *Decoder) dir(src Value, to R.Value) error {

	var (
		err error
	)

	if src.Type() != lua.LTTable {
		return m.errConvert(src, to.Type())
	}

	var (
		mt   = R.TypeOf(to)
		kt   = mt.Key()
		vt   = mt.Elem()
		dirs = R.MakeMap(mt)
		tbl  = src.(*Table)
	)

	tbl.ForEach(func(key lua.LValue, value lua.LValue) {

		if err != nil {
			return
		}

		kv := R.New(kt)

		err = m.decode(key, kv)

		if err != nil {
			return
		}

		vv := R.New(vt)

		err = m.decode(value, vv)

		if err != nil {
			return
		}

		dirs.SetMapIndex(kv, vv)

	})

	if err != nil {
		return err
	}

	to.Set(dirs)
	return nil
}

func (m *Decoder) ptr(src Value, to R.Value) error {

	if m.checkNil(src, to) {
		return nil
	}

	var (
		ptr = R.New(to.Type())
	)

	err := m.decode(src, ptr.Elem())

	if err != nil {
		return err
	}

	to.Set(ptr.Elem())

	return nil
}

func (m *Decoder) builtin(src Value, to R.Value) error {

	var (
		i = to.Interface().(Value)

		check = func(lvt lua.LValueType) error {

			if src.Type() != lvt {
				return m.errorBuiltin(src, to.Type())
			}

			to.Set(R.ValueOf(src))

			return nil
		}
	)

	switch i.(type) {
	case *Function:
		return check(lua.LTFunction)
	case Value:
		to.Set(R.ValueOf(src))
		return nil
	case Number:
		return check(lua.LTNumber)
	case *Table:
		return check(lua.LTTable)
	case Bool:
		return check(lua.LTBool)
	case String:
		return check(lua.LTString)
	case lua.LChannel:
		return check(lua.LTTable)
	case *lua.LUserData:
		return check(lua.LTUserData)
	default:
		return m.errorBuiltin(src, to.Type())
	}

}

func (m *Decoder) fn(from Value, to R.Value) error {
	return nil
}

func (m *Decoder) channel(src Value, to R.Value) error {

	if m.checkNil(src, to) {
		return nil
	}

	var (
		t = R.TypeOf(to)
	)

	if t != typeChannel {
		return m.error(errChanType)
	}

	to.Set(R.ValueOf(src))
	return nil
}

func (m *Decoder) str(src Value, to R.Value) error {

	if src.Type() != lua.LTString {
		return m.errConvert(src, to.Type())
	}

	to.SetString(lua.LVAsString(src))
	return nil
}

func (m *Decoder) int(src Value, to R.Value) error {

	if src.Type() != lua.LTNumber {
		return m.errConvert(src, to.Type())
	}

	to.SetInt(int64(lua.LVAsNumber(src)))
	return nil
}

func (m *Decoder) uint(src Value, to R.Value) error {
	if src.Type() != lua.LTNumber {
		return m.errConvert(src, to.Type())
	}

	to.SetUint(uint64(lua.LVAsNumber(src)))
	return nil
}

func (m *Decoder) float(src Value, to R.Value) error {

	if src.Type() != lua.LTNumber {
		return m.errConvert(src, to.Type())
	}

	to.SetFloat(float64(lua.LVAsNumber(src)))
	return nil
}

func (m *Decoder) decode(src Value, to R.Value, trace ...interface{}) error {

	if len(trace) > 0 {

		var (
			size = len(trace)
		)

		if m.traces == nil {
			m.traces = make([]interface{}, 0)
		}

		m.traces = append(m.traces, trace...)

		defer func() {
			m.traces = m.traces[0:size]
		}()
	}

	if !to.IsValid() {
		return m.error(errInvalidValue)
	}

	var (
		t = to.Type()
	)

	switch {
	case t.Implements(typeClass):
		return m.class(src, to)
	case t.Implements(typeTableMapping):
		return m.mapping(src, to)
	case t.Implements(typeValue):
		return m.builtin(src, to)
	}

	switch to.Kind() {
	case R.Ptr:
		return m.ptr(src, to)
	case R.Chan:
		return m.channel(src, to)
	case R.Int, R.Int8, R.Int16, R.Int32, R.Int64:
		return m.int(src, to)
	case R.Uint, R.Uint8, R.Uint16, R.Uint32, R.Uint64:
		return m.uint(src, to)
	case R.Float32, R.Float64:
		return m.float(src, to)
	case R.String:
		return m.str(src, to)
	case R.Struct:
		return m.errorObject(to.Type())
	case R.Slice:
		return m.slice(src, to)
	case R.Map:
		return m.mapping(src, to)
	case R.Func:
		return NotSupportFunc
	case R.Interface, R.Complex64, R.Complex128, R.Array, R.UnsafePointer, R.Invalid, R.Uintptr:
		fallthrough
	default:
		return NotSupport
	}

	return nil
}
