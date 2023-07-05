package py_rpc

import (
	"encoding/binary"
	"bytes"
)

type PyRpcListObject struct {
	Value []PyRpcObject
}

func (o *PyRpcListObject) Marshal() []byte {
	var header []byte
	if(len(o.Value)<=0xf) {
		header=[]byte{0x90+byte(len(o.Value))}
	}else if(len(o.Value)<=0xffff) {
		header=[]byte{0xdc,0,0}
		binary.BigEndian.PutUint16(header[1:], uint16(len(o.Value)))
	}else if(uint(len(o.Value))<=uint(0xffffffff)) {
		header=make([]byte, 5)
		header[0]=0xdd
		binary.BigEndian.PutUint32(header[1:], uint32(len(o.Value)))
	}else{
		panic("PyRpcListObject/Marshal: List length exceeded maximum length")
	}
	arr:=make([][]byte,1+len(o.Value))
	arr[0]=header
	for i:=0;i<len(o.Value);i++ {
		arr[i+1]=o.Value[i].Marshal()
	}
	return bytes.Join(arr, []byte{})
}

func (o *PyRpcListObject) Parse(v []byte) uint {
	consumed:=uint(0)
	var nokori []byte
	if(v[0]==0x90) {
		o.Value=[]PyRpcObject{}
		return 1
	}else if(0x90<v[0]&&v[0]<=0x9f) {
		o.Value=make([]PyRpcObject, v[0]-0x90)
		nokori=v[1:]
		consumed++
	}else if(v[0]==0xdc) {
		len:=binary.BigEndian.Uint16(v[1:])
		o.Value=make([]PyRpcObject, len)
		nokori=v[3:]
		consumed+=3
	}else if(v[0]==0xdd) {
		len:=binary.BigEndian.Uint32(v[1:])
		o.Value=make([]PyRpcObject, len)
		nokori=v[5:]
		consumed+=5
	}else{
		panic("PyRpcListObject/Parse: Not even a list object")
	}
	for i:=0;i<len(o.Value);i++ {
		dv, dl:=unmarshal(nokori)
		o.Value[i]=dv
		consumed+=dl
		nokori=nokori[dl:]
	}
	return consumed
}

func (_ *PyRpcListObject) Type() uint {
	return ListType
}

func (o *PyRpcListObject) MakeGo() interface{} {
	ret:=make([]interface{}, len(o.Value))
	for k, v:=range o.Value {
		ret[k]=v.MakeGo()
	}
	return ret
}

func (o *PyRpcListObject) FromGo(v interface{}) {
	pv:=v.([]interface{})
	o.Value=make([]PyRpcObject, len(pv))
	for k, value:=range pv {
		o.Value[k]=FromGo(value)
	}
}