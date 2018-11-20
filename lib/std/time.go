package std

import (
	"time"

	"github.com/mz-eco/lua"
)

func init() {
	module.Define(&lua.Type{
		UUID: "7ba15f02-09c1-4b8d-9c34-7dcd29599dd4",
		Name: "Time",
		Type: lua.GoType((*Time)(nil)),
	})
}

type AX struct {
	DD int
}

type Time struct {
	lua.Typed
	tm time.Time

	Number *AX
}

func (*Members) Now() *Time {
	return &Time{
		tm:     time.Now(),
		Number: &AX{DD: 14},
	}
}

func (m *Time) String() string {
	return m.tm.String()
}
