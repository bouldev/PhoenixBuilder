package main

import (
	"fmt"
	NBTAssigner "phoenixbuilder/fastbuilder/bdump/nbt_assigner"
)

func main() {
	fmt.Println(
		NBTAssigner.UpgradeExecuteCommand("execute@s~1.0"),
	)
	//core.Bootstrap()
}
