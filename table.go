package lua

import R "reflect"

type Fields map[string]interface{}

type tableMapping interface {
	table()
}

type TableMapping struct {
}

func (m *TableMapping) table() {}

var (
	typeTableMapping = R.TypeOf((*tableMapping)(nil))
)
