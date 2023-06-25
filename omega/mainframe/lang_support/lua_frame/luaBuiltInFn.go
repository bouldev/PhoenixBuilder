package luaFrame

import (
	"fmt"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	omgApi "phoenixbuilder/omega/mainframe/lang_support/lua_frame/omgcomponentapi"
	"sync"

	lua "github.com/yuin/gopher-lua"
)

// 内置函数加载器
type BuiltlnFn struct {
	//omg组件
	OmegaFrame *omgApi.OmgApi
	Listener   sync.Map
	mainframe  defines.MainFrame
}

// 写入
func (b *BuiltlnFn) LoadFn(L *lua.LState) error {

	// 创建一个Lua table
	// 注册Listener类型
	mt := L.NewTypeMetatable("listener")
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"NextMsg": NextMsg,
	}))
	//注册skynet
	skynet := L.NewTable()

	//注入方法 GetListener GetControl
	L.SetField(skynet, "GetListener", L.NewFunction(b.GetListener))
	L.SetField(skynet, "GetControl", L.NewFunction(b.GetControl))
	// 将table命名为ComplexStruct，并将其设为全局变量
	L.SetGlobal("skynet", skynet)
	return nil
}

func (b *BuiltlnFn) GetListener(L *lua.LState) int {
	listener := L.NewTable()
	//listener的方法 listen("可变参数") 获取参数  listenPackage(Id)

	L.SetField(listener, "GetMsgListner", L.NewFunction(b.GetMsgListener))
	L.SetField(listener, "listenPackage", L.NewFunction(func(l *lua.LState) int {

		return 1
	}))
	//返回listener对象
	L.Push(listener)
	return 1
}
func (b *BuiltlnFn) GetControl(L *lua.LState) int {
	GameControl := L.NewTable()
	L.SetField(GameControl, "SendWsCmd", L.NewFunction(func(l *lua.LState) int {

		args := L.CheckString(1)
		b.OmegaFrame.MainFrame.GetGameControl().SendCmd(args)

		return 1
	}))
	L.SetField(GameControl, "SendCmdAndInvokeOnResponse", L.NewFunction(func(l *lua.LState) int {
		if l.GetTop() == 1 {
			args := L.CheckString(1)
			ch := make(chan bool)
			b.OmegaFrame.MainFrame.GetGameControl().SendCmdAndInvokeOnResponse(args, func(output *packet.CommandOutput) {
				cmdBack := L.NewTable()
				if output.SuccessCount > 0 {
					L.SetField(cmdBack, "Success", lua.LBool(true))
				} else {
					L.SetField(cmdBack, "Success", lua.LBool(false))
				}
				L.SetField(cmdBack, "outputmsg", lua.LString(fmt.Sprintf("%v", output.OutputMessages)))
				L.Push(cmdBack)
				ch <- true
			})
			<-ch
		}
		return 1
	}))
	L.Push(GameControl)
	return 1
}

// 模拟消息
type Message struct {
	Type    string
	Content string
}

// 监听器
// Listener 结构体
type Listener struct {
	MsgChannel chan Message // 每个监听器都有一个独立的消息通道
}

// GetListener 创建一个新的监听器并返回其引用
func (f *BuiltlnFn) GetMsgListener(L *lua.LState) int {
	listener := &Listener{MsgChannel: make(chan Message, 25)} // 创建一个新监听器实例，并初始化其消息通道容量为25
	ptr := &f.Listener
	ptr.Store(listener, struct{}{}) // 将新监听器添加到监听器集合中

	ud := L.NewUserData()                              // 创建一个新的UserData，用于在Lua中表示监听器实例
	ud.Value = listener                                // 将监听器实例存储在UserData中
	L.SetMetatable(ud, L.GetTypeMetatable("listener")) // 设置UserData的元表

	L.Push(ud) // 将UserData返回给Lua
	return 1
}

// NextMsg 用于从监听器的消息通道中获取下一个消息
func NextMsg(L *lua.LState) int {
	ud := L.CheckUserData(1)         // 从Lua参数中获取UserData
	listener := ud.Value.(*Listener) // 从UserData中提取监听器实例

	msg := <-listener.MsgChannel // 从监听器的消息通道中读取下一个消息，如果没有消息，则阻塞等待
	L.Push(lua.LString(msg.Type))
	L.Push(lua.LString(msg.Content))
	return 2
}
