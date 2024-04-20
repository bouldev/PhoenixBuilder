package minecraft

import (
	"phoenixbuilder/fastbuilder/py_rpc/mod_event/client_to_server/minecraft/vip_event_system"
	mei "phoenixbuilder/fastbuilder/py_rpc/mod_event/interface"
)

// 网易会员系统
type VIPEventSystem struct{ mei.Module }

// Return the module name of v
func (v *VIPEventSystem) ModuleName() string {
	return "vipEventSystem"
}

// Return a pool/map that contains all the event of v
func (v *VIPEventSystem) EventPool() map[string]mei.Event {
	return map[string]mei.Event{
		"PlayerUiInit": &vip_event_system.PlayerUiInit{},
	}
}
