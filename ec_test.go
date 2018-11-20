package lua

import (
	"fmt"
	"reflect"
	"testing"
)

type AA struct {
	CC int
}

func TestName(t *testing.T) {

	at := reflect.TypeOf(AA{})
	ax := reflect.Zero(at)

	fmt.Println(ax)
}
