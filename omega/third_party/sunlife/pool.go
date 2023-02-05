package sunlife

import (
	_ "embed"
	"phoenixbuilder/omega/defines"
)

// 插件所属组织或者作者名
const NAMESPACE = "温柔"

// 插件默认配置
//
//go:embed default_components.json
var DefaultComponentConfigByte []byte

// 插件列表
var Components = map[string]defines.Component{
	"抽奖":         &NewLottery{BasicComponent: &defines.BasicComponent{}},
	"银行系统":       &Bank{BasicComponent: &defines.BasicComponent{}},
	"虚拟称号系统":     &Designation{BasicComponent: &defines.BasicComponent{}},
	"虚拟装备系统":     &Equip{BasicComponent: &defines.BasicComponent{}},
	"Pvp系统":      &Pvp{BasicComponent: &defines.BasicComponent{}},
	"自定义附魔升级系统":  &Enchant{BasicComponent: &defines.BasicComponent{}},
	"无前缀简单自定义指令": &CustomCmd{BasicComponent: &defines.BasicComponent{}},
	"王国系统":       &King{BasicComponent: &defines.BasicComponent{}},
}
