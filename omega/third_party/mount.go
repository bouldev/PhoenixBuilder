package mount

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/third_party/Happy2018new"
	// "phoenixbuilder/omega/third_party/example"
	"phoenixbuilder/omega/third_party/liliya"
	"phoenixbuilder/omega/third_party/sunlife"
)

type NameSpace string

type ComponentsGroup struct {
	NameSpace
	Components     map[string]defines.Component
	DefaultConfigs map[string]*defines.ComponentConfig
}

// 插件组
var componentsGroups map[NameSpace]ComponentsGroup

func GetAllThirdPartComponents() map[NameSpace]ComponentsGroup {
	if componentsGroups == nil {
		componentsGroups = make(map[NameSpace]ComponentsGroup)
	}
	return componentsGroups
}

func mountComponents(nameSpace NameSpace, Components map[string]defines.Component, DefaultConfigsData []byte) {
	nameSpace = "第三方::" + nameSpace
	if componentsGroups == nil {
		componentsGroups = make(map[NameSpace]ComponentsGroup)
	}
	if _, found := componentsGroups[nameSpace]; found {
		panic(fmt.Errorf("name space %v already occupied", nameSpace))
	}
	DefaultConfigs := []defines.ComponentConfig{}
	err := json.Unmarshal(DefaultConfigsData, &DefaultConfigs)
	if err != nil {
		panic(err)
	}
	DefaultConfigsMap := map[string]*defines.ComponentConfig{}
	for _, _cfg := range DefaultConfigs {
		cfg := _cfg
		cfg.Disabled = true
		cfg.Source = string(nameSpace)
		cfg.Name = string(nameSpace) + "::" + cfg.Name
		DefaultConfigsMap[cfg.Name] = &cfg
	}
	RenamedComponents := map[string]defines.Component{}
	for name, component := range Components {
		RenamedComponents[string(nameSpace)+"::"+name] = component
	}
	componentsGroups[nameSpace] = ComponentsGroup{
		NameSpace:      nameSpace,
		Components:     RenamedComponents,
		DefaultConfigs: DefaultConfigsMap,
	}
}

func init() {
	// mountComponents(example.NAMESPACE, example.Components, example.DefaultComponentConfigByte)
	mountComponents(liliya.NAMESPACE, liliya.Components, liliya.DefaultComponentConfigByte)
	mountComponents(Happy2018new.NAMESPACE, Happy2018new.Components, Happy2018new.DefaultComponentConfigByte)
	mountComponents(sunlife.NAMESPACE, sunlife.Components, sunlife.DefaultComponentConfigByte)
}
