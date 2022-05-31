package upgrade

import (
	_ "embed"
	"encoding/json"
	"io/ioutil"
	"path"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"strings"
)

//go:embed new_menu_structure_for_menu.json
var newMenuStructureForMenu []byte

func Policy_6(root string) {
	d := path.Join(root, "配置")
	entries, err := ioutil.ReadDir(d)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), "组件") {
			continue
		}
		p := path.Join(root, "配置", entry.Name())
		c := &defines.ComponentConfig{}
		if err := utils.GetJsonData(p, c); err != nil {
			panic("处理[" + p + "]时出错" + err.Error())
		}
		if c.Name != "菜单显示" {
			continue
		}
		if _, hask := c.Configs["目录结构"]; !hask {
			var structure interface{}
			json.Unmarshal(newMenuStructureForMenu, &structure)
			c.Configs["目录结构"] = structure
			utils.WriteJsonData(p, c)
		}
	}
}
