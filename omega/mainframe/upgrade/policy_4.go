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

//go:embed new_regex_for_ids.json
var newRegexForIDSData []byte

func Policy_4(root string) {
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
		if c.Name != "手持32k检测" {
			continue
		}
		if _, hask := c.Configs["使用以下正则表达式检查"]; !hask {
			var regexs interface{}
			json.Unmarshal(newRegexForIDSData, &regexs)
			c.Configs["使用以下正则表达式检查"] = regexs
			utils.WriteJsonData(p, c)
		}
	}
}
