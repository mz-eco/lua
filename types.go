package lua

import lua "github.com/yuin/gopher-lua"

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
