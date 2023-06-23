package liliya

import (
	"encoding/json"
	"fmt"
	"math"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	"time"
)

type PickBlock struct {
	*defines.BasicComponent
	Triggers       []string `json:"菜单触发词"`
	Usage          string   `json:"菜单项描述"`
	NeedPermission bool     `json:"OP权限验证"`
}

func (o *PickBlock) Init(cfg *defines.ComponentConfig, storage defines.StorageAndLogProvider) {
	marshal, _ := json.Marshal(cfg.Configs)
	if err := json.Unmarshal(marshal, o); err != nil {
		panic(err)
	}
}

func (o *PickBlock) Inject(frame defines.MainFrame) {
	o.Frame = frame
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
	o.Frame.GetGameControl().SendCmd("clear")
	o.Frame.GetGameControl().SendMCPacket(&packet.BlockPickRequest{
		Position:    protocol.BlockPos{x, y, z},
		AddBlockNBT: true,
	})
}

func (o *PickBlock) throwItem() bool {
	// 切换手持物品栏为 0 并且刷新背包数据
	o.Frame.GetGameControl().SendMCPacket(&packet.PlayerHotBar{
		SelectedHotBarSlot: 0,
		WindowID:           0,
		SelectHotBarSlot:   true,
	})
	o.Frame.GetGameControl().SendCmd("replaceitem entity @s slot.inventory 0 apple 1")
	time.Sleep(time.Second)
	// 尝试丢出快捷栏第一位的物品
	uq := o.Frame.GetUQHolder()
	fmt.Println(uq.InventoryContent[0])
	if len(uq.InventoryContent[0]) > 0 {
		if ii := uq.InventoryContent[0][0]; ii.Stack.Count > 0 {
			o.Frame.GetGameControl().SendMCPacket(&packet.ItemStackRequest{
				Requests: []protocol.ItemStackRequest{
					{
						RequestID: int32(-1),
						Actions: []protocol.StackRequestAction{
							&protocol.DropStackRequestAction{
								Count: byte(ii.Stack.Count),
								Source: protocol.StackRequestSlotInfo{
									ContainerID:    28,
									Slot:           0,
									StackNetworkID: ii.StackNetworkID,
								},
								Randomly: false,
							},
						},
					},
				},
			})
			return true
		}
	}
	return false
}

func (o *PickBlock) onInvoke(chat *defines.GameChat) bool {
	// 权限验证
	if o.NeedPermission && !o.isOP(chat.Name) {
		o.Frame.GetGameControl().SayTo(chat.Name, "§c需要OP权限")
		return true
	}
	go func() {
		// 前往玩家位置
		o.Frame.GetGameControl().SendCmd(fmt.Sprintf("tp @s @a[name=\"%s\"]", chat.Name))
		// 获取脚下坐标
		time.Sleep(time.Second)
		pos := o.Frame.GetUQHolder().BotPos.Position
		x, y, z := int32(math.Floor(float64(pos.X()))), int32(math.Floor(float64(pos.Y())))-2, int32(math.Floor(float64(pos.Z())))
		// 尝试Pick方块
		o.blockPick(x, y, z)
		// 面向玩家并尝试丢出方块
		o.Frame.GetGameControl().SendCmd(fmt.Sprintf("tp ~~~ facing @a[name=\"%s\"]", chat.Name))
		if o.throwItem() {
			o.Frame.GetGameControl().SayTo(chat.Name, fmt.Sprintf("§a已成功Pick位于§7(%d, %d, %d)§a的方块并丢出", x, y, z))
		} else {
			o.Frame.GetGameControl().SayTo(chat.Name, fmt.Sprintf("§c无法Pick位于§7(%d, %d, %d)§c的方块", x, y, z))
		}
	}()
	return true
}
