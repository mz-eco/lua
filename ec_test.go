package lua

import (
	"fmt"
	"testing"
)

type AA struct {
	CC int
}

func TestName(t *testing.T) {

	v := []int{1, 2, 3, 4}

	fmt.Println(v[0:3])
}
