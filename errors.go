package lua

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/yuin/gopher-lua"
)

type errTypeConvert struct {
	field string
	src   lua.LValueType
	dst   reflect.Type
}

func (m errTypeConvert) Error() string {
	return fmt.Sprintf(
		"cloud not convert lua type %s to go field:%s, type %s", m.src, m.field, m.dst)
}

var (
	ErrNil = errors.New("nil value")
)

type errNewType struct {
	t reflect.Type
}

func (m errNewType) Error() string {
	return fmt.Sprintf(
		"cloud not new go type %s", m.t)
}

type errValueConvert struct {
	field string
	src   string
	dst   reflect.Type
}

func (m errValueConvert) Error() string {
	return fmt.Sprintf(
		"cloud not convert value %s to go field:%s, type %s", m.src, m.field, m.dst)
}
