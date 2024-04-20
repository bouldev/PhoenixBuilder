package mod_event_interface

import "fmt"

// Default event, which used to
// describe the unsupported parts
type DefaultEvent struct {
	EVENT_NAME string         // The name of this event
	Data       map[string]any // The data contained in this event
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
	result, success := res.(map[string]any)
	if !success {
		return fmt.Errorf("FromGo: Failed to covert res to map[string]interface{}; res = %#v", res)
	}
	// convert data
	d.Data = result
	return nil
	// sync data and return
}
