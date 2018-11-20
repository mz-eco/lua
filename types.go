package lua

import (
	R "reflect"

	lua "github.com/yuin/gopher-lua"
)

type VM = lua.LState

type (
	Table     = lua.LTable
	Number    = lua.LNumber
	String    = lua.LString
	Bool      = lua.LBool
	Function  = lua.LFunction
	GFunction = lua.LGFunction
	Value     = lua.LValue
)

var (
	Nil = lua.LNil
)

type tableMapping interface {
	table()
}

type class interface {
	setValue(value Value)
	getValue() Value
}

type moduleMembers interface {
	module()
}

func typedCaller(_ *Call) int { return 0 }

var (
	typeNil          = R.TypeOf(nil)
	typeTableMapping = R.TypeOf((*tableMapping)(nil))
	typeClass        = R.TypeOf((*class)(nil)).Elem()
	typeCall         = R.TypeOf((*Call)(nil))
	typeCaller       = R.TypeOf(typedCaller)
	typeValue        = R.TypeOf((*Value)(nil)).Elem()
)
