package Happy2018new

import (
	_ "embed"
	"phoenixbuilder/omega/defines"
)

// 插件所属组织或者作者名
const NAMESPACE = "Happy2018new"

// 插件默认配置
//
//go:embed default_components.json
var DefaultComponentConfigByte []byte

// 插件列表
var Components = map[string]defines.Component{
	"记录方块改动日志": &RecordBlockChanges{BasicComponent: &defines.BasicComponent{}},
	"修改物品名称": &ChangeItemNameByUseAnvil{BasicComponent: &defines.BasicComponent{}},
	"提取 MCStructure": &ExportMCStructure{BasicComponent: &defines.BasicComponent{}},
}
