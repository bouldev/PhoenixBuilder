package upgrade

import (
	"encoding/json"
	"os"
	"path"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
)

type ComponentConfig struct {
	Name        string                 `yaml:"name" json:"name"`
	Description string                 `yaml:"description" json:"description"`
	Disabled    bool                   `yaml:"disabled" json:"disabled"`
	Version     string                 `yaml:"version" json:"version"`
	Source      string                 `yaml:"source" json:"source"`
	Configs     map[string]interface{} `yaml:"configs" json:"configs"`
}

type OmegaConfig struct {
	Name                     string                 `yaml:"name" json:"name,omitempty"`
	Version                  string                 `yaml:"version" json:"version,omitempty"`
	ComponentsConfig         []*ComponentConfig     `yaml:"组件配置" json:"组件配置"`
	Trigger                  *defines.TriggerConfig `yaml:"触发词" json:"触发词"`
	CommandFeedBackByDefault bool                   `yaml:"默认情况下是否有commandfeedback" json:"默认情况下是否有commandfeedback,omitempty"`
}

func Policy_2(root string) {
	configFile := path.Join(root, "config.json")
	if !utils.IsFile(configFile) {
		return
	}
	if fp, err := os.OpenFile(configFile, os.O_RDONLY, 0644); err != nil {
		panic("位于 " + configFile + " 的配置文件有问题: " + err.Error())
	} else {
		fullConfig := &OmegaConfig{}
		err := json.NewDecoder(fp).Decode(&fullConfig)
		if err != nil {
			panic("位于 " + configFile + " 的配置文件有问题: " + err.Error())
		} else {
			policy_2_upgrade_config(fullConfig, root)
		}
		fp.Close()
	}
}

type Renamer struct {
	NewName  string
	FileName string
}

func getRenamers() map[string]Renamer {
	return map[string]Renamer{
		"Bonjour":               Renamer{"入服欢迎", ""},
		"ChatLogger":            Renamer{"聊天记录", ""},
		"Banner":                Renamer{"系统上线提示", ""},
		"FeedBack":              Renamer{"反馈信息", ""},
		"Memo":                  Renamer{"玩家留言", ""},
		"PlayerTP":              Renamer{"玩家互传", ""},
		"BackToHQ":              Renamer{"返回主城", ""},
		"SetSpawnPoint":         Renamer{"设置重生点", ""},
		"Respawn":               Renamer{"玩家自杀", ""},
		"AboutMe":               Renamer{"玩家信息", ""},
		"Portal":                Renamer{"自定义传送点", ""},
		"Immortal":              Renamer{"返回死亡点", ""},
		"Kick":                  Renamer{"踢人", ""},
		"Shop":                  Renamer{"商店", "商品清单文件"},
		"QGroupLink":            Renamer{"群服互通", ""},
		"Recycle":               Renamer{"物品回收", "回收清单文件"},
		"FakeOP":                Renamer{"OP权限模拟", "授权文件"},
		"SimpleCmd":             Renamer{"简单自定义指令", ""},
		"Schedule":              Renamer{"计划任务", ""},
		"TimeSync":              Renamer{"时间同步", ""},
		"MoneyTransfer":         Renamer{"玩家转账", ""},
		"StructureBackup":       Renamer{"自助建筑备份", ""},
		"Crash":                 Renamer{"同步退出", ""},
		"IntrusionDetectSystem": Renamer{"手持32k检测", ""},
		"WhoAreYou":             Renamer{"违规昵称检测", ""},
		"ContainerScan":         Renamer{"32k方块检测", ""},
		"OpCheck":               Renamer{"管理员检测", ""},
		"Menu":                  Renamer{"菜单显示", ""},
		"CmdSender":             Renamer{"指令发送", ""},
		"NoSQLDBUtil":           Renamer{"数据库导入导出工具", ""},
		"NameRecord":            Renamer{"改名记录", ""},
		"KeepAlive":             Renamer{"假死检测", ""},
	}
}

func policy_2_upgrade_config(fullConfig *OmegaConfig, root string) {
	newOmegaConfig := &defines.OmegaConfig{Name: fullConfig.Name, Version: fullConfig.Version, Trigger: fullConfig.Trigger, CommandFeedBackByDefault: fullConfig.CommandFeedBackByDefault}
	ComponentConfigs := []*defines.ComponentConfig{}
	renamers := getRenamers()
	filesToRemove := []string{path.Join(root, "config.json")}
	for _, cfg := range fullConfig.ComponentsConfig {
		Name := cfg.Name
		Version := cfg.Version
		Source := cfg.Source
		Description := cfg.Description
		Disabled := cfg.Disabled
		Cfg := cfg.Configs
		newName, addFile := "", ""
		{
			newName = renamers[Name].NewName
			addFile = renamers[Name].FileName
		}
		if addFile != "" {
			addConfigPath := path.Join(root, "data", Cfg[addFile].(string))
			if fp, err := os.OpenFile(addConfigPath, os.O_RDONLY, 0644); err != nil {
				panic("位于 " + addConfigPath + " 的配置文件有问题: " + err.Error())
			} else {
				var configData interface{}
				err := json.NewDecoder(fp).Decode(&configData)
				if err != nil {
					panic("位于 " + addConfigPath + " 的配置文件有问题: " + err.Error())
				} else {
					Cfg[addFile] = configData
					filesToRemove = append(filesToRemove, addConfigPath)
				}
				fp.Close()
			}
		}
		newConfig := &defines.ComponentConfig{
			Name:        newName,
			Version:     Version,
			Source:      Source,
			Description: Description,
			Disabled:    Disabled,
			Configs:     Cfg,
		}
		ComponentConfigs = append(ComponentConfigs, newConfig)
	}
	utils.MakeDirP(path.Join(root, "配置"))
	if err := utils.DeployOmegaConfig(newOmegaConfig, root); err != nil {
		panic(err)
	}
	if err := utils.DeployComponentConfigs(ComponentConfigs, root); err != nil {
		panic(err)
	}
	for _, fname := range filesToRemove {
		os.Remove(fname)
	}

}
