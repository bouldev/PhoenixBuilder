package mainframe

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"phoenixbuilder/omega/components"
	"phoenixbuilder/omega/defines"
	third_party_omega_components "phoenixbuilder/omega/third_party"
	"phoenixbuilder/omega/utils"
	"strings"

	"github.com/pterm/pterm"
)

const Version = "0.0.2"

//go:embed default_config.json
var defaultConfigBytes []byte

//go:embed default_components.json
var defaultComponentsBytes []byte

func (o *Omega) systemWideUpgrade(root string) (err error) {
	// omega system wide upgrade
	o.OmegaConfig = utils.CollectOmegaConfig(root)
	if o.OmegaConfig.MigrationVersion < 1202 {

		// move all dir with path root/第三方/name to root/第三方_by_name
		thirdPartyDir := path.Join(root, "配置", "第三方")
		thirdPartyByDir := path.Join(root, "配置", "第三方_by_")
		thirdPartyDirs, err := ioutil.ReadDir(thirdPartyDir)
		if err != nil {
			return err
		}

		for _, dir := range thirdPartyDirs {
			if dir.IsDir() {
				oldPath := path.Join(thirdPartyDir, dir.Name())
				newPath := thirdPartyByDir + dir.Name()
				if err := utils.MoveDir(oldPath, newPath); err != nil {
					return err
				}
			}
		}
		// if dir root/第三方/ is empty, delete it
		if utils.IsDirEmpty(thirdPartyDir) {
			if err := os.Remove(thirdPartyDir); err != nil {
				return err
			}
		}

		o.OmegaConfig.MigrationVersion = 1202
		if err := utils.DeployOmegaConfig(o.OmegaConfig, root); err != nil {
			panic(err)
		}
	}
	return nil
}

func (o *Omega) checkAndLoadConfig() {
	defer func() {
		r := recover()
		if r != nil {
			pterm.Error.Printfln("配置文件有问题 %v", r)
			pterm.Error.Println("错误的修改了配置文件或者使用windows记事本打开配置文件都可能导致这种错误")
			pterm.Warning.Println("但是，不用担心，你总是可以通过删除故障的配置文件使Omega恢复工作！")
			fullp, _ := filepath.Abs(path.Join(o.storageRoot, "配置"))
			pterm.Warning.Printfln("配置文件应该位于", fullp, "文件夹中")
			panic(fmt.Sprintf("请修正配置文件(%v)", r))
		}
	}()
	root := o.storageRoot
	if !utils.IsDir(root) {
		utils.MakeDirP(root)
	}
	if !utils.IsDir(path.Join(root, "配置")) {
		utils.MakeDirP(path.Join(root, "配置"))
	}
	if !utils.HasOmegaConfig(root) {
		c := &defines.OmegaConfig{}
		if err := json.Unmarshal(defaultConfigBytes, &c); err != nil {
			panic(err)
		}
		if err := utils.DeployOmegaConfig(c, root); err != nil {
			panic(err)
		}
	}

	err := o.systemWideUpgrade(root)
	if err != nil {
		pterm.Error.Printfln("系统配置升级失败: %v", err)
	}

	{
		existComponentConfigs := utils.CollectComponentConfigs(root)
		if len(existComponentConfigs) == 0 {
			// unpack all default configs
			if err := json.Unmarshal(defaultComponentsBytes, &existComponentConfigs); err != nil {
				panic(err)
			}
			if err := utils.DeployComponentConfigs(existComponentConfigs, root); err != nil {
				panic(err)
			}
			thirdPartConfigs := make([]*defines.ComponentConfig, 0)
			for _, g := range third_party_omega_components.GetAllThirdPartComponents() {
				for _, c := range g.DefaultConfigs {
					thirdPartConfigs = append(thirdPartConfigs, c)
				}
			}
			if err := utils.DeployComponentConfigs(thirdPartConfigs, root); err != nil {
				panic(err)
			}
		} else {
			// check for new configs
			defaultComponentConfigs := []*defines.ComponentConfig{}
			newComponentConfigs := []*defines.ComponentConfig{}
			groupedDefaultConfigs := map[string][]*defines.ComponentConfig{}
			if err := json.Unmarshal(defaultComponentsBytes, &defaultComponentConfigs); err != nil {
				panic(err)
			}
			for _, c := range defaultComponentConfigs {
				if groupedDefaultConfigs[c.Name] == nil {
					groupedDefaultConfigs[c.Name] = []*defines.ComponentConfig{c}
				} else {
					groupedDefaultConfigs[c.Name] = append(groupedDefaultConfigs[c.Name], c)
				}
			}
			for _, g := range third_party_omega_components.GetAllThirdPartComponents() {
				for _, c := range g.DefaultConfigs {
					if groupedDefaultConfigs[c.Name] == nil {
						groupedDefaultConfigs[c.Name] = []*defines.ComponentConfig{c}
					} else {
						groupedDefaultConfigs[c.Name] = append(groupedDefaultConfigs[c.Name], c)
					}
				}
			}
			for _, c := range existComponentConfigs {
				if groupedDefaultConfigs[c.Name] != nil {
					delete(groupedDefaultConfigs, c.Name)
				}
			}
			for _, group := range groupedDefaultConfigs {
				for _, c := range group {
					if c.Source == "Core" {
						pterm.Success.Println("有新核心组件 " + c.Name + " 可用，已自动加入配置并[启用]")
						c.Disabled = false
					} else if c.Source == "Built-In" {
						pterm.Success.Println("有新内置组件 " + c.Name + " 可用，已自动加入配置并[关闭]，请前往 omega_storage/配置/" + c.Name + " 打开")
						c.Disabled = true
					} else if strings.HasPrefix(c.Source, "第三方::") {
						pterm.Success.Println("有新第三方组件 " + c.Name + " 可用，已自动加入配置并[关闭]，请前往 omega_storage/配置/" + c.Name + " 打开")
						c.Disabled = true
					}
					newComponentConfigs = append(newComponentConfigs, c)
				}
			}
			if err := utils.DeployComponentConfigs(newComponentConfigs, root); err != nil {
				panic(err)
			}
			if len(newComponentConfigs) > 0 {
				pterm.Warning.Println("组件已变更...将重新加载")
				existComponentConfigs = utils.CollectComponentConfigs(root)
			}
		}
		// fix source
		componentsSource := map[string]string{}
		for name, _ := range getCoreComponentsPool() {
			componentsSource[name] = "Core"
		}
		for name, _ := range components.GetComponentsPool() {
			componentsSource[name] = "Built-In"
		}
		for _, c := range existComponentConfigs {
			if source, found := componentsSource[c.Name]; found {
				if source != c.Source {
					c.Source = source
					pterm.Error.Printfln("组件 %v 来源信息错误，现在已经更正为 %v", c.Name, c.Source)
					c.Upgrade()
				}
				if source == "Core" && c.Disabled {
					c.Disabled = false
					pterm.Error.Printfln("核心组件 %v 不可被禁用，现在已经打开了", c.Name)
					c.Upgrade()
				}
			}
		}
	}
	componentConfigs := utils.CollectComponentConfigs(root)
	preferredOrder := map[string]int{
		"假死检测":  0,
		"返回主城":  1,
		"改名记录":  2,
		"设置重生点": 3,
		"玩家自杀":  4,
	}
	thirdPart := make([]*defines.ComponentConfig, 0)
	groupedOrder := make([][]*defines.ComponentConfig, len(preferredOrder)+1)
	for i := range groupedOrder {
		groupedOrder[i] = make([]*defines.ComponentConfig, 0)
	}
	defaultI := len(preferredOrder)
	for _, c := range componentConfigs {
		if gi, hask := preferredOrder[c.Name]; hask {
			groupedOrder[gi] = append(groupedOrder[gi], c)
		} else if strings.HasPrefix(c.Name, "第三方::") {
			thirdPart = append(thirdPart, c)
		} else {
			groupedOrder[defaultI] = append(groupedOrder[defaultI], c)
		}
	}
	reorderedConfig := []*defines.ComponentConfig{}
	for _, group := range groupedOrder {
		reorderedConfig = append(reorderedConfig, group...)
	}
	reorderedConfig = append(reorderedConfig, thirdPart...)
	o.ComponentConfigs = reorderedConfig
}
