package packet

import (
	"phoenixbuilder/minecraft/protocol"

	// A Python library which named "msgpack"
	"github.com/ugorji/go/codec"
)

type PyRpc struct {
	Value any
}

// ID ...
func (*PyRpc) ID() uint32 {
	return IDPyRpc
}

func (pk *PyRpc) goValueToMsgPackBytes() (outBytes []byte) {
	var msgPackHandler codec.MsgpackHandle
	codec.NewEncoderBytes(&outBytes, &msgPackHandler).Encode(pk.Value)
	return outBytes
}
func (pk *PyRpc) goValueFromMsgPackBytes(inBytes []byte) {
	var msgPackHandler codec.MsgpackHandle
	msgPackHandler.RawToString = true
	codec.NewDecoderBytes(inBytes, &msgPackHandler).Decode(&pk.Value)
}

// Marshal ...
func (pk *PyRpc) Marshal(w *protocol.Writer) {
	content := pk.goValueToMsgPackBytes()
	w.ByteSlice(&content)
	w.Bytes(&[]byte{0xae, 0x23, 0xdb, 0x05})
}

// Unmarshal ...
func (pk *PyRpc) Unmarshal(r *protocol.Reader) {
	var content []byte
	r.ByteSlice(&content)
	pk.goValueFromMsgPackBytes(content)
}
