package components

import (
	"phoenixbuilder/omega/components/qqGroupLink"
	"phoenixbuilder/omega/defines"
)

var HintOnRequireMappingUpdate = "该组件暂时无法在 1.18 客户端下正常工作，7.8~7.11号完成 ID 表更新后才可正常工作，请保持耐心,现在，请暂时禁用上述组件"

func GetComponentsPool() map[string]func() defines.Component {
	return map[string]func() defines.Component{
		"入服欢迎": func() defines.Component {
			return &Bonjour{BasicComponent: &defines.BasicComponent{}}
		},
		"聊天记录": func() defines.Component {
			return &ChatLogger{BasicComponent: &defines.BasicComponent{}}
		},
		"系统上线提示": func() defines.Component {
			return &Banner{BasicComponent: &defines.BasicComponent{}}
		},
		"反馈信息": func() defines.Component {
			return &FeedBack{BasicComponent: &defines.BasicComponent{}}
		},
		"玩家留言": func() defines.Component {
			return &Memo{BasicComponent: &defines.BasicComponent{}}
		},
		"玩家互传": func() defines.Component {
			return &PlayerTP{BasicComponent: &defines.BasicComponent{}}
		},
		"返回主城": func() defines.Component {
			return &BackToHQ{BasicComponent: &defines.BasicComponent{}}
		},
		"设置重生点": func() defines.Component {
			return &SetSpawnPoint{BasicComponent: &defines.BasicComponent{}}
		},
		"玩家自杀": func() defines.Component {
			return &Respawn{BasicComponent: &defines.BasicComponent{}}
		},
		"玩家信息": func() defines.Component {
			return &AboutMe{BasicComponent: &defines.BasicComponent{}}
		},
		"自定义传送点": func() defines.Component {
			return &Portal{BasicComponent: &defines.BasicComponent{}}
		},
		"返回死亡点": func() defines.Component {
			return &Immortal{BasicComponent: &defines.BasicComponent{}}
		},
		"踢人": func() defines.Component {
			return &Kick{BasicComponent: &defines.BasicComponent{}}
		},
		"商店": func() defines.Component {
			return &Shop{BasicComponent: &defines.BasicComponent{}}
		},
		"群服互通": func() defines.Component {
			return &qqGroupLink.QGroupLink{}
		},
		"物品回收": func() defines.Component {
			return &Recycle{BasicComponent: &defines.BasicComponent{}}
		},
		"OP权限模拟": func() defines.Component {
			return &FakeOp{BasicComponent: &defines.BasicComponent{}}
		},
		"简单自定义指令": func() defines.Component {
			return &SimpleCmd{BasicComponent: &defines.BasicComponent{}}
		},
		"计划任务": func() defines.Component {
			return &Schedule{BasicComponent: &defines.BasicComponent{}}
		},
		"时间同步": func() defines.Component {
			return &TimeSync{BasicComponent: &defines.BasicComponent{}}
		},
		"玩家转账": func() defines.Component {
			return &MoneyTransfer{BasicComponent: &defines.BasicComponent{}}
		},
		"自助建筑备份": func() defines.Component {
			return &StructureBackup{BasicComponent: &defines.BasicComponent{}}
		},
		"同步退出": func() defines.Component {
			return &Crash{BasicComponent: &defines.BasicComponent{}}
		},
		"手持32k检测": func() defines.Component {
			// TODO: Mapping Update
			return &defines.StubComponent{BasicComponent: &defines.BasicComponent{}, Hint: HintOnRequireMappingUpdate}
			// return &IntrusionDetectSystem{BasicComponent: &defines.BasicComponent{}}
		},
		"违规昵称检测": func() defines.Component {
			return &WhoAreYou{BasicComponent: &defines.BasicComponent{}}
		},
		"32k方块检测": func() defines.Component {
			// TODO: Mapping Update
			return &defines.StubComponent{BasicComponent: &defines.BasicComponent{}, Hint: HintOnRequireMappingUpdate}
			// return &ContainerScan{BasicComponent: &defines.BasicComponent{}}
		},
		"管理员检测": func() defines.Component {
			return &OpCheck{BasicComponent: &defines.BasicComponent{}}
		},
		"发言频率限制": func() defines.Component {
			return &ShutUp{BasicComponent: &defines.BasicComponent{}}
		},
		"计分板UID追踪": func() defines.Component {
			return &UIDTracking{BasicComponent: &defines.BasicComponent{}}
		},
		"区域扫描": func() defines.Component {
			return &Scanner{BasicComponent: &defines.BasicComponent{}}
		},
		"刷怪笼检测": func() defines.Component {
			return &MobSpawnerScan{BasicComponent: &defines.BasicComponent{}}
		},
		"快递系统": func() defines.Component {
			return &Express{BasicComponent: &defines.BasicComponent{}}
		},
		"高频红石检查": func() defines.Component {
			// TODO: Mapping Update
			return &defines.StubComponent{BasicComponent: &defines.BasicComponent{}, Hint: HintOnRequireMappingUpdate}
			// return &RedStoneUpdateLimit{BasicComponent: &defines.BasicComponent{}}
		},
		"兑换码": func() defines.Component {
			return &CDkey{BasicComponent: &defines.BasicComponent{}}
		},
		"切换": func() defines.Component {
			return &StatusToggle{BasicComponent: &defines.BasicComponent{}}
		},
		"排行榜": func() defines.Component {
			return &Ranking{BasicComponent: &defines.BasicComponent{}}
		},
		"每日签到": func() defines.Component {
			return &DailyAttendance{BasicComponent: &defines.BasicComponent{}}
		},
		"小木斧": func() defines.Component {
			// TODO: Mapping Update
			return &defines.StubComponent{BasicComponent: &defines.BasicComponent{}, Hint: HintOnRequireMappingUpdate}
			// return &woodaxe.WoodAxe{BasicComponent: &defines.BasicComponent{}}
		},
		"存档修复": func() defines.Component {
			// TODO: Mapping Update
			return &defines.StubComponent{BasicComponent: &defines.BasicComponent{}, Hint: HintOnRequireMappingUpdate}
			// return &DifferRecover{BasicComponent: &defines.BasicComponent{}}
		},
		"玩家商店": func() defines.Component {
			return &PlayerShop{BasicComponent: &defines.BasicComponent{}}
		},
		"封禁时间": func() defines.Component {
			return &BanTime{BasicComponent: &defines.BasicComponent{}}
		},
		"消除方块": func() defines.Component {
			// TODO: Mapping Update
			return &defines.StubComponent{BasicComponent: &defines.BasicComponent{}, Hint: HintOnRequireMappingUpdate}
			// return &RemoveBlock{BasicComponent: &defines.BasicComponent{}}
		},
	}
}
