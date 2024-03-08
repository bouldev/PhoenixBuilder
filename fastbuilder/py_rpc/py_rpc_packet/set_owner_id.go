package py_rpc

import "fmt"

// 设置机器人的 Netease User ID
type SetOwnerId []SingleSetOwnerId
type SingleSetOwnerId struct {
	OwnerID int64
}

// Return the name of s
func (s *SetOwnerId) Name() string {
	return "SetOwnerId"
}

// Convert s to go object which only contains go-built-in types
func (s *SetOwnerId) MakeGo() (res any) {
	if s == nil {
		return nil
	}
	// check
	result := make([]any, len(*s))
	for key, value := range *s {
		result[key] = []any{value.OwnerID}
	}
	// sync data
	return result
	// return
}

// Sync data to s from obj
func (s *SetOwnerId) FromGo(obj any) error {
	object, success := obj.([]any)
	if !success {
		return fmt.Errorf("FromGo: Failed to convert obj to []interface{}; obj = %#v", obj)
	}
	// convert data
	*s = make(SetOwnerId, len(object))
	for key, value := range object {
		val, success := value.(int64)
		if !success {
			return fmt.Errorf("FromGo: Failed to convert value to int64; value = %#v", value)
		}
		// convert data
		(*s)[key] = SingleSetOwnerId{OwnerID: val}
		// submit result
	}
	// make object and sync data
	return nil
	// return
}
