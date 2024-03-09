package vip_event_system

import "fmt"

// Used on initialize the
// Minecraft connection
type PlayerUiInit struct {
	RuntimeID string
}

// Return the event name of p
func (p *PlayerUiInit) EventName() string {
	return "PlayerUiInit"
}

// Convert p to go object which only contains go-built-in types
func (p *PlayerUiInit) MakeGo() (res any) {
	return p.RuntimeID
}

// Sync data to p from obj
func (p *PlayerUiInit) FromGo(obj any) error {
	object, success := obj.(string)
	if !success {
		return fmt.Errorf("FromGo: Failed to convert obj to string; obj = %#v", obj)
	}
	p.RuntimeID = object
	// convert and sync data
	return nil
	// return
}
