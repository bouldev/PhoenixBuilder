package upgrade

import (
	"fmt"
	"io/ioutil"
	"path"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"strings"
)

func checkMigrationVersion(root string) (int, error) {
	d := path.Join(root, "配置")
	entries, err := ioutil.ReadDir(d)
	if err != nil {
		return 0, err
	}
	for _, entry := range entries {
		if entry.Name() == "主系统.json" {
			p := path.Join(root, "配置", entry.Name())
			c := &defines.OmegaConfig{}
			if err := utils.GetJsonData(p, c); err != nil {
				return 0, fmt.Errorf("读取[" + p + "]时出错" + err.Error())
			}
			return c.MigrationVersion, nil
		}
	}
	return 0, fmt.Errorf("配置/主系统.json 未找到，配置文件夹可能损坏了")
}

func setMigrationVersion(root string, version int) error {
	d := path.Join(root, "配置")
	entries, err := ioutil.ReadDir(d)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.Name() == "主系统.json" {
			p := path.Join(root, "配置", entry.Name())
			c := &defines.OmegaConfig{}
			if err := utils.GetJsonData(p, c); err != nil {
				return fmt.Errorf("读取[" + p + "]时出错" + err.Error())
			}
			c.MigrationVersion = version
			return utils.WriteJsonData(p, c)
		}
	}
	return fmt.Errorf("配置/主系统.json 未找到，配置文件夹可能损坏了")
}

func updateComponentConfig(root string, name string, handleFn func(c *defines.ComponentConfig)) error {
	d := path.Join(root, "配置")
	entries, err := ioutil.ReadDir(d)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), "组件") {
			continue
		}
		p := path.Join(root, "配置", entry.Name())
		c := &defines.ComponentConfig{}
		if err := utils.GetJsonData(p, c); err != nil {
			return fmt.Errorf("读取[" + p + "]时出错" + err.Error())
		}
		if c.Name != name {
			continue
		}
		handleFn(c)
		return utils.WriteJsonData(p, c)
	}
	return fmt.Errorf("配置/组件-" + name + ".json 未找到，配置文件夹可能损坏了")
}
