package lua

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

}
