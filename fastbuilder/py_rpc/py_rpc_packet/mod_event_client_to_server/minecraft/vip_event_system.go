package minecraft

import (
	"phoenixbuilder/fastbuilder/py_rpc/py_rpc_packet/interface/mod_event"
	"phoenixbuilder/fastbuilder/py_rpc/py_rpc_packet/mod_event_client_to_server/minecraft/vip_event_system"
)

// 网易会员系统
type VIPEventSystem struct{ mod_event.Module }

// Return the module name of v
func (v *VIPEventSystem) ModuleName() string {
	return "vipEventSystem"
}

// Return a pool/map that contains all the event of v
func (v *VIPEventSystem) EventPool() map[string]mod_event.Event {
	return map[string]mod_event.Event{
		"PlayerUiInit": &vip_event_system.PlayerUiInit{},
	}
}
