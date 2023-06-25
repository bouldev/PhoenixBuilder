package luaFrame

import (
	"errors"
	"fmt"
	"path/filepath"
	omgApi "phoenixbuilder/omega/mainframe/lang_support/lua_frame/omgcomponentapi"
	"sync"

	lua "github.com/yuin/gopher-lua"
)

const (
	COMPONENT_INIT_FN   = "init"
	COMPONENT_ACTIVE_FN = "active"
	OMGPATH             = "omega_storage" + SEPA + "data"
	OMGCONFIGPATH       = OMGROOTPATH + SEPA + "配置"
	OMGROOTPATH         = "omega_storage"
	OMGDATAPATH         = OMGROOTPATH + SEPA + "data"
	SEPA                = string(filepath.Separator)
	LUASOURCE           = "Lua-Component"
)

// 插件监测器
type Monitor struct {
	//每个插件拥有自己的lua运行环境 并且每个插件的名字都将是这个插件唯一的指示标志
	//在运行的初期就会初始化所有的插件 并且根据产生的配置文件决定是否开启 这与omg普通插件没有区别
	//区别点在于lua的优势导致 这个插件能够热重载以及能够修改其中的主要逻辑
	ComponentPoll map[string]*LuaComponent
	//omg框架
	LuaComponentData map[string]Result
	OmgFrame         *omgApi.OmgApi
	FileControl      *FileControl
	BuiltlnFner      *BuiltlnFn
}

// 插件
type LuaComponent struct {
	L *lua.LState
	//排队中的消息
	Msg map[string]string
	//是否运行
	Running bool
	//插件的配置
	Config LuaCommpoentConfig
}

func NewMonitor(lc *omgApi.OmgApi) *Monitor {
	return &Monitor{
		ComponentPoll: make(map[string]*LuaComponent),
		//获取omg框架
		OmgFrame: lc,
		BuiltlnFner: &BuiltlnFn{
			OmegaFrame: lc,
			Listener:   sync.Map{},
		},
		FileControl: &FileControl{},
	}

}

// 更新lua插件的地址
func (m *Monitor) InintComponents() {
	//m.FileControl.CheckFilePath()
	//获取路径
	luaComponentData, err_first := m.FileControl.GetLuaComponentData()

	m.LuaComponentData = luaComponentData
	if err_first != nil {
		PrintInfo(NewPrintMsg("警告", err_first))
	}

}

// 单独加载某个插件
func (m *Monitor) InjectComponent(name string) error {
	if _v, ok := m.LuaComponentData[name]; !ok {
		errors.New(fmt.Sprintf("你正在尝试运行组件:%v 但是它并未在配置文件之中找到 请确定它存在", name))
	} else {
		k := name
		v := _v
		_config := v.JsonConfig
		//如果配置文件是开启
		if !_config.Disabled {
			L := lua.NewState()
			// 为 Lua 虚拟机提供一个安全的环境 提供基础的方法
			if err := m.BuiltlnFner.LoadFn(L); err != nil {
				return err
			}
			m.ComponentPoll[k] = &LuaComponent{
				L:       L,
				Msg:     make(map[string]string),
				Running: false, //初始化完成但是未运行
				Config:  _config,
			}
		} else {
			errors.New(fmt.Sprintf("找到了%v插件配置文件 但是配置文件处于关闭状态", name))
		}
	}

	return nil
}

// 加载配置文件 创始pool池子
func (m *Monitor) InjectComponents() error {
	//开启配置文件为开启的 将决定开启的加入componentPool
	for _k, _v := range m.LuaComponentData {
		k := _k
		v := _v
		_config := v.JsonConfig
		//如果配置文件是开启
		if !_config.Disabled {
			L := lua.NewState()
			// 为 Lua 虚拟机提供一个安全的环境 提供基础的方法
			if err := m.BuiltlnFner.LoadFn(L); err != nil {
				return err
			}
			m.ComponentPoll[k] = &LuaComponent{
				L:       L,
				Msg:     make(map[string]string),
				Running: false, //初始化完成但是未运行
				Config:  _config,
			}
		}
	}
	return nil
}

// 接受指令处理并且执行
func (m *Monitor) CmdCenter(msg string) error {

	CmdMsg := FormateCmd(msg)
	if !CmdMsg.isCmd {
		return errors.New(fmt.Sprintf("很显然%v并不是指令的任何一种 请输入lua luas help寻求帮助", msg))
	}

	switch CmdMsg.Head {
	case HEADLUA:
		//lua指令
		if err := m.luaCmdHandler(&CmdMsg); err != nil {
			PrintInfo(NewPrintMsg("警告", err))
		}
	case HEADRELOAD:
		go func() {
			if err := m.Reload(&CmdMsg); err != nil {
				PrintInfo(NewPrintMsg("警告", err))
			}
		}()
		/*
			case HEADSTART:
				go func() {
					if err := m.StartCmdHandler(&CmdMsg); err != nil {
						PrintInfo(NewPrintMsg("警告", err))
					}
				}()
		*/
	}
	return nil
}

// 插件行为 重加载某个插件 如果参数为all则全部插件重加载 记住reload和startComponent是有区别的
// reload是再次扫描对应的插件然后默认不开启 而startCompent是直接在插件池子里面开启插件
func (m *Monitor) Reload(cmdmsg *CmdMsg) error {

	switch cmdmsg.Behavior {
	case "component":
		args := cmdmsg.args
		if len(args) != 1 {
			return errors.New("lua reload compoent指令后面应该有且仅有一个参数")
		}
		componentName := args[0]
		//更新一次文件
		m.InintComponents()
		//检查

		if args[0] == "all" {
			//依次关闭插件
			for k, _ := range m.ComponentPoll {
				m.CloseLua(k)
			}
			m.InjectComponents()
			//开启组件
			for k, _ := range m.ComponentPoll {
				err := m.StartComponent(k, m.LuaComponentData[k].LuaFile)
				if err != nil {
					PrintInfo(NewPrintMsg("警告", err))
				}
			}
			return nil
		}
		//初始化
		if err := m.CloseLua(componentName); err != nil {
			return err
		}
		if err := m.InjectComponent(componentName); err != nil {
			return err
		}
		//运行组件
		if v, ok := m.LuaComponentData[componentName]; !ok {
			return errors.New(fmt.Sprintf("你正在尝试运行组件:%v 但是它并未在配置文件之中找到 请确定它存在", componentName))
		} else {
			if err := m.StartComponent(componentName, v.LuaFile); err != nil {
				return err
			}
		}
		/*
			if err := m.Load(componentName); err != nil {
				return err
			}*/
		PrintInfo(NewPrintMsg("提示", fmt.Sprintf("%v已经重新加载", componentName)))
		return nil
	default:
		PrintInfo(NewPrintMsg("警告", "无效指令"))

	}
	return nil
}

/*
// 处理cmd
func (m *Monitor) StartCmdHandler(CmdMsg *CmdMsg) error {
	args := CmdMsg.args
	switch CmdMsg.Behavior {
	case "component":
		if len(args) != 1 {
			return errors.New("lua start compoent指令后面应该有且仅有一个参数")
		}
		if args[0] == "all" {
			//to do
			PrintInfo(NewPrintMsg("提示", fmt.Sprintf("全部插件已经开启")))
			return nil
		}
		// to do (修改)
		//componentName := args[0]
		//if err := m.Run(componentName); err != nil {
		//PrintInfo(NewPrintMsg("警告", err))
		//} else {
		//PrintInfo(NewPrintMsg("提示", fmt.Sprintf("%v插件已经开启", componentName)))
		//}

	default:
		PrintInfo(NewPrintMsg("警告", "这不是一个合理的指令"))
	}
	return nil
}*/

// 启动插件 name为插件名字luapath为lua代码的路径 启动插件时保证每次代码都是新的 所以会删除原有的插件
func (m *Monitor) StartComponent(name string, luapath string) error {
	//将原有插件删除
	if err := m.CloseLua(name); err != nil {
		return err
	}
	//判断是否存在插件
	if _, ok := m.ComponentPoll[name]; !ok {
		//查找配置 看是否存在
		return errors.New(fmt.Sprintf("你正在尝试运行组件:%v 但是它并未在配置文件之中找到 请确定它存在", name))
	}
	//执行代码
	go func(Name string, luaPath string) {
		m.ComponentPoll[Name].Running = true
		PrintInfo(NewPrintMsg("lua插件", fmt.Sprintf("%v插件启动成功 版本:%v", Name, m.ComponentPoll[Name].Config.Version)))

		if err := m.ComponentPoll[Name].L.DoFile(luaPath); err != nil {
			PrintInfo(NewPrintMsg("lua代码报错", err))
		}

	}(name, luapath)
	return nil
}

// 安全地关闭组件并且从配置文件中删除
func (m *Monitor) CloseLua(name string) error {
	if v, ok := m.ComponentPoll[name]; ok && v.Running {
		v.L.Close()
		v.Running = false
		delete(m.ComponentPoll, name)
		return nil
	}
	return nil
}

// lua指令类执行
func (m *Monitor) luaCmdHandler(CmdMsg *CmdMsg) error {
	args := CmdMsg.args
	switch CmdMsg.Behavior {
	case "help":
		warning := []string{
			"lua luas help 寻求指令帮助\n",
			"lua reload component [重加载的插件名字] 加载/重加载指定插件 如果参数是all就是全部插件重载\n",
			"lua luas new [新插件名字] [描述]创建一个自定义空白插件[描述为选填]\n",
			"lua luas delect [插件名字]\n",
			"lua luas list 列出当前正在运行的插件\n",
			"lua luas stop [插件名字] 暂停插件运行 参数为all则暂停所有插件运行",
		}
		msg := ""
		for _, v := range warning {
			msg += v
		}
		PrintInfo(NewPrintMsg("提示", msg))
	case "new":
		//参数检查
		if len(args) != 1 && len(args) != 2 {
			return errors.New("lua luas new后面应该加上[插件名字]或者说[插件名字]")
		}
		componentName := args[0]
		/*
			componentUsage := ""
			if len(args) == 2 {
				componentUsage = args[1]
			}*/
		//检查当前是否有
		if _, ok := m.LuaComponentData[componentName]; ok {
			return errors.New(fmt.Sprintf("已经含有%v插件 无法创立", componentName))
		}
		//如果没有则创建文件
		if err := m.FileControl.CreateDirAndFiles(componentName); err != nil {
			return err
		}
		PrintInfo(NewPrintMsg("提示", fmt.Sprintf("已经创建文件基本结构请到目录%v 修改", OMGCONFIGPATH+SEPA+componentName)))

	case "delect":
		if len(args) != 1 {
			return errors.New("lua luas delect指令后面应该加上需要删除的插件名字")
		}
		//从运行中删除
		m.CloseLua(args[0])
		//文件删除
		m.FileControl.DelectCompoentFile(args[0]) //DelectCompoent(args[0])
	case "list":
		msg := ""
		for k, v := range m.ComponentPoll {
			if v.Running {
				msg += fmt.Sprintf("[%v]", k)
			}
		}
		PrintInfo(NewPrintMsg("信息", msg+"处于开启状态"))
	case "stop":
		if len(args) != 1 {
			return errors.New("lua luas stop指令后面应该加上需要删除的插件名字")
		}
		name := args[0]
		if name == "all" {
			for k, _ := range m.ComponentPoll {
				m.CloseLua(k)
				PrintInfo(NewPrintMsg("提示", fmt.Sprintf("%v插件关闭成功", k)))
			}
			PrintInfo(NewPrintMsg("提示", "全部组件已经关闭"))
			return nil
		}
		if _, ok := m.ComponentPoll[name]; !ok {
			return errors.New(fmt.Sprintf("我们并没有在加载的插件池子中找到%v", name))
		}
		m.CloseLua(name)
		PrintInfo(NewPrintMsg("提示", fmt.Sprintf("%v插件关闭成功", name)))

	default:
		return errors.New("未知指令 请输入lua luas help寻求帮助")
	}
	return nil
}
