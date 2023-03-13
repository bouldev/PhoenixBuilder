package liliya

import (
	"encoding/json"
	"fmt"
	"math"
	blockNBT_API "phoenixbuilder/fastbuilder/bdump/blockNBT/API"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
)

type PickBlock struct {
	*defines.BasicComponent
	Triggers       []string `json:"菜单触发词"`
	Usage          string   `json:"菜单项描述"`
	NeedPermission bool     `json:"OP权限验证"`
	apis           blockNBT_API.GlobalAPI
}

func (o *PickBlock) Init(cfg *defines.ComponentConfig, storage defines.StorageAndLogProvider) {
	marshal, _ := json.Marshal(cfg.Configs)
	if err := json.Unmarshal(marshal, o); err != nil {
		panic(err)
	}
}

func (o *PickBlock) Inject(frame defines.MainFrame) {
	o.Frame = frame
	o.apis = blockNBT_API.GlobalAPI{
		WritePacket: func(p packet.Packet) error {
			o.Frame.GetGameControl().SendMCPacket(p)
			return nil
		},
		BotName:            o.Frame.GetUQHolder().GetBotName(),
		BotIdentity:        "",
		BotUniqueID:        o.Frame.GetUQHolder().BotUniqueID,
		BotRunTimeID:       o.Frame.GetUQHolder().BotRuntimeID,
		PacketHandleResult: o.Frame.GetNewUQHolder(),
	}
	o.Frame.GetGameListener().SetGameMenuEntry(&defines.GameMenuEntry{
		MenuEntry: defines.MenuEntry{
			Triggers:     o.Triggers,
			FinalTrigger: false,
			Usage:        o.Usage,
		},
		OptionalOnTriggerFn: o.onInvoke,
	})
}

func (o *PickBlock) isOP(name string) bool {
	return o.Frame.GetGameControl().GetPlayerKit(name).GetRelatedUQ().OPPermissionLevel > 1
}

func (o *PickBlock) blockPick(x, y, z int32) {
	_, err := o.apis.SendWSCommandWithResponce("clear")
	if err != nil {
		panic(fmt.Sprintf("blockPick: %v", err))
	}
	o.apis.WritePacket(&packet.BlockPickRequest{
		Position:    protocol.BlockPos{x, y, z},
		AddBlockNBT: true,
	})
}

func (o *PickBlock) throwItem() bool {
	_, err := o.apis.SendWSCommandWithResponce("list")
	if err != nil {
		panic(fmt.Sprintf("throwItem: %v", err))
	}
	// 刷新背包数据(等待更改)
	datas, err := o.apis.PacketHandleResult.Inventory.GetItemStackInfo(0, 0)
	if err != nil {
		return false
	}
	// 取得快捷栏 0 的物品数据
	if datas.Stack.Count > 0 {
		ans, err := o.apis.SendItemStackRequestWithResponce(&packet.ItemStackRequest{
			Requests: []protocol.ItemStackRequest{
				{
					Actions: []protocol.StackRequestAction{
						&protocol.DropStackRequestAction{
							Count: byte(datas.Stack.Count),
							Source: protocol.StackRequestSlotInfo{
								ContainerID:    28,
								Slot:           0,
								StackNetworkID: datas.StackNetworkID,
							},
							Randomly: false,
						},
					},
				},
			},
		})
		if err != nil {
			return false
		}
		// 发送数据包
		if ans[0].Status == 0 {
			return true
		}
		// 返回值
	}
	// 尝试丢出物品
	return false
	// 返回值
}

func (o *PickBlock) onInvoke(chat *defines.GameChat) bool {
	// 权限验证
	if o.NeedPermission && !o.isOP(chat.Name) {
		o.Frame.GetGameControl().SayTo(chat.Name, "§c需要OP权限")
		return true
	}
	go func() {
		// 前往玩家位置
		o.apis.BotName = o.Frame.GetUQHolder().GetBotName()
		err := o.apis.SendSettingsCommand(fmt.Sprintf("tp @s @a[name=\"%s\"]", chat.Name), true)
		if err != nil {
			panic(fmt.Sprintf("onInvoke: %v", err))
		}
		// 获取脚下坐标
		resp, err := o.apis.SendWSCommandWithResponce("querytarget @s")
		respString := resp.OutputMessages[0].Parameters[0]
		var respList []interface{}
		json.Unmarshal([]byte(respString), &respList)
		if len(respList) <= 0 {
			return
		}
		respMap := respList[0].(map[string]interface{})
		x, y, z := int32(math.Floor(respMap["position"].(map[string]interface{})["x"].(float64))), int32(math.Floor(respMap["position"].(map[string]interface{})["y"].(float64)))-2, int32(math.Floor(float64(respMap["position"].(map[string]interface{})["z"].(float64))))
		// 尝试Pick方块
		o.blockPick(x, y, z)
		// 面向玩家并尝试丢出方块
		o.apis.SendSettingsCommand(fmt.Sprintf("tp ~ ~ ~ facing @a[name=\"%s\"]", chat.Name), true)
		if o.throwItem() {
			o.Frame.GetGameControl().SayTo(chat.Name, fmt.Sprintf("§a已成功 §fPick §a位于 §7(§b%d§f, §b%d§f, §b%d§7) §a的方块并丢出", x, y, z))
		} else {
			o.Frame.GetGameControl().SayTo(chat.Name, fmt.Sprintf("§c无法 §fPick §c位于 §7(§b%d§f, §b%d§f, §b%d§7) §c的方块", x, y, z))
		}
	}()
	return true
}
