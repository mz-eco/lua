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
	errObject       = errors.New("struct didn't implement lua.TableMapping")
	NotSupport      = errors.New("type not support")
	NotSupportFunc  = errors.New("type func support func")
)

type encoder struct {
	traces []interface{}
}

type EncodeError struct {
	traces string
	What   error
}

func (m *EncodeError) Error() string {
	return fmt.Sprintf("encoding field [%s] fail, %s", m.traces, m.What.Error())
}

type errorClass struct {
	class R.Type
}

func (m *errorClass) Error() string {
	return fmt.Sprintf("type <%s> is implementd lua.Typed, but not found in this vm", m.class)
}

type errorBuiltin struct {
	from lua.LValueType
}

func (m *errorBuiltin) Error() string {
	return fmt.Sprintf("unsupport convert builtin type <%s> to lua.Value", m.from)
}

func (m *encoder) errorBuiltin(from Value) error {
	return m.error(
		&errorBuiltin{
			from: from.Type(),
		})
}

func (m *encoder) errorClass(x R.Type) error {
	return m.error(
		&errorClass{
			class: x,
		})
}

func (m *encoder) error(err error) error {

	var (
		traces string = "src"
	)

	if m.traces != nil && len(m.traces) > 0 {
		x := make([]string, 0)

		for _, trace := range m.traces {
			x = append(x, fmt.Sprintf("%s", trace))
		}

		traces = strings.Join(x, ".")
	}

	return &EncodeError{
		traces: traces,
		What:   err,
	}
}
