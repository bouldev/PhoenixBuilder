// +build !with_v8

package script_kickstarter

import "phoenixbuilder/fastbuilder/script_engine/bridge"
import "fmt"

func LoadScript(scriptPath string, hb bridge.HostBridge) (func(),error) {
	//panic("LoadScript() called with no v8 linked.")
	return func(){},fmt.Errorf("Scripts are not available for non-v8-linked versions.")
}