package py_rpc

// Express the content which contains in a PyRpc packet
type PyRpc interface {
	// Return the name of this content
	Name() string
	// Convert the target content to go object which only contains go-built-in types
	MakeGo() (res any)
	// Sync data to the target content from obj
	FromGo(obj any) error
}
