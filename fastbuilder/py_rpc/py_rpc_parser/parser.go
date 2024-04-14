package py_rpc_parser

import (
	"github.com/ugorji/go/codec"
)

// This is the Netease PyRpcPacket parser.
// Python lib, msgpack
type PyRpcObject struct {
	Value any
}

var (
	mh codec.MsgpackHandle
)

func init() {
	mh.RawToString = true
	mh.StringToRaw = true
}

func FromGo(v any) PyRpcObject {
	return PyRpcObject{Value: v}
}

func Unmarshal(val []byte) (retval PyRpcObject) {
	codec.NewDecoderBytes(val, &mh).Decode(&retval.Value)
	return
}

func (p *PyRpcObject) Marshal() (retval []byte) {
	codec.NewEncoderBytes(&retval, &mh).Encode(p.Value)
	return
}
