package mod_event_interface

// Default event, which used to
// describe the unsupported parts
type DefaultEvent struct {
	EVENT_NAME string // The name of this event
	Data       any    // The data contained in this event
}

// Return the event name of d
func (d *DefaultEvent) EventName() string {
	return d.EVENT_NAME
}

// Convert d to go object which only contains go-built-in types
func (d *DefaultEvent) MakeGo() (res any) {
	return d.Data
}

// Sync data to d from obj
func (d *DefaultEvent) FromGo(res any) error {
	d.Data = res
	return nil
}
