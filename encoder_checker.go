package lua

import (
	R "reflect"
)

type EncodeChecker struct {
	encoder
	skipMethod bool
}

func (m *EncodeChecker) Encode(src interface{}) error {

	switch x := src.(type) {
	case R.Value:
		return m.encode(x.Type())
	case R.Type:
		return m.encode(x)
	default:
		return m.encode(R.TypeOf(src))
	}

}

func (m *EncodeChecker) class(src R.Type) error {
	return nil
}

func (m *EncodeChecker) mapping(src R.Type) error {

	for index := 0; index < src.NumField(); index++ {

		var (
			ft  = src.Field(index)
			tag = makeTags(ft)
		)

		if tag.skip {
			continue
		}

		err := m.encode(ft.Type, ft.Name)

		if err != nil {
			if err != NotSupportFunc {
				return err
			}
		}

	}

	if !m.skipMethod {
		for index := 0; index < src.NumMethod(); index++ {
			var (
				x  = src.Method(index)
				mt = x.Type
			)

			for in := 0; in < mt.NumIn(); in++ {
				err := m.encode(mt.In(index), x.Name, "in", in)

				if err != nil {
					return err
				}
			}

			for ou := 0; ou < mt.NumOut(); ou++ {
				err := m.encode(mt.Out(index), x.Name, "out", ou)

				if err != nil {
					return err
				}
			}
		}
	}

	return nil

}

func (m *EncodeChecker) bool(src R.Type) error {
	return nil
}

func (m *EncodeChecker) slice(src R.Type) error {

	if err := m.encode(src.Elem(), src.Name()); err != nil {
		return err
	}

	return nil
}

func (m *EncodeChecker) dir(src R.Type) error {

	if err := m.encode(src.Key(), src.Name()); err != nil {
		return err
	}

	if err := m.encode(src.Elem(), src.Name()); err != nil {
		return err
	}

	return nil
}

func (m *EncodeChecker) ptr(src R.Type) error {
	return m.encode(src.Elem())
}

func (m *EncodeChecker) builtin(src R.Type) error {
	return nil
}

func (m *EncodeChecker) fn(src R.Type) error {
	return nil
}

func (m *EncodeChecker) channel(src R.Type) error {
	return nil
}

func (m *EncodeChecker) str(src R.Type) error {
	return nil
}

func (m *EncodeChecker) int(src R.Type) error {
	return nil
}

func (m *EncodeChecker) uint(src R.Type) error {
	return nil
}

func (m *EncodeChecker) float(src R.Type) error {
	return nil
}

func (m *EncodeChecker) encode(src R.Type, trace ...interface{}) error {

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

	var (
		t = src
	)

	switch {
	case t.Implements(typeClass):
		return m.class(src)
	case t.Implements(typeTableMapping):
		return m.mapping(src)
	case t.Implements(typeValue):
		return m.builtin(src)
	}

	switch src.Kind() {
	case R.Ptr:
		return m.ptr(src)
	case R.Chan:
		return m.channel(src)
	case R.Int, R.Int8, R.Int16, R.Int32, R.Int64:
		return m.int(src)
	case R.Uint, R.Uint8, R.Uint16, R.Uint32, R.Uint64:
		return m.uint(src)
	case R.Float32, R.Float64:
		return m.float(src)
	case R.String:
		return m.str(src)
	case R.Struct:
		return m.error(errObject)
	case R.Slice:
		return m.slice(src)
	case R.Map:
		return m.mapping(src)
	case R.Func:
		fallthrough
	case R.Interface, R.Complex64, R.Complex128, R.Array, R.UnsafePointer, R.Invalid, R.Uintptr:
		fallthrough
	default:
		return NotSupport
	}
}
