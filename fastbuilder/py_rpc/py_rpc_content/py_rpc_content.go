package py_rpc_content

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/fastbuilder/py_rpc/py_rpc_parser"
)

// Convert content to go object which
// only contains go-built-in types
func Marshal(content PyRpcContent) (res any) {
	return []any{
		content.Name(),
		content.MakeGo(),
		nil,
	}
}

// Convert obj to PyRpcContent
func Unmarshal(obj any) (content PyRpcContent, err error) {
	object, success := obj.([]any)
	if !success {
		return nil, fmt.Errorf("Unmarshal: Unsupported PyRpc packet; obj = %#v", obj)
	}
	if len(object) < 2 {
		return nil, fmt.Errorf("Unmarshal: The length of object is less than 2; object = %#v", object)
	}
	// check data
	name, success := object[0].(string)
	if !success {
		return nil, fmt.Errorf("Unmarshal: Failed to convert object[0] to string; object[0] = %#v", object[0])
	}
	// get name
	var ok bool
	content, ok = Pool()[name]
	if !ok {
		content = &Default{NAME: name}
	}
	// get content of initial state
	content.FromGo(object[1])
	return
	// put data and return
}

// Package content to py_rpc_parser.PyRpcObject
func PackageContent(content PyRpcContent) py_rpc_parser.PyRpcObject {
	return py_rpc_parser.FromGo(Marshal(content))
}

// Output the object as json
func Debug(object any) (res string, err error) {
	str, err := json.MarshalIndent(object, "", "\t")
	if err != nil {
		err = fmt.Errorf("Debug: %v", err)
	}
	return string(str), err
}
