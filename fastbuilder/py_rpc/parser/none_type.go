package py_rpc_parser

type PyRpcNoneObject struct {
}

func (*PyRpcNoneObject) Marshal() []byte {
	return []byte{0xc0}
}

func (*PyRpcNoneObject) Parse(_ []byte) uint {
	return 1
}

func (*PyRpcNoneObject) Type() uint {
	return NoneType
}

func (*PyRpcNoneObject) MakeGo() interface{} {
	return nil
}

func (*PyRpcNoneObject) FromGo(_ interface{}) {
}
