package plugin_structs

type Function struct {
	Name string
	OwnedKeywords []string
	
	FunctionType byte
	SFMinSliceLen uint16
	SFArgumentTypes []byte
	FunctionContent interface{} // Regular/Simple: func(interface{}(~~*minecraft.Conn~~),interface{})
				    // Continue: map[string]*FunctionChainItem
}

type FunctionChainItem struct {
	FunctionType byte
	ArgumentTypes []byte
	Content interface{}
}

const (
	FunctionTypeSimple    = 0 // End of simple chain
	FunctionTypeContinue  = 1 // Simple chain
	FunctionTypeRegular   = 2
)

const (
	SimpleFunctionArgumentString  = 0
	SimpleFunctionArgumentDecider = 1
	SimpleFunctionArgumentInt     = 2
	//SimpleFunctionArgumentEnum  = ---->
)

var FunctionMap = make(map[string]*Function)

func RegisterFunction(function *Function) {
	for _, nm := range function.OwnedKeywords {
		if _, ok := FunctionMap[nm]; !ok {
			FunctionMap[nm]=function
		}
	}
}

type EnumInfo struct {
	WantedValuesDescription string // "discrete, continuous, none"
	Parser func(string)byte
	InvalidValue byte
}

var SimpleFunctionEnums []*EnumInfo

func RegisterEnum(desc string,parser func(string)byte,inv byte) int {
	SimpleFunctionEnums=append(SimpleFunctionEnums,&EnumInfo{WantedValuesDescription:desc,InvalidValue:inv,Parser:parser})
	return len(SimpleFunctionEnums)-1+3
}
