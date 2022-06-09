package upgrade

import (
	"io/ioutil"
	"path"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
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
	offNbtBlockCheck := false
	if version, err := checkMigrationVersion(root); err == nil && version < 547 {
		offNbtBlockCheck = true
	}
	updateComponentConfig(root, "32k方块检测", func(c *defines.ComponentConfig) {
		if s := c.Configs["使用以下正则表达式检查"]; s != nil {
			if ts, success := s.([]interface{}); success {
				for i, te := range ts {
					if tte, success := te.(map[string]interface{}); success {
						if offNbtBlockCheck {
							tte["启用"] = false
						}
						if raw, hasK := tte["附加指令"]; hasK {
							switch tr := raw.(type) {
							case string:
								tte["附加指令"] = []interface{}{raw}
							case []interface{}:
								if cmd, success := tr[0].([]interface{}); success {
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

func Policy_9(root string) {
	if version, err := checkMigrationVersion(root); err == nil && version < 559 {
		updateComponentConfig(root, "群服互通", func(c *defines.ComponentConfig) {
			c.Configs["允许这些人透过QQ执行命令"] = map[string]bool{
				"1634268014": true,
				"1634268015": true,
				"1634268016": true,
			}
			c.Configs["屏蔽这些指令"] = map[string]string{
				"deop":         "达咩，不要啦~~~",
				"spreadplayer": "呜呜呜",
				"kill":         "nonono",
			}
		})
		setMigrationVersion(root, 559)
	} else if err != nil {
		// panic(err)
	}
}

func Policy_10(root string) {
	if version, err := checkMigrationVersion(root); err == nil && version < 575 {
		updateComponentConfig(root, "每日签到", func(c *defines.ComponentConfig) {
			c.Configs["玩家登录延迟"] = 20
			c.Configs["提醒玩家签到的消息"] = "@[player] 今日还未签到，记得签到喔"
			c.Configs["为true时被动签到false时主动签到"] = false
		})
		setMigrationVersion(root, 575)
	}
}

func Policy_11(root string) {
	if version, err := checkMigrationVersion(root); err == nil && version < 582 {
		d := path.Join(root, "配置")
		entries, err := ioutil.ReadDir(d)
		if err != nil {
			return
		}
		for _, entry := range entries {
			if entry.Name() == "主系统.json" {
				p := path.Join(root, "配置", entry.Name())
				c := &defines.OmegaConfig{}
				if err := utils.GetJsonData(p, c); err != nil {
					return
				}
				c.MigrationVersion = 582
				c.MemLimit = 400
				c.ShowMemUsagePeriod = 120
				utils.WriteJsonData(p, c)
			}
		}
		updateComponentConfig(root, "菜单显示", func(c *defines.ComponentConfig) {
			c.Configs["强制修改菜单信息"] = map[string]string{}
		})
	}
}
