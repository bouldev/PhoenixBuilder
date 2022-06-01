package upgrade

import "phoenixbuilder/omega/defines"

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
