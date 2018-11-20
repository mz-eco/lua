package lua

import (
	"fmt"

	R "reflect"
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
