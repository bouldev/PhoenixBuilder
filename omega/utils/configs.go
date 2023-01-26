package utils

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"phoenixbuilder/omega/defines"
	"runtime"
	"strings"

	"github.com/pterm/pterm"
)

func DeployComponentConfigs(ComponentConfigs []*defines.ComponentConfig, root string) error {
	groupedConfigs := map[string][]*defines.ComponentConfig{}
	for _, cfg := range ComponentConfigs {
		if groupedConfigs[cfg.Name] == nil {
			groupedConfigs[cfg.Name] = make([]*defines.ComponentConfig, 0)
		}
		groupedConfigs[cfg.Name] = append(groupedConfigs[cfg.Name], cfg)
	}
	for _, cfgs := range groupedConfigs {
		// dir := path.Join(root, "配置")
		name := strings.ReplaceAll(cfgs[0].Name, "::", "__")
		subFolders := strings.Split(cfgs[0].Name, "::")
		if len(subFolders) == 3 {
			s := []string{strings.Join(subFolders[:2], "_by_")}
			subFolders = append(s, subFolders[2:]...)
		}
		// if len(cfgs) > 1 || IsDir(path.Join(root, "配置", name)) {
		dir := path.Join(root, "配置", path.Join(subFolders...))
		os.MkdirAll(dir, 0755)
		// }
		for i, cfg := range cfgs {
			p := path.Join(dir, fmt.Sprintf("组件-%v-%v.json", name, i+1))
			if err := WriteJsonData(p, cfg); err != nil {
				return err
			}
		}
	}
	return nil
}

func CollectComponentConfigs(root string) (ComponentConfigs []*defines.ComponentConfig) {
	d := path.Join(root, "配置")
	_, err := ioutil.ReadDir(d)
	if err != nil {
		panic("读取[" + d + "]时出错" + err.Error())
	}
	ComponentConfigs = []*defines.ComponentConfig{}
	if err := filepath.Walk(d, func(filePath string, info fs.FileInfo, err error) error {
		if runtime.GOOS == "windows" {
			filePath = strings.ReplaceAll(filePath, "\\", "/")
		}
		if info.IsDir() {
			return nil
		}
		fileBaseName := path.Base(filePath)
		if !strings.HasPrefix(fileBaseName, "组件") || !strings.HasSuffix(fileBaseName, ".json") {
			return nil
		}
		c := &defines.ComponentConfig{}
		attachments := make(map[string]map[string]string)
		if attachments, err = GetJsonDataWithAttachment(filePath, c); err != nil {
			return fmt.Errorf("处理[" + filePath + "]时出错" + err.Error())
		}
		c.SetAttachedField(attachments["fields"])
		c.SetUpgradeFn(func(cc *defines.ComponentConfig) error {
			pterm.Info.Println("正在升级配置: ", filePath)
			return WriteJsonDataWithAttachment(filePath, cc, attachments)
		})
		ComponentConfigs = append(ComponentConfigs, c)
		return nil
	}); err != nil {
		panic(err)
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
