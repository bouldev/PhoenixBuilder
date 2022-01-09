package plugin

import (
	"phoenixbuilder/minecraft"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/fastbuilder/types"
	"phoenixbuilder/fastbuilder/command"
	"phoenixbuilder/fastbuilder/builder"
	"phoenixbuilder/fastbuilder/function"
	"phoenixbuilder/fastbuilder/plugin_structs"
	"github.com/google/uuid"
)

type PluginBridgeImpl struct {
	sessionConnection *minecraft.Conn
}

func (*PluginBridgeImpl) ConvertFunctionChainItemList(list map[string]plugin_structs.FunctionChainItem) interface{} {
	outmap:=make(map[string]*function.FunctionChainItem)
	for key, val := range list {
		mw:=function.FunctionChainItem(val)
		outmap[key]=&mw
	}
	return outmap
}

func (*PluginBridgeImpl) RegisterBuilder(name string, function_cont func(config plugin_structs.MainConfig, blc chan *plugin_structs.Module) error) bool {
	_, ahas:=builder.Builder[name]
	if ahas {
		return false
	}
	builder.Builder[name]=func(config *types.MainConfig, blc chan *types.Module) error {
		conchan:=make(chan *plugin_structs.Module)
		go func() {
			for {
				curblock, ok:=<-conchan
				if !ok {
					break
				}
				convcbdata:=types.CommandBlockData(*curblock.CommandBlockData)
				blc<-&types.Module {
					Block: &types.Block {
						Name:curblock.Block.Name,
						Data:curblock.Block.Data,
					},
					CommandBlockData: &convcbdata,
					Entity: (*types.Entity)(curblock.Entity),
					Point: types.Position {
						curblock.Point.X,
						curblock.Point.Y,
						curblock.Point.Z,
					},
				}
			}
		} ()
		err:=function_cont(plugin_structs.MainConfig{
			Execute: config.Execute,
			Block: &plugin_structs.ConstBlock {
				Name: config.Block.Name,
				Data: config.Block.Data,
			},
			OldBlock: &plugin_structs.ConstBlock {
				Name: config.OldBlock.Name,
				Data: config.OldBlock.Data,
			},
			End: plugin_structs.Position {
				config.End.X,
				config.End.Y,
				config.End.Z,
			},
			Position: plugin_structs.Position {
				config.Position.X,
				config.Position.Y,
				config.Position.Z,
			},
			Radius: config.Radius,
			Length: config.Length,
			Width: config.Width,
			Height: config.Height,
			Method: config.Method,
			OldMethod: config.OldMethod,
			Facing: config.Facing,
			Path: config.Path,
			Shape: config.Shape,
			ExcludeCommands: config.ExcludeCommands,
			InvalidateCommands: config.InvalidateCommands,
			Strict: config.Strict,
		},conchan)
		close(conchan)
		return err
	}
	return true
}

func (*PluginBridgeImpl) RegisterFunction(function_cont plugin_structs.Function) {
	funcco:=function.Function(function_cont)
	function.RegisterFunction(&funcco)
}

func (*PluginBridgeImpl) RegisterEnum(desc string, parser func(string)byte, inv byte) int {
	return function.RegisterEnum(desc,parser,inv)
}

func (br *PluginBridgeImpl) Tellraw(message string) error {
	return command.Tellraw(br.sessionConnection, message)
}

func (br *PluginBridgeImpl) SendChat(content string) error {
	return command.SendChat(content,br.sessionConnection)
}

func (br *PluginBridgeImpl) SendCommand(commandstr string) error {
	return command.SendSizukanaCommand(commandstr,br.sessionConnection)
}

func (br *PluginBridgeImpl) SendCommandCB(cmd string, cb func([]plugin_structs.CommandOutputMessage,string)) {
	wchan:=make(chan *packet.CommandOutput)
	ud, _ := uuid.NewUUID()
	command.UUIDMap.Store(ud.String(), wchan)
	command.SendCommand(cmd, ud, br.sessionConnection)
	resp:=<-wchan
	close(wchan)
	unk:=resp.DataSet
	arr:=make([]plugin_structs.CommandOutputMessage,len(resp.OutputMessages))
	for i,c:= range resp.OutputMessages {
		arr[i]=plugin_structs.CommandOutputMessage(c)
	}
	cb(arr,unk)
}

func (br *PluginBridgeImpl) SendWSCommandCB(cmd string, cb func([]plugin_structs.CommandOutputMessage,string)) {
	wchan:=make(chan *packet.CommandOutput)
	ud, _ := uuid.NewUUID()
	command.UUIDMap.Store(ud.String(), wchan)
	command.SendWSCommand(cmd, ud, br.sessionConnection)
	resp:=<-wchan
	close(wchan)
	unk:=resp.DataSet
	arr:=make([]plugin_structs.CommandOutputMessage,len(resp.OutputMessages))
	for i,c:= range resp.OutputMessages {
		arr[i]=plugin_structs.CommandOutputMessage(c)
	}
	cb(arr,unk)
}