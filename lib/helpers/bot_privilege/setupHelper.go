package bot_privilege

import (
	"fmt"
	"phoenixbuilder/lib/minecraft/neomega/omega"
	"phoenixbuilder/minecraft/protocol/packet"
	"strings"
	"time"
)

type SetupHelper struct {
	omega.MicroOmega
	hasOpPrivilege bool
}

func NewSetupHelper(omega omega.MicroOmega) *SetupHelper {
	helper := &SetupHelper{
		MicroOmega: omega,
	}
	omega.GetGameListener().SetOnTypedPacketCallBack(packet.IDAdventureSettings, helper.onAdventurePacket)
	return helper
}

func (o *SetupHelper) onAdventurePacket(pk packet.Packet) {
	p := pk.(*packet.AdventureSettings)
	if o.GetBotInfo().GetBotUniqueID() == p.PlayerUniqueID {
		if p.PermissionLevel >= packet.PermissionLevelOperator {
			o.hasOpPrivilege = true
			fmt.Println("机器人已获得管理员权限")
		} else {
			fmt.Println("请给予机器人管理员权限")
			if o.hasOpPrivilege {
				o.lostPrivilege()
			}
			o.hasOpPrivilege = false
		}
	}
}

func (o *SetupHelper) lostPrivilege() {
	panic(fmt.Errorf("机器人OP权限被关闭"))
}

func (o *SetupHelper) WaitOK() {
	time.Sleep(3 * time.Second)
	for !o.hasOpPrivilege {
		o.GetGameControl().BotSay("请给予机器人管理员权限")
		time.Sleep(1 * time.Second)
	}
	cheatOn := false
	first := true
	for !cheatOn {
		o.GetGameControl().SendWSCmdAndInvokeOnResponse("testforblock ~~~ air 0", func(output *packet.CommandOutput) {
			// fmt.Println(output)
			if len(output.OutputMessages) > 0 {
				if strings.Contains(output.OutputMessages[0].Message, "commands.generic.disabled") {
					cheatOn = false
					if first {
						fmt.Println("请打开作弊模式")
						first = false
					}
				} else {
					fmt.Println("作弊模式已经打开")
					cheatOn = true
				}
			}
		})
		time.Sleep(3 * time.Second)
		if !cheatOn {
			o.GetGameControl().BotSay("请打开作弊模式")
		}
	}

}
