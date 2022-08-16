package example

import (
	_ "embed"
	"phoenixbuilder/omega/defines"
)

// 插件所属组织或者作者名
const NAMESPACE = "2401PT"

// 插件默认配置
//go:embed default_components.json
var DefaultComponentConfigByte []byte

// 插件列表
var Components = map[string]defines.Component{
	"喵喵喵": &EchoMiao{BasicComponent: &defines.BasicComponent{}},
}
