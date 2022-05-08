package utils

import (
	"fmt"
	"io/ioutil"
	"path"
	"phoenixbuilder/omega/defines"
	"strings"
)

func DeployComponentConfigs(ComponentConfigs []*defines.ComponentConfig, root string) error {
	counter := map[string]int{}
	for _, cfg := range ComponentConfigs {
		i := 1
		if c, hasK := counter[cfg.Name]; hasK {
			counter[cfg.Name] = c + 1
			i = c + 1
		} else {
			counter[cfg.Name] = 1
		}
		p := path.Join(root, "配置", fmt.Sprintf("组件-%v-%v.json", cfg.Name, i))
		if err := WriteJsonData(p, cfg); err != nil {
			return err
		}
	}
	return nil
}

func CollectComponentConfigs(root string) (ComponentConfigs []*defines.ComponentConfig) {
	d := path.Join(root, "配置")
	entries, err := ioutil.ReadDir(d)
	if err != nil {
		panic("读取[" + d + "]时出错" + err.Error())
	}
	ComponentConfigs = []*defines.ComponentConfig{}
	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), "组件") {
			continue
		}
		p := path.Join(root, "配置", entry.Name())
		c := &defines.ComponentConfig{}
		if err := GetJsonData(p, c); err != nil {
			panic("处理[" + p + "]时出错" + err.Error())
		}
		ComponentConfigs = append(ComponentConfigs, c)
	}
	return ComponentConfigs
}

func DeployOmegaConfig(cfg *defines.OmegaConfig, root string) error {
	p := path.Join(root, "配置", "主系统.json")
	return WriteJsonData(p, cfg)
}

func HasOmegaConfig(root string) bool {
	p := path.Join(root, "配置", "主系统.json")
	return IsFile(p)
}

func CollectOmegaConfig(root string) *defines.OmegaConfig {
	p := path.Join(root, "配置", "主系统.json")
	cfg := &defines.OmegaConfig{}
	if err := GetJsonData(p, cfg); err != nil {
		panic("处理[" + p + "]时出错" + err.Error())
	}
	return cfg
}
