package lua

import (
	"errors"
	"fmt"
	R "reflect"

	"github.com/yuin/gopher-lua"
)

var (
	errInvalidValue = errors.New("invalid value")
	errNilValue     = errors.New("nil value")
	errObject       = errors.New("struct didn't implement lua.TableMapping")
	NotSupport      = errors.New("type not support")
)

type EncodeError struct {
	What error
}

func (m *EncodeError) Error() string {
	return ""
}

type errorKind struct {
	kind R.Kind
}

type errorClass struct {
	class R.Type
}

func (m *errorClass) Error() string {
	return ""
}

type errorBultin struct {
	from lua.LValueType
}

func (m *errorBultin) Error() string {
	return ""
}

func (m *errorKind) Error() string {
	return fmt.Sprintf("")
}

type stack struct {
	names []string
}

type Encoder struct {
	vm         *VM
	skipMethod bool
}

func (m *Encoder) errorBuiltin(from Value) error {
	return m.error(
		&errorBultin{
			from: from.Type(),
		})
}

func (m *Encoder) errorKind(kind R.Kind) error {

	return m.error(
		&errorKind{
			kind: kind,
		})
}

func (m *Encoder) errorClass(x R.Type) error {
	return m.error(
		&errorClass{
			class: x,
		})
}

func (m *Encoder) error(err error) error {

	return &EncodeError{
		What: err,
	}
}

func (m *Encoder) Encode(src interface{}, to *Value) error {

	var v R.Value

	switch x := src.(type) {
	case R.Value:
		v = x
	default:
		v = R.ValueOf(src)
	}

	return m.encode(v, to)
}

func (m *Encoder) class(src R.Value, to *Value) error {

	var (
		vt     = src.Type()
		xt, ok = typeLookup(m.vm, vt)
	)

	if !ok {
		return m.errorClass(vt)
	}

	i := src.Interface().(class)

	value := i.getValue()

	if value != nil {
		*to = value
	} else {
		ud := m.vm.NewUserData()
		ud.Value = i

		m.vm.SetMetatable(ud, m.vm.GetTypeMetatable(xt.name))
		i.setValue(ud)
		*to = ud
	}

	return nil
}

func (m *Encoder) mapping(src R.Value, to *Value) error {

	var (
		tbl = m.vm.NewTable()
	)

}

func (m *Encoder) bool(src R.Value, to *Value) error {
	*to = Bool(src.Bool())
	return nil
}

func (m *Encoder) checkNil(src R.Value, to *Value) bool {

	if src.IsNil() {
		*to = Nil
		return true
	}

	return false
}

func (m *Encoder) slice(src R.Value, to *Value) error {

	var (
		tbl = m.vm.NewTable()
	)

	for index := 0; index < src.Len(); index++ {

		var (
			v = Nil
		)

		err := m.encode(src.Index(index), &v)

		if err != nil {
			return err
		}

		tbl.Append(v)
	}

	*to = tbl

	return nil
}

func (m *Encoder) dir(src R.Value, to *Value) error {

	var (
		tbl = m.vm.NewTable()
		err error
	)

	for _, key := range src.MapKeys() {

		var (
			k, v = Nil, Nil
		)

		err = m.encode(key, &k)

		if err != nil {
			return err
		}

		err = m.encode(src.MapIndex(key), &v)

		if err != nil {
			return err
		}

		tbl.RawSetH(k, v)
	}

	*to = tbl
	return nil
}

func (m *Encoder) ptr(src R.Value, to *Value) error {

	if m.checkNil(src, to) {
		return nil
	}

	return m.encode(src.Elem(), to)
}

func (m *Encoder) builtin(src R.Value, x *Value) error {

	var (
		i = src.Interface().(Value)
	)

	switch v := i.(type) {
	case *Function:
		*x = v
	case Value:
		*x = v
	case Number:
		*x = v
	case *Table:
		*x = v
	case Bool:
		*x = v
	case String:
		*x = v
	case lua.LChannel:
		*x = v
	case *lua.LUserData:
		*x = v
	default:
		return m.errorBuiltin(v)
	}

	return nil
}

func (m *Encoder) fn(src R.Value, to *Value) error {
	return nil
}

func (m *Encoder) channel(src R.Value, to *Value) error {

	if m.checkNil(src, to) {
		return nil
	}

	*to = src.Interface().(lua.LChannel)

	return nil
}

func (m *Encoder) str(src R.Value, to *Value) error {
	*to = String(src.String())
	return nil
}

func (m *Encoder) int(src R.Value, to *Value) error {
	*to = Number(src.Int())
	return nil
}

func (m *Encoder) uint(src R.Value, to *Value) error {
	*to = Number(src.Uint())
	return nil
}

func (m *Encoder) float(src R.Value, to *Value) error {
	*to = Number(src.Float())
	return nil
}

func (m *Encoder) encode(src R.Value, to *Value) error {

	if !src.IsValid() {
		return m.error(errInvalidValue)
	}

	var (
		t = src.Type()
	)

	switch {
	case t.Implements(typeClass):
		return m.class(src, to)
	case t.Implements(typeTableMapping):
		return m.mapping(src, to)
	case t.Implements(typeValue):
		return m.builtin(src, to)
	}

	switch src.Kind() {
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
		return m.error(errObject)
	case R.Slice:
		return m.slice(src, to)
	case R.Map:
		return m.mapping(src, to)
	case R.Func:
		fallthrough
	case R.Interface, R.Complex64, R.Complex128, R.Array, R.UnsafePointer, R.Invalid, R.Uintptr:
		fallthrough
	default:
		return NotSupport
	}

	return nil
}
