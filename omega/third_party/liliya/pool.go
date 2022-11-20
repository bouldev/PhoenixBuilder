package liliya

import (
	_ "embed"
	"phoenixbuilder/omega/defines"
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
}
