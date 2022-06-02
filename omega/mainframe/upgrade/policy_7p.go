package upgrade

import (
	"fmt"
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
	if version, err := checkMigrationVersion(root); err == nil && version < 547 {
		updateComponentConfig(root, "32k方块检测", func(c *defines.ComponentConfig) {
			fmt.Println("00000")
			if s := c.Configs["使用以下正则表达式检查"]; s != nil {
				fmt.Println("11111")
				if ts, success := s.([]interface{}); success {
					fmt.Println("22222")
					for i, te := range ts {
						if tte, success := te.(map[string]interface{}); success {
							tte["启用"] = false
							if raw, hasK := tte["附加指令"]; hasK {
								tte["附加指令"] = []interface{}{raw}
							}
							ts[i] = tte
						} else {
							ts[i] = te
						}
					}
					c.Configs["使用以下正则表达式检查"] = ts
				}
				fmt.Println("33333")
			}
		})
		setMigrationVersion(root, 547)
	}
}
