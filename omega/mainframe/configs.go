package mainframe

import (
	_ "embed"
	"encoding/json"
	"path"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
)

const Version = "0.0.2"

//go:embed default_config.json
var defaultConfigBytes []byte

//go:embed default_components.json
var defaultComponentsBytes []byte

func (o *Omega) readConfig() {
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
	omegaConfig := utils.CollectOmegaConfig(root)
	componentConfigs := []*defines.ComponentConfig{}
	if componentConfigs = utils.CollectComponentConfigs(root); len(componentConfigs) == 0 {
		if err := json.Unmarshal(defaultComponentsBytes, &componentConfigs); err != nil {
			panic(err)
		}
		if err := utils.DeployComponentConfigs(componentConfigs, root); err != nil {
			panic(err)
		}
	}
	o.OmegaConfig = omegaConfig
	o.ComponentConfigs = componentConfigs
}
