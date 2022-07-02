package defines

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
	Signal(int) error
}

type CoreComponent interface {
	Component
	SetSystem(interface{})
}

const (
	// 设计失误之一，由于希望使用者可以直接阅读数据，就没有上数据库，后果就是进程被强杀时会掉数据
	// 所以需要 这个 SIGNAL，让组件时不时的保存一下数据
	SIGNAL_DATA_CHECKPOINT = iota
)

type BasicComponent struct {
	Config   *ComponentConfig
	Frame    MainFrame
	Ctrl     GameControl
	Listener GameListener
}

func (bc *BasicComponent) Init(cfg *ComponentConfig) {
	bc.Config = cfg
}

func (bc *BasicComponent) Inject(frame MainFrame) {
	bc.Frame = frame
	bc.Listener = frame.GetGameListener()
}

func (bc *BasicComponent) Activate() {
	bc.Ctrl = bc.Frame.GetGameControl()
}

func (bc *BasicComponent) Stop() error {
	return nil
}

func (bc *BasicComponent) Signal(signal int) error {
	return nil
}
