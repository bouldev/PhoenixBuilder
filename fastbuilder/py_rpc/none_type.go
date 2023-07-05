package py_rpc

type PyRpcNoneObject struct {
}

func (_ *PyRpcNoneObject) Marshal() []byte {
	return []byte{0xc0}
}

func (_ *PyRpcNoneObject) Parse(_ []byte) uint {
	return 1
}

func (_ *PyRpcNoneObject) Type() uint {
	return NoneType
}

func (_ *PyRpcNoneObject) MakeGo() interface{} {
	return nil
}

func (_ *PyRpcNoneObject) FromGo(_ interface{}) {
}