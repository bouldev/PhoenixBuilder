package py_rpc

import (
	"fmt"
	"reflect"
)

// This is the Netease PyRpcPacket parser.

type PyRpcObject interface {
	Marshal() []byte
	Parse([]byte) uint
	
	Type() uint
	
	MakeGo() interface{}
	FromGo(interface{})
}

func unmarshal(val []byte) (PyRpcObject, uint) {
	if(val[0]==0xc0) {
		return &PyRpcNoneObject{}, 1
	}else if(val[0]==0xc2) {
		return &PyRpcBoolObject{Value:false}, 1
	}else if(val[0]==0xc3) {
		return &PyRpcBoolObject{Value:true}, 1
	}else if((0<=val[0]&&val[0]<0x80)||(-0x20<=int8(val[0])&&int8(val[0])<0)||(0xcc<=val[0]&&val[0]<=0xcf)||(0xd0<=val[0]&&val[0]<=0xd3)) {
		o:=&PyRpcIntObject{}
		l:=o.Parse(val)
		return o,l
	}else if(0xc4<=val[0]&&val[0]<=0xc6) {
		o:=&PyRpcBytesObject{}
		l:=o.Parse(val)
		return o,l
	}else if((0x90<=val[0]&&val[0]<=0x9f)||0xdc==val[0]||0xdd==val[0]) {
		o:=&PyRpcListObject{}
		l:=o.Parse(val)
		return o,l
	}else if((0x80<=val[0]&&val[0]<=0x8f)||0xde==val[0]||0xdf==val[0]) {
		o:=&PyRpcMapObject{}
		l:=o.Parse(val)
		return o,l
	}else{
		panic(fmt.Errorf("py_rpc.unmarshal: Unrecognized type %X", val[0]))
	}
}

func Unmarshal(val []byte) PyRpcObject {
	obj, _:=unmarshal(val)
	return obj
}

func FromGo(v interface{}) PyRpcObject {
	_type:=reflect.TypeOf(v)
	if(_type==nil) {
		return &PyRpcNoneObject{}
	}
	kind:=_type.Kind()
	if(kind==reflect.Bool) {
		return &PyRpcBoolObject{Value:v.(bool)}
	}else if(kind==reflect.Int) {
		return &PyRpcIntObject{Value:int64(v.(int))}
	}else if(kind==reflect.Int8) {
		return &PyRpcIntObject{Value:int64(v.(int8))}
	}else if(kind==reflect.Int16) {
		return &PyRpcIntObject{Value:int64(v.(int16))}
	}else if(kind==reflect.Int32) {
		return &PyRpcIntObject{Value:int64(v.(int32))}
	}else if(kind==reflect.Int64) {
		return &PyRpcIntObject{Value:v.(int64)}
	}else if(kind==reflect.Uint) {
		return &PyRpcIntObject{Value:int64(v.(uint))}
	}else if(kind==reflect.Uint8) {
		return &PyRpcIntObject{Value:int64(v.(uint8))}
	}else if(kind==reflect.Uint16) {
		return &PyRpcIntObject{Value:int64(v.(uint16))}
	}else if(kind==reflect.Uint32) {
		return &PyRpcIntObject{Value:int64(v.(uint32))}
	}else if(kind==reflect.Uint64) {
		return &PyRpcIntObject{Value:int64(v.(uint64))}
	}else if(kind==reflect.Float32) {
		return &PyRpcIntObject{Value:int64(v.(float32))}
	}else if(kind==reflect.Float64) {
		return &PyRpcIntObject{Value:int64(v.(float64))}
	}else if(kind==reflect.Map) {
		o:=&PyRpcMapObject{}
		o.FromGo(v)
		return o
	}else if(kind==reflect.Slice) {
		elem_kind:=_type.Elem().Kind()
		if(elem_kind==reflect.Uint8) {
			return &PyRpcBytesObject{Value:v.([]byte)}
		}
		o:=&PyRpcListObject{}
		o.FromGo(v)
		return o
	}else if(kind==reflect.String) {
		return &PyRpcBytesObject{Value:[]byte(v.(string))}
	}else{
		panic(fmt.Errorf("py_rpc.FromGo: Unsupported type %T", v))
	}
}