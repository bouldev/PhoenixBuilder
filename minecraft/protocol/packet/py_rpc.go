package packet

import (
	"fmt"
	"phoenixbuilder/minecraft/protocol"
)

type PyRpc struct {
	Value any
	Error error
}

// ID ...
func (*PyRpc) ID() uint32 {
	return IDPyRpc
}

// Marshal ...
func (pk *PyRpc) Marshal(w *protocol.Writer) {
	content, err := protocol.MarshalMsgpack(pk.Value)
	if err != nil {
		pk.Error = fmt.Errorf("Marshal: %v", err)
		panic(pk.Error)
	}
	// marshal pk.Value to content
	w.ByteSlice(&content)
	w.Bytes(&[]byte{0xae, 0x23, 0xdb, 0x05})
	// write content with magic bytes to the writer
}

// Unmarshal ...
func (pk *PyRpc) Unmarshal(r *protocol.Reader) {
	var content []byte
	var err error
	r.ByteSlice(&content)
	// prepare
	pk.Value, err = protocol.UnmarshalMsgpack(content)
	if err != nil {
		panic(fmt.Sprintf("Unmarshal: %v", err))
	}
	// unmarshal content
	value, success := pk.Value.([]any)
	if !success {
		return
	}
	pk.Value, err = protocol.FormatSliceInMsgpack(value)
	if err != nil {
		pk.Error = fmt.Errorf("Unmarshal: %v", err)
		panic(pk.Error)
	}
	// format the decoded data
}
