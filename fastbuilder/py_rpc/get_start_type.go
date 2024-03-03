package py_rpc

import "fmt"

// 机器人进服后的挑战之一
type GetStartType []SingleGetStartType
type SingleGetStartType struct {
	Content string
}

// Return the name of g
func (g *GetStartType) Name() string {
	return "GetStartType"
}

// Convert g to go object which only contains go-built-in types
func (g *GetStartType) MakeGo() (res any) {
	if g == nil {
		return nil
	}
	// check
	result := make([]any, len(*g))
	for key, value := range *g {
		result[key] = []any{value.Content}
	}
	// sync data
	return result
	// return
}

// Sync data to g from obj
func (g *GetStartType) FromGo(obj any) error {
	object, success := obj.([]any)
	if !success {
		return fmt.Errorf("FromGo: Failed to convert obj to []interface{}; obj = %#v", obj)
	}
	// convert data
	*g = make(GetStartType, len(object))
	for key, value := range object {
		val, success := value.(string)
		if !success {
			return fmt.Errorf("FromGo: Failed to convert value to string; value = %#v", value)
		}
		// convert data
		(*g)[key] = SingleGetStartType{Content: val}
		// submit result
	}
	// make object and sync data
	return nil
	// return
}
