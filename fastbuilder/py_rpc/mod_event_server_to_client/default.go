package mod_event_server_to_client

import (
	"fmt"
	"phoenixbuilder/fastbuilder/py_rpc/mod_event_server_to_client/general_interface"
)

// Default Package, which used to
// describe the unsupported parts
type Default struct {
	package_name   string // The name of this package
	default_module        // The module contained in this package
}

// Return the package name of d
func (d *Default) PackageName() string {
	return d.package_name
}

// Return a pool/map that contains all the module of m
func (d *Default) ModulePool() map[string]general_interface.Module {
	return map[string]general_interface.Module{
		d.ModuleName(): &default_module{},
	}
}

// Default Module, which used to
// describe the unsupported parts
type default_module struct {
	module_name   string // The name of this module
	default_event        // The event contained in this module
}

// Return the module name of d
func (d *default_module) ModuleName() string {
	return d.module_name
}

// Return a pool/map that contains all the event of d
func (d *default_module) EventPool() map[string]general_interface.Event {
	return map[string]general_interface.Event{
		d.EventName(): &default_event{},
	}
}

// Default event, which used to
// describe the unsupported parts
type default_event struct {
	event_name string         // The name of this event
	data       map[string]any // The data contained in this event
}

// Return the event name of d
func (d *default_event) EventName() string {
	return d.event_name
}

// Convert d to go object which only contains go-built-in types
func (d *default_event) MakeGo() (res any) {
	return d.data
}

// Sync data to d from obj
func (d *default_event) FromGo(res any) error {
	result, success := res.(map[string]any)
	if !success {
		return fmt.Errorf("FromGo: Failed to covert res to map[string]interface{}; res= %#v", res)
	}
	// convert data
	d.data = result
	return nil
	// sync data and return
}
