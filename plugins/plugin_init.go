package main

import (
	"plugin_example/plugin_structs"
	"unsafe"
)

func PluginInit(bridgeif unsafe.Pointer,mainref interface{}) string {
	bridge:=*(*plugin_structs.PluginBridge)(bridgeif)
	mainfunc:=mainref.(func(plugin_structs.PluginBridge)string)
	return mainfunc(bridge)
}