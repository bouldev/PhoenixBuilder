package achievement

import "fmt"

type InitInformation struct {
	AllNodeInformation map[string]any `json:"AllNodeInformation"` // e.g. map[string]interface {}{}
	NodeInformation    map[string]any `json:"NodeInformation"`    // e.g. map[string]interface {}{}
	Parent             map[string]any `json:"Parent"`             // e.g. map[string]interface {}{}
	PlayerNodeProgress map[string]any `json:"PlayerNodeProgress"` // e.g. map[string]interface {}{}
	RootNode           []any          `json:"RootNode"`           // e.g. []interface {}
}

// Return the event name of i
func (i *InitInformation) EventName() string {
	return "InitInformation"
}

// Convert i to go object which only contains go-built-in types
func (i *InitInformation) MakeGo() (res any) {
	return map[string]any{
		"AllNodeInformation": i.AllNodeInformation,
		"NodeInformation":    i.NodeInformation,
		"Parent":             i.Parent,
		"PlayerNodeProgress": i.PlayerNodeProgress,
		"RootNode":           i.RootNode,
	}
}

// Sync data to i from obj
func (i *InitInformation) FromGo(obj any) error {
	object, success := obj.(map[string]any)
	if !success {
		return fmt.Errorf("FromGo: Failed to convert obj to map[string]interface{}; obj = %#v", obj)
	}
	// convert data
	all_node_information, success := object["AllNodeInformation"].(map[string]any)
	if !success {
		return fmt.Errorf(`FromGo: Failed to convert object["AllNodeInformation"] to map[string]interface{}; object["AllNodeInformation"] = %#v`, object["AllNodeInformation"])
	}
	node_information, success := object["NodeInformation"].(map[string]any)
	if !success {
		return fmt.Errorf(`FromGo: Failed to convert object["NodeInformation"] to map[string]interface{}; object["NodeInformation"] = %#v`, object["NodeInformation"])
	}
	parent, success := object["Parent"].(map[string]any)
	if !success {
		return fmt.Errorf(`FromGo: Failed to convert object["Parent"] to map[string]interface{}; object["Parent"] = %#v`, object["Parent"])
	}
	player_node_progress, success := object["PlayerNodeProgress"].(map[string]any)
	if !success {
		return fmt.Errorf(`FromGo: Failed to convert object["PlayerNodeProgress"] to map[string]interface{}; object["PlayerNodeProgress"] = %#v`, object["PlayerNodeProgress"])
	}
	root_node, success := object["RootNode"].([]any)
	if !success {
		return fmt.Errorf(`FromGo: Failed to convert object["RootNode"] to []interface{}; object["RootNode"] = %#v`, object["RootNode"])
	}
	// get data
	*i = InitInformation{
		AllNodeInformation: all_node_information,
		NodeInformation:    node_information,
		Parent:             parent,
		PlayerNodeProgress: player_node_progress,
		RootNode:           root_node,
	}
	// sync data
	return nil
	// return
}
