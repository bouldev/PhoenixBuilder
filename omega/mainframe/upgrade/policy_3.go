package upgrade

import (
	"io/ioutil"
	"path"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"strings"
)

func Policy_3(root string) {
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
		if c.Name != "玩家互传" {
			continue
		}
		if _, hask := c.Configs["传送指令"]; !hask {
			c.Configs["传送指令"] = "tp @a[name=[src],tag=!banTP] @a[name=[dst],tag=!banTP]"
			utils.WriteJsonData(p, c)
		}
	}
}
