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
	FBEval(cmd string)
}

// System 描述了可以通过哪些接口控制这个租赁服框架
// 事实上，配置主要是通过配置文件实现的，所以这些接口很简单
// 顺序 .Bootstrap(ConnectionAdaptor) -> Activate() -> 连接断开时 -> Stop
// 请勿尝试 recover 其中的错误，如果发生崩溃，应该整个重启
type System interface {
	FullyStopped() chan struct{}
	Stop() error
	Bootstrap(ConnectionAdaptor)
	Activate()
}

type TriggerConfig struct {
	DefaultTigger     string   `yaml:"默认触发词" json:"默认触发词"`
	TriggerWords      []string `yaml:"允许的触发词" json:"允许的触发词"`
	AllowNoSpace      bool     `yaml:"允许关键词之间没有空格" json:"允许关键词之间没有空格"`
	RemoveSuffixColor bool     `yaml:"去除花哨的用户名" json:"去除花哨的用户名"`
	AllowWisper       bool     `yaml:"允许悄悄话唤醒" json:"允许悄悄话唤醒"`
}

type OmegaConfig struct {
	Name                     string         `yaml:"系统名" json:"name,omitempty"`
	Version                  string         `yaml:"版本" json:"version,omitempty"`
	Trigger                  *TriggerConfig `yaml:"触发词" json:"触发词"`
	MigrationVersion         int            `yanl:"自动升级识别号" json:"自动升级识别号"`
	CommandFeedBackByDefault bool           `yaml:"默认情况下是否有commandfeedback" json:"默认情况下是否有commandfeedback,omitempty"`
	MemLimit                 int            `yaml:"限制内存占用" json:"限制内存占用"`
	ShowMemUsagePeriod       int            `yaml:"内存占用显示周期" json:"内存占用显示周期"`
}

// ComponentConfig 描述了 插件 的配置内容，必须保证可被 yaml 正确处理
type ComponentConfig struct {
	Name        string                 `json:"名称"`
	Description string                 `json:"描述"`
	Disabled    bool                   `json:"是否禁用"`
	Version     string                 `json:"版本"`
	Source      string                 `json:"来源"`
	Configs     map[string]interface{} `json:"配置"`
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
