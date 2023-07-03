package bot_privilege

import (
	"context"
	"errors"
	"fmt"
	"phoenixbuilder/lib/minecraft/neomega/omega"
	"phoenixbuilder/minecraft/protocol/packet"
	"strings"
	"time"
)

type SetupHelper struct {
	omega.MicroOmega
	hasOpPrivilege  bool
	cheatOn         bool
	lostPrivilegeCB func()
}

func NewSetupHelper(omega omega.MicroOmega, lostPrivilegeCallBack func()) *SetupHelper {
	if lostPrivilegeCallBack == nil {
		lostPrivilegeCallBack = func() {
			panic(fmt.Errorf("机器人OP权限被关闭"))
		}
	}
	helper := &SetupHelper{
		MicroOmega:      omega,
		lostPrivilegeCB: lostPrivilegeCallBack,
	}
	omega.GetGameListener().SetOnTypedPacketCallBack(packet.IDAdventureSettings, helper.onAdventurePacket)
	return helper
}

func (o *SetupHelper) onAdventurePacket(pk packet.Packet) {
	p := pk.(*packet.AdventureSettings)
	if o.GetMicroUQHolder().GetBotBasicInfo().GetBotUniqueID() == p.PlayerUniqueID {
		if p.PermissionLevel >= packet.PermissionLevelOperator {
			o.hasOpPrivilege = true
			fmt.Println("机器人已获得管理员权限")
		} else {
			if o.hasOpPrivilege {
				o.lostPrivilegeCB()
			}
			fmt.Println("请给予机器人管理员权限")
			o.hasOpPrivilege = false
		}
	}
}

var ErrMaximumWaitTimeExceed = errors.New("未能在指定时间内获得机器人所需权限")
var ErrFBServerCannotResponseZeroKnowledgeProof = errors.New("FB服务器未能在一定时间内完成网易租赁服零知识身份证明")

func (o *SetupHelper) WaitOK(ctx context.Context, challengeCompleted func(ctx context.Context) bool) (err error) {
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
		if ctx.Err() != nil {
			return ErrMaximumWaitTimeExceed
		}
	}
	fmt.Println("机器人已获得管理权限")
	first := true
	fmt.Println("等待FB服务器答复网易租赁服 challenge ...")
	success := challengeCompleted(ctx)
	if !success {
		return ErrFBServerCannotResponseZeroKnowledgeProof
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
		time.Sleep(1 * time.Second)
		if ctx.Err() != nil {
			return ErrMaximumWaitTimeExceed
		}
		if !o.cheatOn {
			o.GetGameControl().BotSay("请打开作弊模式")
			fmt.Println("请打开作弊模式")
		}
	}
	return nil
}
