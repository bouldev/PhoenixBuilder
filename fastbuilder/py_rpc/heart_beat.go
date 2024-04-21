package py_rpc

import "fmt"

// Describe the type of PyRpc/HeartBeat packet
const (
	ServerToClientHeartBeat = "S2CHeartBeat"
	ClientToServerHeartBeat = "C2SHeartBeat"
)

// 心跳包
type HeartBeat struct {
	Content HeartBeatContent
	Type    string
}

type HeartBeatContent struct {
	Is64      any    `json:"is_64"`      // e.g. interface {}(nil)
	IsAndroid any    `json:"is_android"` // e.g. interface {}(nil)
	T         uint64 `json:"t"`          // e.g. uint64(1709452338)
}

// Return the name of s
func (h *HeartBeat) Name() string {
	return h.Type
}

// Convert s to go object which only contains go-built-in types
func (h *HeartBeat) MakeGo() (res any) {
	return []any{
		map[string]any{
			"is_64":      h.Content.Is64,
			"is_android": h.Content.IsAndroid,
			"t":          h.Content.T,
		},
	}
}

// Sync data to s from obj
func (h *HeartBeat) FromGo(obj any) error {
	object, success := obj.([]any)
	if !success {
		return fmt.Errorf("FromGo: Failed to convert obj to []interface{}; obj = %#v", obj)
	}
	if len(object) != 1 {
		return fmt.Errorf("FromGo: The length of object is not equal to 1; object = %#v", object)
	}
	// convert and check data
	content, success := object[0].(map[string]any)
	if !success {
		return fmt.Errorf("FromGo: Failed to convert object[0] to map[string]interface{}; object[0] = %#v", object[0])
	}
	T, success := content["t"].(uint64)
	if !success {
		return fmt.Errorf(`FromGo: Failed to convert content["t"] to uint64; content["t"] = %#v`, content["t"])
	}
	// get data
	h.Content = HeartBeatContent{
		Is64:      content["is_64"],
		IsAndroid: content["is_android"],
		T:         T,
	}
	// sync data
	return nil
	// return
}
