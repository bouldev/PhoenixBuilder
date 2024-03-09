package mod_event_interface

// Default Module, which used to
// describe the unsupported parts
type DefaultModule struct {
	MODULE_NAME string // The name of this module
	Event              // The event contained in this module
}

// Return the module name of d
func (d *DefaultModule) ModuleName() string {
	return d.MODULE_NAME
}

// Return a pool/map that contains all the event of d
func (d *DefaultModule) EventPool() map[string]Event {
	return map[string]Event{}
}

// Init the event data for d from pool
func (d *DefaultModule) InitEventFromPool(
	event_name string,
	pool map[string]Event,
) {
	event, ok := pool[event_name]
	if !ok {
		d.Event = &DefaultEvent{EVENT_NAME: event_name}
		return
	}
	d.Event = event
}
