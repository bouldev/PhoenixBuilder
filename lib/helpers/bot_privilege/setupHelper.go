package bot_privilege

import (
	"errors"
	"fmt"
	"phoenixbuilder/lib/minecraft/neomega/omega"
	"phoenixbuilder/minecraft/protocol/packet"
	"strings"
	"time"
)

type SetupHelper struct {
	omega.MicroOmega
	hasOpPrivilege bool
	cheatOn        bool
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

var ErrMaximumWaitTimeExeceed = errors.New("maximum wait time execeed")

func (o *SetupHelper) WaitOK(challengeCompleted func() bool) {
	time.Sleep(3 * time.Second)
	for !o.hasOpPrivilege {
		o.GetGameControl().SendWSCmdAndInvokeOnResponse("tp @s ~~~", func(output *packet.CommandOutput) {
			if output.SuccessCount > 0 {
				o.hasOpPrivilege = true
			}
		})
		o.GetGameControl().BotSay("请给予机器人管理员权限")
		fmt.Println("请给予机器人管理权限")
		time.Sleep(1 * time.Second)
	}
	fmt.Println("机器人已获得管理权限")
	first := true
	fmt.Println("等待FB服务器答复网易租赁服 challenge ...")
	success := challengeCompleted()
	if !success {
		panic(fmt.Errorf("FB服务器未能在一定时间内完成网易租赁服零知识身份证明"))
	} else {
		fmt.Println("已完成网易租赁服关于机器人的零知识身份证明")
	}
	for !o.cheatOn {
		o.GetGameControl().SendWSCmdAndInvokeOnResponse("testforblock ~~~ air 0", func(output *packet.CommandOutput) {
			if output.SuccessCount > 0 {
				o.cheatOn = true
			}
			if len(output.OutputMessages) > 0 {
				if strings.Contains(output.OutputMessages[0].Message, "commands.generic.disabled") {
					o.cheatOn = false
					if first {
						fmt.Println("请打开作弊模式")
						first = false
					}
				} else {
					fmt.Println("作弊模式已经打开")
					o.cheatOn = true
				}
			}
		})
		o.GetGameControl().SendWSCmdAndInvokeOnResponse("tp @s ~~~", func(output *packet.CommandOutput) {
			if output.SuccessCount > 0 {
				o.cheatOn = true
			}
		})
		time.Sleep(3 * time.Second)
		if !o.cheatOn {
			o.GetGameControl().BotSay("请打开作弊模式")
			fmt.Println("请打开作弊模式")
		}
	}

}
