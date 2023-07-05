package py_rpc

import "encoding/binary"

type PyRpcIntObject struct {
	Value int64
}

func (o *PyRpcIntObject) Marshal() []byte {
	if(0<=o.Value&&o.Value<0x80) {
		return []byte{byte(o.Value)}
	}else if(-0x20<=o.Value&&o.Value<0) {
		return []byte{byte(int8(o.Value))}
	}else if(0x80<=o.Value&&o.Value<=0xff) {
		return []byte{0xcc, byte(o.Value)}
	}else if(-0x80<=o.Value&&o.Value<0) {
		return []byte{0xd0, byte(int8(o.Value))}
	}else if(0xff<o.Value&&o.Value<=0xffff) {
		buf:=make([]byte, 3)
		buf[0]=0xcd
		binary.BigEndian.PutUint16(buf[1:], uint16(o.Value))
		return buf
	}else if(-0x8000<=o.Value&&o.Value< -0x80) {
		buf:=make([]byte, 3)
		buf[0]=0xd1
		binary.BigEndian.PutUint16(buf[1:], uint16(int16(o.Value)))
		return buf
	}else if(0xffff<o.Value&&o.Value<=0xffffffff) {
		buf:=make([]byte, 1+4)
		buf[0]=0xce
		binary.BigEndian.PutUint32(buf[1:], uint32(o.Value))
		return buf
	}else if(-0x80000000<=o.Value&&o.Value< -0x8000) {
		buf:=make([]byte, 1+4)
		buf[0]=0xd2
		binary.BigEndian.PutUint32(buf[1:], uint32(int32(o.Value)))
		return buf
	}else if(0xffffffff<o.Value&&o.Value<=/*0xffffffffffffffff*/0x7FFFFFFFFFFFFFFF) {
		buf:=make([]byte, 1+8)
		buf[0]=0xcf
		binary.BigEndian.PutUint64(buf[1:], uint64(o.Value))
		return buf
	}else if(-0x8000000000000000<=o.Value&&o.Value< -0x80000000) {
		buf:=make([]byte, 1+8)
		buf[0]=0xd3
		binary.BigEndian.PutUint64(buf[1:], uint64(int64(o.Value)))
		return buf
	}
	panic("PyRpcIntObject: Invalid value")
}

func (o *PyRpcIntObject) Parse(v []byte) uint {
	if(0<=v[0]&&v[0]<0x80) {
		o.Value=int64(v[0])
		return 1
	}else if(-0x20<=int8(v[0])&&int8(v[0])<0) {
		o.Value=int64(int8(v[0]))
		return 1
	}else if(v[0]==0xcc) {
		o.Value=int64(v[1])
		return 2
	}else if(v[0]==0xd0) {
		o.Value=int64(int8(v[1]))
		return 2
	}else if(v[0]==0xcd) {
		o.Value=int64(binary.BigEndian.Uint16(v[1:]))
		return 3
	}else if(v[0]==0xd1) {
		o.Value=int64(int16(binary.BigEndian.Uint16(v[1:])))
		return 3
	}else if(v[0]==0xce) {
		o.Value=int64(binary.BigEndian.Uint32(v[1:]))
		return 1+4
	}else if(v[0]==0xd2) {
		o.Value=int64(int32(binary.BigEndian.Uint32(v[1:])))
		return 1+4
	}else if(v[0]==0xcf) {
		// Buggy but I don't think such support is even necessary
		o.Value=int64(binary.BigEndian.Uint64(v[1:]))
		return 1+8
	}else if(v[0]==0xd3) {
		o.Value=int64(binary.BigEndian.Uint64(v[1:]))
		return 1+8
	}
	panic("PyRpcIntObject/Parse: Invalid value")
}

func (_ *PyRpcIntObject) Type() uint {
	return IntType
}

func (o *PyRpcIntObject) MakeGo() interface{} {
	return o.Value
}

func (o *PyRpcIntObject) FromGo(v interface{}) {
	pv:=v.(int64)
	o.Value=pv
}