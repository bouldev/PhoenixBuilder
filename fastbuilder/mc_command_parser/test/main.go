package main

import "C"
import (
	NBTAssigner "phoenixbuilder/fastbuilder/bdump/nbt_assigner"

	"github.com/pterm/pterm"
)

//export UEC
func UEC(cmd *C.char) {
	new, warn, err := NBTAssigner.UpgradeExecuteCommand(C.GoString(cmd))
	if err != nil {
		pterm.Error.Printf(
			"UEC: Conversion failure; C.GoString(cmd) = %#v, err = %v\n",
			C.GoString(cmd),
			err,
		)
	} else if len(warn) > 0 {
		pterm.Warning.Printf(
			"UEC: The mapping of the block data value to the block state was not found in some detect fields; failure_blocks = %#v, C.GoString(cmd) = %#v; err = %v\n",
			C.GoString(cmd),
			warn,
			err,
		)
	} else if new != C.GoString(cmd) {
		pterm.Success.Printf(
			"UEC: Successful to upgrade; C.GoString(cmd) = %#v, new = %#v\n",
			C.GoString(cmd),
			new,
		)
	} else {
		pterm.Info.Printf(
			"UEC: The commands have not been changed; C.GoString(cmd) = %#v\n",
			C.GoString(cmd),
		)
	}
}

func main() {
	// windows: go build -o upgrade_execute_commands_test.dll -buildmode=c-shared fastbuilder/mc_command_parser/test/main.go
}
