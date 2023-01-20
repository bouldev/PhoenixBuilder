package liliya

import (
	_ "embed"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/third_party/liliya/QGuildLink"
)

// 插件所属组织或者作者名
const NAMESPACE = "Liliya233"

// 插件默认配置
//
//go:embed default_components.json
var DefaultComponentConfigByte []byte

// 插件列表
var Components = map[string]defines.Component{
	"选取方块": &PickBlock{BasicComponent: &defines.BasicComponent{}},
	"能力管理": &AbilityManage{BasicComponent: &defines.BasicComponent{}},
	"挂机检测": &AntiAFK{BasicComponent: &defines.BasicComponent{}},
	"入睡检测": &PlayersSleepingPercentage{BasicComponent: &defines.BasicComponent{}},
	"频服互通": &QGuildLink.QGuildLink{BasicComponent: &defines.BasicComponent{}},
}
