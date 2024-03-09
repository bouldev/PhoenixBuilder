package py_rpc_content

// Express a content which contains in a PyRpc packet
type PyRpcContent interface {
	// Return the name of this content which included on PyRpc packet
	Name() string
	// Convert the target content to go object which only contains go-built-in types
	MakeGo() (res any)
	// Sync data to the target content from obj
	FromGo(obj any) error
}
