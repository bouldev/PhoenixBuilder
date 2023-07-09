package challenges

import (
	"context"
	"fmt"
	"phoenixbuilder/fastbuilder/lib/minecraft/neomega/omega"
	"phoenixbuilder/minecraft/protocol/packet"
	"strings"
	"time"
)

type OperatorChallenge struct {
	omega.MicroOmega
	hasOpPrivilege  bool
	cheatOn         bool
	lostPrivilegeCB func()
}

func NewOperatorChallenge(omega omega.MicroOmega, lostPrivilegeCallBack func()) *OperatorChallenge {
	if lostPrivilegeCallBack == nil {
		lostPrivilegeCallBack = func() {
			panic(fmt.Errorf("Operator privilege lost"))
		}
	}
	helper := &OperatorChallenge{
		MicroOmega:      omega,
		lostPrivilegeCB: lostPrivilegeCallBack,
	}
	omega.GetGameListener().SetOnTypedPacketCallBack(packet.IDAdventureSettings, helper.onAdventurePacket)
	return helper
}

func (o *OperatorChallenge) onAdventurePacket(pk packet.Packet) {
	p := pk.(*packet.AdventureSettings)
	if o.GetMicroUQHolder().GetBotBasicInfo().GetBotUniqueID() == p.PlayerUniqueID {
		if p.PermissionLevel >= packet.PermissionLevelOperator {
			o.hasOpPrivilege = true
			fmt.Println("Operator privilege granted")
		} else {
			if o.hasOpPrivilege {
				o.lostPrivilegeCB()
			}
			fmt.Println("Please grant operator privilege")
			o.hasOpPrivilege = false
		}
	}
}

func (o *OperatorChallenge) WaitForPrivilege(ctx context.Context) (err error) {
	for !o.hasOpPrivilege {
		o.GetGameControl().SendWSCmdAndInvokeOnResponse("tp @s ~~~", func(output *packet.CommandOutput) {
			if output.SuccessCount > 0 {
				o.hasOpPrivilege = true
			}
		})
		o.GetGameControl().BotSay("请给予机器人Op权限或检查作弊模式")
		fmt.Println("请给予机器人Op权限或检查作弊模式")
		time.Sleep(1 * time.Second)
		if ctx.Err() != nil {
			return fmt.Errorf("Operator privilege granting timed out")
		}
	}
	fmt.Println("Privilege granted")
	first := true
	for !o.cheatOn {
		o.GetGameControl().SendWSCmdAndInvokeOnResponse("testforblock ~~~ air 0", func(output *packet.CommandOutput) {
			if output.SuccessCount > 0 {
				o.cheatOn = true
			}
			if len(output.OutputMessages) > 0 {
				if strings.Contains(output.OutputMessages[0].Message, "commands.generic.disabled") {
					o.cheatOn = false
					if first {
						o.GetGameControl().BotSay("请打开作弊模式")
						fmt.Println("请打开作弊模式")
						first = false
					}
				} else {
					o.cheatOn = true
				}
			}
		})
		o.GetGameControl().SendWSCmdAndInvokeOnResponse("tp @s ~~~", func(output *packet.CommandOutput) {
			if output.SuccessCount > 0 {
				o.cheatOn = true
			}
		})
		time.Sleep(1 * time.Second)
		if ctx.Err() != nil {
			return fmt.Errorf("Ceased waiting")
		}
		if !o.cheatOn {
			o.GetGameControl().BotSay("Please enable cheating.")
			fmt.Println("Please enable cheating.")
		}
	}
	return nil
}
