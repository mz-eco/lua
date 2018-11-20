package lua

import (
	"fmt"
	"reflect"
	"testing"
)

func X(call *Call) {

}

type AX struct {
}

func (m *AX) call(call *Call) {

}

func TestName(t *testing.T) {

	var (
		aa = &AX{}
		x  = reflect.TypeOf(X)
		ax = reflect.TypeOf(aa.call)
	)

	fmt.Println(x == typeCaller, ax == typeCaller, ax)

}
