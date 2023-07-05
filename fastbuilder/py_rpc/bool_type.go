package py_rpc

type PyRpcBoolObject struct {
	Value bool
}

func (o *PyRpcBoolObject) Marshal() []byte {
	if o.Value {
		return []byte{0xc3}
	}
	return []byte{0xc2}
}

func (o *PyRpcBoolObject) Parse(v []byte) uint {
	if v[0]==0xc2 {
		o.Value=false
	}else{
		o.Value=true
	}
	return 1
}

func (_ *PyRpcBoolObject) Type() uint {
	return BoolType
}

func (o *PyRpcBoolObject) MakeGo() interface{} {
	return o.Value
}

func (o *PyRpcBoolObject) FromGo(v interface{}) {
	pv:=v.(bool)
	o.Value=pv
}