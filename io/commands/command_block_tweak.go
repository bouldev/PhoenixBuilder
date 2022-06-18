// +build is_tweak

package commands

import (
	"phoenixbuilder/fastbuilder/types"
)

/*
// All strings are callee free
void phoenixbuilder_update_command_block(int x,int y,int z,unsigned int mode,char *command,char *customName,char *lastOutput,int tickDelay,char executeOnFirstTick,char trackOutput,char conditional,char needsRedstone);
*/
import "C"

func booltochar(gobool bool) C.char {
	if gobool {
		return C.char(1)
	}
	return C.char(0)
}

func (sender *CommandSender) UpdateCommandBlock(x int32,y int32,z int32,d *types.CommandBlockData) {
	C.phoenixbuilder_update_command_block(C.int(x),C.int(y),C.int(z),
					C.uint(d.Mode), C.CString(d.Command), C.CString(d.CustomName),
					C.CString(d.LastOutput), C.int(d.TickDelay),
					booltochar(d.ExecuteOnFirstTick),
					booltochar(d.TrackOutput), booltochar(d.Conditional),
					booltochar(d.NeedRedstone))
}