package mainframe

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/pterm/pterm"
	"gopkg.in/yaml.v3"
	"os"
	"path"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
)

const Version = "0.0.1"

type TriggerConfig struct {
	DefaultTigger     string   `yaml:"default_tigger" json:"default_tigger"`
	TriggerWords      []string `yaml:"system_trigger_words" json:"trigger_words"`
	AllowNoSpace      bool     `yaml:"allow_no_space" json:"allow_no_space"`
	RemoveSuffixColor bool     `yaml:"remove_suffix_color" json:"remove_suffix_color"`
	AllowWisper       bool     `yaml:"allow_wisper" json:"allow_wisper"`
}

type OmegaConfig struct {
	Name             string                     `yaml:"name" json:"name,omitempty"`
	Version          string                     `yaml:"version" json:"version,omitempty"`
	ComponentsConfig []*defines.ComponentConfig `yaml:"components" json:"components_config,omitempty"`
	//configByName             map[string]*defines.ComponentConfig
	Trigger                  *TriggerConfig `yaml:"trigger" json:"trigger,omitempty"`
	CommandFeedBackByDefault bool           `yaml:"expeted_commandfeedback" json:"command_feed_back_by_default,omitempty"`
}

//go:embed default_config.json
var defaultConfigBytes []byte

func newConfig() *OmegaConfig {
	c := &OmegaConfig{
		Name:             "Omega",
		Version:          Version,
		ComponentsConfig: []*defines.ComponentConfig{},
		//configByName:     map[string]*defines.ComponentConfig{},
		Trigger: &TriggerConfig{},
	}
	return c
}

func (c *OmegaConfig) Check() error {
	if c.Name != "Omega" {
		panic(pterm.Error.Sprintln("配置文件并非为 Omega 系统设计"))
	}
	if c.Version != Version {
		panic("配置文件版本: " + c.Version + " 与系统版本: " + Version + " 冲突")
	}

	if c.Trigger.DefaultTigger == "" {
		panic(pterm.Error.Sprintln("没有默认触发词，如果你不想要触发词，你应该将' '添加到system_trigger_words中"))
	}
	flag := false
	for _, trig := range c.Trigger.TriggerWords {
		if trig == c.Trigger.DefaultTigger {
			flag = true
			break
		}
	}
	if !flag {
		c.Trigger.TriggerWords = append(c.Trigger.TriggerWords, c.Trigger.DefaultTigger)
	}
	for _, cfg := range c.ComponentsConfig {
		name := cfg.Name
		if name == "" {
			panic(fmt.Errorf("组件名不能为空 %v", cfg))
		}
	}
	return nil
}

func (c *OmegaConfig) QueryConfig(topic string) interface{} {
	if topic == "DefaultTigger" {
		return c.Trigger.DefaultTigger
	} else if topic == "TriggerWords" {
		return c.Trigger.TriggerWords
	}
	return ""
}

func (o *Omega) writeBackConfig() {
	configFile := path.Join(o.storageRoot, "config.yaml")
	//fmt.Println("Updating Config File" + configFile)
	if fp, err := os.OpenFile(configFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644); err != nil {
		panic(err)
	} else {
		encoder := yaml.NewEncoder(fp)
		//encoder.SetIndent("", "\t")
		err := encoder.Encode(o.fullConfig)
		if err != nil {
			panic(err)
		}
		fp.Close()
	}
}

func (o *Omega) genMap() {
	for _, cfg := range o.fullConfig.ComponentsConfig {
		name := cfg.Name
		if name == "" {
			panic(fmt.Errorf("组件名不能为空 %v", cfg))
		}
		//if oldcfg,hasK:=o.fullConfig.configByName[name];hasK{
		//	panic(fmt.Errorf("存在同名组件 %v",cfg))
		//}
	}
}

func (o *Omega) readConfig() {
	configFile := path.Join(o.storageRoot, "config.json")
	if !utils.IsFile(configFile) || true {
		fmt.Println("未发现配置文件，展开默认配置文件")
		if fp, err := os.OpenFile(configFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644); err != nil {
			panic(err)
		} else {
			if _, err := fp.Write(defaultConfigBytes); err != nil {
				panic(err)
			}
			fp.Close()
		}
	}
	{
		o.fullConfig = newConfig()
		if fp, err := os.OpenFile(configFile, os.O_RDONLY, 0644); err != nil {
			panic(err)
		} else {
			err := json.NewDecoder(fp).Decode(o.fullConfig)
			if err != nil {
				panic(err)
			}
			fp.Close()
		}
		if err := o.fullConfig.Check(); err != nil {
			panic(err)
		}
	}
}
