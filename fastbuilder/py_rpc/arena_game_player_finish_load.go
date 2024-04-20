package py_rpc

import "fmt"

// Used on initialize the
// Minecraft connection
type ArenaGamePlayerFinishLoad []any

// Return the name of a
func (a *ArenaGamePlayerFinishLoad) Name() string {
	return "arenaGamePlayerFinishLoad"
}

// Convert a to go object which only contains go-built-in types
func (a *ArenaGamePlayerFinishLoad) MakeGo() (res any) {
	if a == nil {
		return nil
	}
	return []any(*a)
}

// Sync data to a from obj
func (a *ArenaGamePlayerFinishLoad) FromGo(obj any) error {
	object, success := obj.([]any)
	if !success {
		return fmt.Errorf("FromGo: Failed to convert obj to []interface{}; obj = %#v", obj)
	}
	*a = object
	// convert and sync data
	return nil
	// return
}
