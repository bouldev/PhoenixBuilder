package utils

import (
	"fmt"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	"strings"
)

func QueryBlockName(ctrl defines.GameControl, x, y, z int, onResult func(string)) {
	ctrl.SendCmdAndInvokeOnResponseWithFeedback(fmt.Sprintf("testforblock %v %v %v air", x, y, z), func(output *packet.CommandOutput) {
		if output.SuccessCount > 0 {
			onResult("air")
		} else {
			if len(output.OutputMessages) > 0 && len(output.OutputMessages[0].Parameters) == 5 {
				blkName := strings.Split(output.OutputMessages[0].Parameters[3], ".")
				if len(blkName) == 3 {
					onResult(blkName[1])
				} else {
					onResult("get_error")
				}
			} else {
				onResult("get_error")
			}
		}
	})
}
