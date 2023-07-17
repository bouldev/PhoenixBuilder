package main

import (
	"context"
	"fmt"
	"path"
	"phoenixbuilder/fastbuilder/lib/utils/crypto_wrapper"
	"phoenixbuilder/fastbuilder/lib/utils/download_wrapper"
	"phoenixbuilder/fastbuilder/lib/utils/file_wrapper"
	"phoenixbuilder/solutions/omega_lua/omega_lua/lua_utils"
)

type ComponentConfig struct {
	Name              string `json:"名称"`
	Description       string `json:"描述"`
	Disabled          bool   `json:"是否禁用"`
	Version           string `json:"版本"`
	Source            string `json:"来源"`
	RemoteFileAddress string `json:"远程地址"`
	RemoteHashAddress string `json:"远程哈希地址"`
	Configs           any    `json:"配置"`
}

type CombinedLuaComponent struct {
	Config       any
	Code         string
	ConfigUpdate func(any)
}

func LoadLuaConfigFromFile(ctx context.Context, localLuaComponentsDir, configFle string) (c *CombinedLuaComponent, err error) {
	fullCfg := &ComponentConfig{}
	if err := file_wrapper.GetJsonData(configFle, &fullCfg); err != nil {
		return nil, err
	}
	localFileName := path.Join(localLuaComponentsDir, fullCfg.Name)

	upgradeUrl := ""
	if fullCfg.RemoteHashAddress != "" {
		if fullCfg.RemoteFileAddress == "" {
			return nil, fmt.Errorf("remote hash address is set, but remote file address is not")
		}
		// check local file
		data, _ := file_wrapper.GetFileData(localFileName)
		localFileHash := crypto_wrapper.BytesMD5Str(data)
		remoteHash, err := download_wrapper.DownloadMicroContentWithCtx(ctx, fullCfg.RemoteHashAddress)
		if err != nil {
			return nil, err
		}
		if localFileHash != string(remoteHash) {
			upgradeUrl = fullCfg.RemoteFileAddress
		}
	} else if fullCfg.RemoteFileAddress != "" {
		upgradeUrl = fullCfg.RemoteFileAddress
	}

	if upgradeUrl == "" && !file_wrapper.Exists(localFileName) {
		return nil, fmt.Errorf("no local nor remote file not exists")
	}

	fileData := ""

	if upgradeUrl != "" {
		if data, err := download_wrapper.DownloadMicroContentWithCtx(ctx, upgradeUrl); err != nil {
			return nil, err
		} else {
			file_wrapper.WriteFile(localFileName, data, 0755)
			fileData = string(data)
		}
	} else {
		data, err := file_wrapper.GetFileData(localFileName)
		if err != nil {
			return nil, err
		}
		fileData = string(data)
	}
	return &CombinedLuaComponent{
		Code:   fileData,
		Config: fullCfg.Configs,
		ConfigUpdate: func(newConfig any) {
			newConfig = lua_utils.CleanMapKey(newConfig)
			fullCfg.Configs = newConfig
			err := file_wrapper.WriteJsonData(configFle, fullCfg)
			if err != nil {
				panic(err)
			}
		},
	}, nil
}

func LoadLuaConfigsFromFile(configFiles []string, localLuaComponentsDir string) (configs []*CombinedLuaComponent, err error) {
	ctx, cancelFn := context.WithCancel(context.Background())
	errChan := make(chan error)
	luaComponentsChan := make(chan *CombinedLuaComponent, len(configFiles))

	for _, configFile := range configFiles {
		go func(cf string) {
			if c, err := LoadLuaConfigFromFile(ctx, localLuaComponentsDir, cf); err != nil {
				cancelFn()
				errChan <- err
			} else {
				luaComponentsChan <- c
			}

		}(configFile)
	}

	allDone := make(chan struct{})
	configs = []*CombinedLuaComponent{}
	go func() {
		for i := 0; i < len(configFiles); i++ {
			select {
			case err := <-errChan:
				errChan <- err
				return
			case c := <-luaComponentsChan:
				configs = append(configs, c)
			}
		}
		close(allDone)
	}()

	select {
	case err := <-errChan:
		return nil, err
	case <-allDone:
		return configs, nil
	}
}
