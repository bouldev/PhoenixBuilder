package py_rpc

import "fmt"

// Used on initialize the
// Minecraft connection
type SyncUsingMod []any

// Return the name of s
func (s *SyncUsingMod) Name() string {
	return "SyncUsingMod"
}

// Convert s to go object which only contains go-built-in types
func (s *SyncUsingMod) MakeGo() (res any) {
	if s == nil {
		return nil
	}
	return []any(*s)
}

// Sync data to s from obj
func (s *SyncUsingMod) FromGo(obj any) error {
	object, success := obj.([]any)
	if !success {
		return fmt.Errorf("FromGo: Failed to convert obj to []interface{}; obj = %#v", obj)
	}
	*s = object
	// convert and sync data
	return nil
	// return
}
