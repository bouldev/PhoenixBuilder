package py_rpc

import "encoding/binary"

type PyRpcBytesObject struct {
	Value []byte
}

func (o *PyRpcBytesObject) Marshal() []byte {
	var header []byte
	if len(o.Value)<=0xff {
		header=[]byte{0xc4, byte(len(o.Value))}
	}else if(len(o.Value)<=0xffff) {
		header=[]byte{0xc5,0x00,0x00}
		binary.BigEndian.PutUint16(header[1:], uint16(len(o.Value)))
	}else if(uint(len(o.Value))<=uint(0xffffffff)) {
		header=make([]byte, 1+4)
		header[0]=0xc6
		binary.BigEndian.PutUint32(header[1:], uint32(len(o.Value)))
	}else{
		panic("PyRpcBytesObject: Maximum allowed size exceeded")
	}
	return append(header, o.Value...)
}

func (o *PyRpcBytesObject) Parse(v []byte) uint {
	var size uint32
	var header_consumed uint8
	var nokori []byte
	if v[0]==0xc4 {
		size=uint32(v[1])
		nokori=v[2:]
		header_consumed=2
	}else if v[0]==0xc5 {
		size=uint32(binary.BigEndian.Uint16(v[1:]))
		nokori=v[3:]
		header_consumed=3
	}else if v[0]==0xc6 {
		size=uint32(binary.BigEndian.Uint32(v[1:]))
		nokori=v[5:]
		header_consumed=1+4
	}else{
		panic("PyRpcBytesObject/Parse: Not even a bytes object")
	}
	if len(nokori)<int(size) {
		panic("PyRpcBytesObject/Parse: Early EOF")
	}
	o.Value=nokori[0:size]
	return uint(size+uint32(header_consumed))
}

func (_ *PyRpcBytesObject) Type() uint {
	return BytesType
}

func (o *PyRpcBytesObject) MakeGo() interface{} {
	return string(o.Value)
}

func (o *PyRpcBytesObject) FromGo(v interface{}) {
	pv:=v.([]byte)
	o.Value=pv
}