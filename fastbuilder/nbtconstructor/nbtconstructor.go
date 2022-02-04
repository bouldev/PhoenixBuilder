package nbtconstructor

import (
	"phoenixbuilder/fastbuilder/function"
	"phoenixbuilder/minecraft"
	"encoding/json"
	"bufio"
	"fmt"
)

func evalNBT(content string) {
	decoder:=json.NewDecoder(bufio.NewReader(content))
	c
}

func InitNBTConstructor() {
	function.RegisterFunction(&function.Function {
		Name: "constructItem",
		OwnedKeywords: []string{"simpleconstruct"},
		FunctionType: function.FunctionTypeSimple,
		SFArgumentTypes: []byte { function.SimpleFunctionArgumentMessage },
		SFMinSliceLen: 1,
		FunctionContent: func(conn *minecraft.Conn, args []interface{}) {
			toEval:=args[0].(string)
			fmt.Printf("%s\n",toEval)
		},
	})
}