package plugin_structs

type PluginBridge interface {
	ConvertFunctionChainItemList(list map[string]FunctionChainItem) interface{}
	RegisterBuilder(name string, function func(config MainConfig, blc chan *Module) error) bool
	// --> function.RegisterFunction
	RegisterFunction(function Function)
	// --> function.RegisterEnum
	RegisterEnum(desc string, parser func(string)byte, inv byte) int
	// --> command.Tellraw -> command.SendChat
	Tellraw(message string) error
	// --> command.SendChat
	SendChat(content string) error
	// --> command.SendSizukanaCommand
	SendCommand(command string) error
	// "CB" stands for callback.
	SendCommandCB(command string, cb func([]CommandOutputMessage,string))
	SendWSCommandCB(command string, cb func([]CommandOutputMessage,string))
}

type CommandOutputMessage struct {
	Success bool
	Message string
	Parameters []string
}

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

