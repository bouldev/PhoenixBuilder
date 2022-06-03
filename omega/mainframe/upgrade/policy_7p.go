package upgrade

import (
	"phoenixbuilder/omega/defines"
)

func Policy_7(root string) {
	if version, err := checkMigrationVersion(root); err == nil && version < 542 {
		updateComponentConfig(root, "菜单显示", func(c *defines.ComponentConfig) {
			if s := c.Configs["目录结构"]; s != nil {
				if ts, success := s.([]interface{}); success {
					newTS := make([]interface{}, 0)
					for _, e := range ts {
						if te, success := e.(map[string]interface{}); success {
							if v, hasK := te["触发词"]; hasK {
								if tv, success := v.(string); success {
									if tv == "传送" {
										continue
									}
								}
							}
						}
						newTS = append(newTS, e)
					}
					c.Configs["目录结构"] = newTS
				}
			}
		})
		setMigrationVersion(root, 542)
	} else if err != nil {
		// panic(err)
	}
}

func Policy_8(root string) {
	updateComponentConfig(root, "32k方块检测", func(c *defines.ComponentConfig) {
		if s := c.Configs["使用以下正则表达式检查"]; s != nil {
			if ts, success := s.([]interface{}); success {
				for i, te := range ts {
					if tte, success := te.(map[string]interface{}); success {
						tte["启用"] = false
						if raw, hasK := tte["附加指令"]; hasK {
							switch tr:=raw.(type){
								case string:
									tte["附加指令"] = []interface{}{raw}
								case []interface{}:
									if cmd,success:=tr[0].([]interface{});success{
										tte["附加指令"] = cmd
									}
							}
						}
						ts[i] = tte
					} else {
						ts[i] = te
					}
				}
				c.Configs["使用以下正则表达式检查"] = ts
			}
		}
	})

	if version, err := checkMigrationVersion(root); err == nil && version < 547 {
		setMigrationVersion(root, 547)
	}
}
