package lua

import (
	R "reflect"
)

type moduleMembers interface {
	module()
}

type ModuleMembers struct {
}

func (m *ModuleMembers) module() {

}

type ModuleLoader func() *Module

type Module struct {
	Name    string
	Members moduleMembers
	types   []*Type
}

func (m *Module) Define(x *Type) {
	m.types = append(m.types, x)
}

func LoadModule(vm *VM, loader ModuleLoader) {

	var (
		m = loader()
	)

	vm.PreloadModule(
		m.Name, func(vm *VM) int {

			var (
				tbl = vm.NewTable()
			)

			for _, x := range m.types {
				Define(vm, x)
			}

			if m.Members != nil {

				members := memberFunctions(
					m.Members,
					func(v R.Value, index int, i *Invoker) {
						i.Caller = func(vm *VM) (R.Value, error) {
							return v.Method(index), nil
						}
					},
				)

				vm.SetFuncs(tbl, members)
			}

			vm.Push(tbl)

			return 1
		})
}
