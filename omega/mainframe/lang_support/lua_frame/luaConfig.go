package luaFrame

// 描述了一个lua插件配置该有的东西
type LuaCommpoentConfig struct {
	Name     string                 `json:"名称"`
	Usage    string                 `json:"描述"`
	Disabled bool                   `json:"是否禁用"`
	Version  string                 `json:"版本"`
	Author   string                 `json:"作者"`
	Config   map[string]interface{} `json:"配置"`
	Source   string                 `json:"来源"`
}
