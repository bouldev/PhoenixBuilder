package py_rpc

import "fmt"

// 心跳包
type S2CHeartBeat []SingleS2CHeartBeat
type SingleS2CHeartBeat struct {
	Is64      any   `json:"is_64"`      // e.g. interface {}(nil)
	IsAndroid any   `json:"is_android"` // e.g. interface {}(nil)
	T         int64 `json:"t"`          // e.g. int64(1709452338)
}

// Return the name of s
func (s *S2CHeartBeat) Name() string {
	return "S2CHeartBeat"
}

// Convert s to go object which only contains go-built-in types
func (s *S2CHeartBeat) MakeGo() (res any) {
	if s == nil {
		return nil
	}
	// check
	result := make([]any, len(*s))
	for key, value := range *s {
		result[key] = map[string]any{
			"is_64":      value.Is64,
			"is_android": value.IsAndroid,
			"t":          value.T,
		}
	}
	// sync data
	return result
	// return
}

// Sync data to s from obj
func (s *S2CHeartBeat) FromGo(obj any) error {
	object, success := obj.([]any)
	if !success {
		return fmt.Errorf("FromGo: Failed to convert obj to []interface{}; obj = %#v", obj)
	}
	// convert data
	*s = make(S2CHeartBeat, len(object))
	for key, value := range object {
		val, success := value.(map[string]any)
		if !success {
			return fmt.Errorf("FromGo: Failed to convert value to map[string]interface{}; value = %#v", value)
		}
		// convert data
		T, success := val["t"].(int64)
		if !success {
			return fmt.Errorf(`FromGo: Failed to convert val["t"] to int64; val["t"] = %#v`, val["t"])
		}
		// get T
		(*s)[key] = SingleS2CHeartBeat{
			Is64:      val["is_64"],
			IsAndroid: val["is_android"],
			T:         T,
		}
		// submit result
	}
	// make object and sync data
	return nil
	// return
}
