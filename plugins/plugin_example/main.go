package main

import (
	"plugin_example/plugin_structs"
	"fmt"
)

func Main(bridge plugin_structs.PluginBridge) string {
	fmt.Printf("Plugin example is loading, trying to register a simple function\n")
	bridge.RegisterFunction(plugin_structs.Function {
		Name: "Example Plugin Test",
		OwnedKeywords: []string { "example_plugin_test" },
		FunctionType: plugin_structs.FunctionTypeSimple,
		SFMinSliceLen: 1,
		FunctionContent: func(_ interface{},_ []interface{}) {
			fmt.Printf("The function of example plugin is called!\n")
			bridge.Tellraw("Tellraw test from example_plugin_test")
		},
	})
	bridge.SubscribeChat(func(sender string, content string) {
		fmt.Printf("Chat event: <%s> %s\n",sender,content)
	})
	fmt.Printf("Execute command: example_plugin_test to check.\n")
	return "example"
}