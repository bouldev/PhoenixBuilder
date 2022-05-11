package components

import (
	"phoenixbuilder/omega/components/qqGroupLink"
	"phoenixbuilder/omega/defines"
)

type BasicComponent struct {
	Config   *defines.ComponentConfig
	Frame    defines.MainFrame
	Ctrl     defines.GameControl
	Listener defines.GameListener
}

func (bc *BasicComponent) Init(cfg *defines.ComponentConfig) {
	bc.Config = cfg
}

func (bc *BasicComponent) Inject(frame defines.MainFrame) {
	bc.Frame = frame
	bc.Listener = frame.GetGameListener()
}

func (bc *BasicComponent) Activate() {
	bc.Ctrl = bc.Frame.GetGameControl()
}

func (bc *BasicComponent) Stop() error {
	return nil
}

func GetComponentsPool() map[string]func() defines.Component {
	return map[string]func() defines.Component{
		"入服欢迎": func() defines.Component {
			return &Bonjour{BasicComponent: &BasicComponent{}}
		},
		"聊天记录": func() defines.Component {
			return &ChatLogger{BasicComponent: &BasicComponent{}}
		},
		"系统上线提示": func() defines.Component {
			return &Banner{BasicComponent: &BasicComponent{}}
		},
		"反馈信息": func() defines.Component {
			return &FeedBack{BasicComponent: &BasicComponent{}}
		},
		"玩家留言": func() defines.Component {
			return &Memo{BasicComponent: &BasicComponent{}}
		},
		"玩家互传": func() defines.Component {
			return &PlayerTP{BasicComponent: &BasicComponent{}}
		},
		"返回主城": func() defines.Component {
			return &BackToHQ{BasicComponent: &BasicComponent{}}
		},
		"设置重生点": func() defines.Component {
			return &SetSpawnPoint{BasicComponent: &BasicComponent{}}
		},
		"玩家自杀": func() defines.Component {
			return &Respawn{BasicComponent: &BasicComponent{}}
		},
		"玩家信息": func() defines.Component {
			return &AboutMe{BasicComponent: &BasicComponent{}}
		},
		"自定义传送点": func() defines.Component {
			return &Portal{BasicComponent: &BasicComponent{}}
		},
		"返回死亡点": func() defines.Component {
			return &Immortal{BasicComponent: &BasicComponent{}}
		},
		"踢人": func() defines.Component {
			return &Kick{BasicComponent: &BasicComponent{}}
		},
		"商店": func() defines.Component {
			return &Shop{BasicComponent: &BasicComponent{}}
		},
		"群服互通": func() defines.Component {
			return &qqGroupLink.QGroupLink{}
		},
		"物品回收": func() defines.Component {
			return &Recycle{BasicComponent: &BasicComponent{}}
		},
		"OP权限模拟": func() defines.Component {
			return &FakeOp{BasicComponent: &BasicComponent{}}
		},
		"简单自定义指令": func() defines.Component {
			return &SimpleCmd{BasicComponent: &BasicComponent{}}
		},
		"计划任务": func() defines.Component {
			return &Schedule{BasicComponent: &BasicComponent{}}
		},
		"时间同步": func() defines.Component {
			return &TimeSync{BasicComponent: &BasicComponent{}}
		},
		"玩家转账": func() defines.Component {
			return &MoneyTransfer{BasicComponent: &BasicComponent{}}
		},
		"自助建筑备份": func() defines.Component {
			return &StructureBackup{BasicComponent: &BasicComponent{}}
		},
		"同步退出": func() defines.Component {
			return &Crash{BasicComponent: &BasicComponent{}}
		},
		"手持32k检测": func() defines.Component {
			return &IntrusionDetectSystem{BasicComponent: &BasicComponent{}}
		},
		"违规昵称检测": func() defines.Component {
			return &WhoAreYou{BasicComponent: &BasicComponent{}}
		},
		"32k方块检测": func() defines.Component {
			return &ContainerScan{BasicComponent: &BasicComponent{}}
		},
		"管理员检测": func() defines.Component {
			return &OpCheck{BasicComponent: &BasicComponent{}}
		},
		"发言频率限制": func() defines.Component {
			return &ShutUp{BasicComponent: &BasicComponent{}}
		},
		"计分板UID追踪": func() defines.Component {
			return &UIDTracking{BasicComponent: &BasicComponent{}}
		},
	}
}
