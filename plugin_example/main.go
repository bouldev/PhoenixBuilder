package main

import (
	"phoenixbuilder/minecraft/function"
	"phoenixbuilder/minecraft"
	"fmt"
)

func plugin_main() string {
	fmt.Printf("Plugin example is loading, trying to register a simple function\n")
	function.RegisterFunction(&function.Function {
		Name: "Example Plugin Test",
		OwnedKeywords: []string { "example_plugin_test" },
		FunctionType: function.FunctionTypeSimple,
		SFMinSliceLen: 1,
		FunctionContent: func(conn *minecraft.Conn,_ []interface{}) {
			fmt.Printf("The function of example plugin is called!\n")
		},
	})
	fmt.Printf("Execute command: example_plugin_test to check.\n")
	return "example"
}