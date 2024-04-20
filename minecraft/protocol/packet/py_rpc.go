package packet

import (
	//"bytes"
	//"fmt"
	"phoenixbuilder/minecraft/protocol"

	"github.com/ugorji/go/codec"
)

var msgPackHandler codec.MsgpackHandle

func init() {
	msgPackHandler.RawToString = true
	msgPackHandler.StringToRaw = true
}

type PyRpc struct {
	Value any
}

// ID ...
func (*PyRpc) ID() uint32 {
	return IDPyRpc
}

func (pk *PyRpc) goValueToMsgPackBytes() (outBytes []byte) {
	var msgPackHandler codec.MsgpackHandle
	msgPackHandler.RawToString = true
	msgPackHandler.StringToRaw = true
	codec.NewEncoderBytes(&outBytes, &msgPackHandler).Encode(pk.Value)
	return outBytes
}
func (pk *PyRpc) goValueFromMsgPackBytes(inBytes []byte) {
	var msgPackHandler codec.MsgpackHandle
	msgPackHandler.RawToString = true
	msgPackHandler.StringToRaw = true
	codec.NewDecoderBytes(inBytes, &msgPackHandler).Decode(&pk.Value)
}

// Marshal ...
func (pk *PyRpc) Marshal(w *protocol.Writer) {
	//w.ByteSlice(&pk.Content)
	content := pk.goValueToMsgPackBytes()
	w.ByteSlice(&content)
	w.Bytes(&[]byte{0xae, 0x23, 0xdb, 0x05})
	//fmt.Printf("%d\n",len(pk.Content))
	//fmt.Printf("%X\n",buf.Bytes())
	//_ = protocol.WriteByteSlice(buf, pk.Content)
	//_ = protocol.WriteVaruint32(buf, pk.DisallowBatching)
	//buf.Write([]byte{0xae,0x23,0xdb,0x05})
	//fmt.Printf("%X\n\n",buf.Bytes())
	//var outuint32 uint32
	//protocol.Varuint32(bytes.NewBuffer([]byte{0xae,0x23,0xdb,0x05}),&outuint32)
	//fmt.Printf("%d\n",outuint32)

}

// Unmarshal ...
func (pk *PyRpc) Unmarshal(r *protocol.Reader) {
	var content []byte
	r.ByteSlice(&content)
	pk.goValueFromMsgPackBytes(content)
	//r.ByteSlice(&pk.Content)
	/*var bt byte
	var bt2 byte
	var bt3 byte
	var bt4 byte
	r.Uint8(&bt)
	r.Uint8(&bt2)
	r.Uint8(&bt3)
	r.Uint8(&bt4)
	fmt.Printf("Addi:%X %X %X %X %X\n", bt, bt2, bt3, bt4)*/
}
