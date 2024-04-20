package py_rpc

// Default content, which used to
// describe the unsupported parts
type Default struct {
	NAME string // The name of this content
	Data any    // The data contained in this content
}

// Return the name of d
func (d *Default) Name() string {
	return d.NAME
}

// Convert d to go object which only contains go-built-in types
func (d *Default) MakeGo() (res any) {
	return d.Data
}

// Sync data to d from obj
func (d *Default) FromGo(res any) error {
	d.Data = res
	return nil
}
