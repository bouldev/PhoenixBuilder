package pet

import "fmt"

type CheckServerContainPet struct{}

// Return the event name of c
func (c *CheckServerContainPet) EventName() string {
	return "check_server_contain_pet"
}

// Convert c to go object which only contains go-built-in types
func (c *CheckServerContainPet) MakeGo() (res any) {
	return map[string]any{}
}

// Sync data to c from obj
func (c *CheckServerContainPet) FromGo(obj any) error {
	object, success := obj.(map[string]any)
	if !success {
		return fmt.Errorf("FromGo: Failed to convert obj to map[string]interface{}; obj = %#v", obj)
	}
	// convert data
	if len(object) != 0 {
		return fmt.Errorf("FromGo: The length of object is not equal to 0; object = %#v", object)
	}
	// check data
	*c = CheckServerContainPet{}
	// sync data
	return nil
	// return
}
