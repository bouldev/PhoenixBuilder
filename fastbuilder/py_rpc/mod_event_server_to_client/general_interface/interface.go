package general_interface

// Express a package which contains in a PyRpc packet
type Package interface {
	// Return the name of Package
	PackageName() string
	// Return a pool/map that contains all the module of this package
	ModulePool() map[string]Module
	// Describe the corresponding module of this package
	Module
}

// Describe a module which contains in a package
type Module interface {
	// Return the module name of Module
	ModuleName() string
	// Return a pool/map that contains all the event of this module
	EventPool() map[string]Event
	// Describe the corresponding event of this module
	Event
}

// Describe an event which contains in a module
type Event interface {
	// Return the event name of this event
	EventName() string
	// Convert this event to go object which only contains go-built-in types
	MakeGo() (res any)
	// Sync data to this event from obj
	FromGo(obj any) error
}
