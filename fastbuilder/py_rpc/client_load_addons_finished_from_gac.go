package py_rpc

import "fmt"

// Used on initialize the
// Minecraft connection
type ClientLoadAddonsFinishedFromGac []any

// Return the name of s
func (s *ClientLoadAddonsFinishedFromGac) Name() string {
	return "ClientLoadAddonsFinishedFromGac"
}

// Convert s to go object which only contains go-built-in types
func (s *ClientLoadAddonsFinishedFromGac) MakeGo() (res any) {
	if s == nil {
		return nil
	}
	return []any(*s)
}

// Sync data to s from obj
func (s *ClientLoadAddonsFinishedFromGac) FromGo(obj any) error {
	object, success := obj.([]any)
	if !success {
		return fmt.Errorf("FromGo: Failed to convert obj to []interface{}; obj = %#v", obj)
	}
	*s = object
	// convert and sync data
	return nil
	// return
}
