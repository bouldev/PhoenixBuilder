package py_rpc

import "fmt"

// 设置机器人的 Netease User ID
type SetOwnerId struct {
	OwnerID uint64
}

// Return the name of s
func (s *SetOwnerId) Name() string {
	return "SetOwnerId"
}

// Convert s to go object which only contains go-built-in types
func (s *SetOwnerId) MakeGo() (res any) {
	return []any{s.OwnerID}
}

// Sync data to s from obj
func (s *SetOwnerId) FromGo(obj any) error {
	object, success := obj.([]any)
	if !success {
		return fmt.Errorf("FromGo: Failed to convert obj to []interface{}; obj = %#v", obj)
	}
	if len(object) != 1 {
		return fmt.Errorf("FromGo: The length of object is not equal to 1; object = %#v", object)
	}
	// convert and check data
	s.OwnerID, success = object[0].(uint64)
	if !success {
		return fmt.Errorf("FromGo: Failed to convert object[0] to uint64; object[0] = %#v", object[0])
	}
	// get and sync data
	return nil
	// return
}
