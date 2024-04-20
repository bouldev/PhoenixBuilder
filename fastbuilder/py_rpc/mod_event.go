package py_rpc

import (
	"fmt"
	cts "phoenixbuilder/fastbuilder/py_rpc/mod_event/client_to_server"
	mei "phoenixbuilder/fastbuilder/py_rpc/mod_event/interface"
	stc "phoenixbuilder/fastbuilder/py_rpc/mod_event/server_to_client"
)

// Describe the type of PyRpc/ModEvent packet
const (
	ModEventServerToClient = "ModEventS2C"
	ModEventClientToServer = "ModEventC2S"
)

type ModEvent struct {
	Package mei.Package
	Type    string
}

// Return the name of m
func (m *ModEvent) Name() string {
	return m.Type
}

// Convert m to go object which only contains go-built-in types
func (m *ModEvent) MakeGo() (res any) {
	return []any{
		m.Package.PackageName(),
		m.Package.ModuleName(),
		m.Package.EventName(),
		m.Package.MakeGo(),
	}
}

// Sync data to m from obj
func (m *ModEvent) FromGo(obj any) error {
	object, success := obj.([]any)
	if !success {
		return fmt.Errorf("FromGo: Failed to convert obj to []interface{}; obj = %#v", obj)
	}
	if len(object) != 4 {
		return fmt.Errorf("FromGo: The length of object is not equal to 4; object = %#v", object)
	}
	// convert data and check it
	package_name, success := object[0].(string)
	if !success {
		return fmt.Errorf(`FromGo: Failed to convert object[0] to string; object[0] = %#v`, object[0])
	}
	module_name, success := object[1].(string)
	if !success {
		return fmt.Errorf(`FromGo: Failed to convert object[1] to string; object[1] = %#v`, object[1])
	}
	event_name, success := object[2].(string)
	if !success {
		return fmt.Errorf(`FromGo: Failed to convert object[2] to string; object[2] = %#v`, object[2])
	}
	event_data, success := object[3].(map[string]any)
	if !success {
		return fmt.Errorf(`FromGo: Failed to convert object[3] to map[string]interface{}; object[3] = %#v`, object[3])
	}
	// get data
	var park mei.Package
	var ok bool
	switch m.Type {
	case ModEventServerToClient:
		park, ok = stc.PackagePool()[package_name]
	case ModEventClientToServer:
		park, ok = cts.PackagePool()[package_name]
	}
	if !ok {
		park = &mei.Default{PACKAGE_NAME: package_name}
	}
	// if this package is not supported
	park.InitModuleFromPool(module_name, park.ModulePool())
	park.InitEventFromPool(event_name, park.EventPool())
	err := park.FromGo(event_data)
	if err != nil {
		return fmt.Errorf(`FromGo: %v`, err)
	}
	m.Package = park
	// init and sync data
	return nil
	// return
}
