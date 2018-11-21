package lua

import (
	"errors"
	"fmt"
	R "reflect"
	"strings"

	"github.com/yuin/gopher-lua"
)

var (
	errInvalidValue = errors.New("invalid value")
	errNilValue     = errors.New("nil value")
	NotSupport      = errors.New("type not support")
	NotSupportFunc  = errors.New("function encoding not support")
	errTypedChild   = errors.New("lua.Typed struct didn't have any lua.TableMapping child")
)

type EncodeError struct {
	traces string
	What   error
}

type errorObject struct {
	src R.Type
}

func (m *errorObject) Error() string {
	return fmt.Sprintf(
		"struct type <%s> didn't implement lua.TableMapping", m.src,
	)
}

func (m *EncodeError) Error() string {
	return fmt.Sprintf("encoding%s fail [%s]", m.traces, m.What.Error())
}

type errorClass struct {
	class R.Type
}

func (m *errorClass) Error() string {
	return fmt.Sprintf("type <%s> is implementd lua.Typed, but not found in this vm", m.class)
}

type errorBuiltin struct {
	from lua.LValueType
	to   R.Type
}

func (m *errorBuiltin) Error() string {
	return fmt.Sprintf("unsupport convert builtin type <%s> to <%s>", m.from, m.to)

}

type encoding struct {
	traces []interface{}
}

func (m *encoding) errorObject(src R.Type) error {

	return m.error(
		&errorObject{
			src: src,
		})
}

func (m *encoding) errorBuiltin(from Value, to R.Type) error {
	return m.error(
		&errorBuiltin{
			from: from.Type(),
			to:   to,
		})
}

func (m *encoding) errorClass(x R.Type) error {
	return m.error(
		&errorClass{
			class: x,
		})
}

func (m *encoding) error(err error) error {

	var (
		traces string = ""
	)

	if m.traces != nil && len(m.traces) > 0 {
		x := make([]string, 0)

		for _, trace := range m.traces {
			x = append(x, fmt.Sprintf("%s", trace))
		}

		traces = strings.Join(x, ".")
		traces = " <" + traces + ">"
	}

	return &EncodeError{
		traces: traces,
		What:   err,
	}
}
