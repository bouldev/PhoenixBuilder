package py_rpc

import (
	"encoding/binary"
	"bytes"
)

type PyRpcMapObject struct{
	Value map[PyRpcObject]PyRpcObject
}

func (o *PyRpcMapObject) Marshal() []byte {
	var header []byte
	if(len(o.Value)<=0xf) {
		header=[]byte{0x80+byte(len(o.Value))}
	}else if(len(o.Value)<=0xffff) {
		header=[]byte{0xde,0,0}
		binary.BigEndian.PutUint16(header[1:], uint16(len(o.Value)))
	}else if(uint(len(o.Value))<=uint(0xffffffff)) {
		header=make([]byte, 5)
		header[0]=0xdf
		binary.BigEndian.PutUint32(header[1:], uint32(len(o.Value)))
	}else{
		panic("PyRpcMapObject/Marshal: Map length exceeded maximum length")
	}
	arr:=make([][]byte,1+(len(o.Value)*2))
	arr[0]=header
	i:=0
	for key, value:=range o.Value {
		arr[i+1]=key.Marshal()
		arr[i+2]=value.Marshal()
		i+=2
	}
	return bytes.Join(arr, []byte{})
}

func (o *PyRpcMapObject) Parse(v []byte) uint {
	consumed:=uint(0)
	var nokori []byte
	o.Value=map[PyRpcObject]PyRpcObject{}
	entries:=0
	if(v[0]==0x80) {
		return 1
	}else if(0x80<v[0]&&v[0]<=0x8f) {
		entries=int(v[0]-0x80)
		nokori=v[1:]
		consumed++
	}else if(v[0]==0xde) {
		_len:=binary.BigEndian.Uint16(v[1:])
		entries=int(_len)
		nokori=v[3:]
		consumed+=3
	}else if(v[0]==0xdf) {
		_len:=binary.BigEndian.Uint32(v[1:])
		entries=int(_len)
		nokori=v[5:]
		consumed+=5
	}else{
		panic("PyRpcMapObject/Parse: Not even a map object")
	}
	for i:=0;i<entries;i++ {
		dk, dl1:=unmarshal(nokori)
		nokori=nokori[dl1:]
		dv, dl2:=unmarshal(nokori)
		o.Value[dk]=dv
		consumed+=(dl1+dl2)
		nokori=nokori[dl2:]
	}
	return consumed
}

func (_ *PyRpcMapObject) Type() uint {
	return MapType
}

func (o *PyRpcMapObject) MakeGo() interface{} {
	ret:=map[interface{}]interface{} {}
	marshalable_ret:=map[string]interface{} {}
	marshalable_ret2:=map[int64]interface{} {}
	for k, v:=range o.Value {
		k_g:=k.MakeGo()
		k_b, is_b:=k_g.(string)
		if is_b {
			marshalable_ret[string(k_b)]=v.MakeGo()
		}else{
			k_i, is_i:=k_g.(int64)
			if is_i {
				marshalable_ret2[k_i]=v.MakeGo()
			}else{
				ret[k_g]=v.MakeGo()
			}
		}
	}
	if len(marshalable_ret)!=0 {
		return marshalable_ret
	}else if(len(marshalable_ret2)!=0) {
		return marshalable_ret2
	}else if(len(ret)==0) {
		return map[string]interface{} {}
	}
	return ret
}

func (o *PyRpcMapObject) FromGo(v interface{}) {
	pv1, pv1succ:=v.(map[string]interface{})
	if pv1succ {
		o.Value=map[PyRpcObject]PyRpcObject {}
		for k, value:=range pv1 {
			o.Value[FromGo(k)]=FromGo(value)
		}
		return
	}
	pv2, pv2succ:=v.(map[int64]interface{})
	if pv2succ {
		o.Value=map[PyRpcObject]PyRpcObject {}
		for k, value:=range pv2 {
			o.Value[FromGo(k)]=FromGo(value)
		}
		return
	}
	pv:=v.(map[interface{}]interface{})
	o.Value=map[PyRpcObject]PyRpcObject {}
	for k, value:=range pv {
		o.Value[FromGo(k)]=FromGo(value)
	}
}