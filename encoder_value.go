package lua

import (
	R "reflect"

	lua "github.com/yuin/gopher-lua"
)

type Encoder struct {
	encoder
	vm         *VM
	skipMethod bool
}

func NewEncoder(vm *VM, skipMethod bool) *Encoder {

	return &Encoder{
		vm:         vm,
		skipMethod: skipMethod,
	}
}

func (m *Encoder) Encode(src interface{}) (to Value, err error) {

	var v R.Value

	switch x := src.(type) {
	case R.Value:
		v = x
	default:
		v = R.ValueOf(src)
	}

	to = Nil
	err = m.encode(v, &to)

	if err != nil {
		return Nil, err
	}

	return to, nil
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

func (m *Encoder) mapping(ov R.Value, to *Value) error {

	var (
		fields  = make(map[string]Value)
		members map[string]GFunction
		ot      = ov.Type()
	)

	for index := 0; index < ov.NumField(); index++ {

		var (
			fv    = ov.Field(index)
			ft    = ot.Field(index)
			tag   = makeTags(ft)
			value = Nil
		)

		if tag.skip {
			continue
		}

		err := m.encode(fv, &value, ft.Name)

		if err != nil {
			if err != NotSupportFunc {
				return err
			}
		}

		if value == Nil && tag.option {
			continue
		}

		fields[ft.Name] = value

	}

	if !m.skipMethod {
		members = memberFunctions(ov, nil)
	}

	tbl := m.vm.NewTable()
	for k, v := range fields {
		m.vm.SetField(tbl, k, v)
	}

	if members != nil {
		m.vm.SetField(tbl, "__index",
			m.vm.SetFuncs(
				m.vm.NewTable(), members),
		)
	}

	*to = tbl

	return nil

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
		values = make([]Value, 0)
	)

	for index := 0; index < src.Len(); index++ {

		var (
			v = Nil
		)

		err := m.encode(src.Index(index), &v, index)

		if err != nil {
			return err
		}

		values = append(values, v)
	}

	tbl := m.vm.NewTable()

	for _, value := range values {
		tbl.Append(value)
	}

	*to = tbl

	return nil
}

func (m *Encoder) dir(src R.Value, to *Value) error {

	var (
		values = make(map[Value]Value, 0)
		err    error
	)

	for _, key := range src.MapKeys() {

		var (
			k, v = Nil, Nil
		)

		err = m.encode(key, &k, "key")

		if err != nil {
			return err
		}

		err = m.encode(src.MapIndex(key), &v, "value")

		if err != nil {
			return err
		}

		values[k] = v
	}

	tbl := m.vm.NewTable()

	for k, v := range values {
		tbl.RawSet(k, v)
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

func (m *Encoder) encode(src R.Value, to *Value, trace ...interface{}) error {

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
		return NotSupportFunc
	case R.Interface, R.Complex64, R.Complex128, R.Array, R.UnsafePointer, R.Invalid, R.Uintptr:
		fallthrough
	default:
		return NotSupport
	}

	return nil
}
