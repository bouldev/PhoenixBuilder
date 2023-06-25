package omgApi

import (
	"phoenixbuilder/fastbuilder/uqHolder"
	"phoenixbuilder/omega/defines"
)

// 完成了一个lua组件所需要omg的全部东西
type OmgApi struct {
	MainFrame defines.MainFrame
	Omega     OmgCoreComponent
}

// 很显然我们是需要是corecomponent的接口
// 但是暂时定义一个简单的接口、
// 用于完成各种各样的操作
type OmgCoreComponent interface {
	QuerySensitiveInfo(key defines.SensitiveInfoType) (string, error)
	GetGlobalContext(key string) (entry interface{})
	SetGlobalContext(key string, entry interface{})
	GetUQHolder() *uqHolder.UQHolder
	GetWorldsDir(elem ...string) string
	GetOmegaNormalCacheDir(elem ...string) string
	GetAllConfigs() []*defines.ComponentConfig
	GetOmegaConfig() *defines.OmegaConfig
	GetPath(elem ...string) string
	GetStorageRoot() string
	GetRelativeFileName(topic string) string
	GetLogger(topic string) defines.LineDst
	GetFileData(topic string) ([]byte, error)
	WriteFileData(topic string, data []byte) error
	WriteJsonData(topic string, data interface{}) error
	WriteJsonDataWithTMP(topic string, tmpSuffix string, data interface{}) error
	GetJsonData(topic string, ptr interface{}) error
	GetBackendDisplay() defines.LineDst
	SetBackendCmdInterceptor(fn func(cmds []string) (stop bool))
	SetBackendMenuEntry(entry *defines.BackendMenuEntry)
}

// 第一个参数为omg的框架本体 第二个为mainframe
func NewOmgCoreComponent(occ OmgCoreComponent, mainframe defines.MainFrame) *OmgApi {
	omgapi := &OmgApi{
		MainFrame: mainframe,
		Omega:     occ,
	}

	return omgapi
}
