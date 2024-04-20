package mod_event_interface

// Default Package, which used to
// describe the unsupported parts
type Default struct {
	PACKAGE_NAME string // The name of this package
	Module              // The module contained in this package
}

// Return the package name of d
func (d *Default) PackageName() string {
	return d.PACKAGE_NAME
}

// Return a pool/map that contains all the module of m
func (d *Default) ModulePool() map[string]Module {
	return map[string]Module{}
}

// Init the module data from pool
func (d *Default) InitModuleFromPool(
	module_name string,
	pool map[string]Module,
) {
	module, ok := pool[module_name]
	if !ok {
		d.Module = &DefaultModule{
			MODULE_NAME: module_name,
			Event:       &DefaultEvent{},
		}
		return
	}
	d.Module = module
}
