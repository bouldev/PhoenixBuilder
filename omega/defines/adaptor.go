package defines

import (
	"phoenixbuilder/fastbuilder/uqHolder"
	"phoenixbuilder/minecraft/protocol/packet"
)

// ConnectionAdaptor 描述了这个租赁服框架在被移植时还需要哪些接口
// GetBackendCommandFeeder 描述的是后台的命令行输入
// 如果可以保证从连接MC开始的每一个包都能通过 Read 获得，
// GetInitUQHolderCopy 可以返回一个空的 *uqHolder.UQHolder
// 如果不能保证，例如，连接到一个远程 FB，且 FB 先于该租赁服框架启动，
// 则必须从远程获得该信息，以保证状态信息的同步
// 如果连接断开，应该调用 System.Stop() 并销毁 System
type ConnectionAdaptor interface {
	Read() packet.Packet
	Write(packet.Packet)
	GetInitUQHolderCopy() *uqHolder.UQHolder
	GetBackendCommandFeeder() chan string
}

// System 描述了可以通过哪些接口控制这个租赁服框架
// 事实上，配置主要是通过 SetRoot 文件夹下的配置文件实现的，所以这些接口很简单
// 顺序 .SetRoot(root string) -> Bootstrap(ConnectionAdaptor) -> Activate() -> 连接断开时 -> Stop
// 请勿尝试 recover 其中的错误，如果发生崩溃，应该整个重启
type System interface {
	FullyStopped() chan struct{}
	Stop() error
	SetRoot(root string)
	Bootstrap(ConnectionAdaptor)
	Activate()
}

// ComponentConfig 描述了 插件 的配置内容，必须保证可被 yaml 正确处理
type ComponentConfig struct {
	Name        string                 `yaml:"name" json:"name"`
	Description string                 `yaml:"description" json:"description"`
	Disabled    bool                   `yaml:"disabled" json:"disabled"`
	Version     string                 `yaml:"version" json:"version"`
	Source      string                 `yaml:"source" json:"source"`
	Configs     map[string]interface{} `yaml:"configs" json:"configs"`
}

// Component 描述了插件应该具有的接口
// 顺序 &Component{} -> .Init(ComponentConfig) -> Activate() -> Stop()
// 每个 Activate 工作在一个独立的 goroutine 下
type Component interface {
	Init(cfg *ComponentConfig)
	Inject(frame MainFrame)
	Activate()
	Stop() error
}

type CoreComponent interface {
	Component
	SetSystem(interface{})
}
