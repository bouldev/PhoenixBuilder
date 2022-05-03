package components

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"time"
)

type IntrusionDetectSystem struct {
	*BasicComponent
	EnableK32Detect bool          `json:"启用32k手持物品检测"`
	K32Threshold    int           `json:"32k手持物品附魔等级阈值"`
	k32Response     []defines.Cmd `json:"32k手持物品反制"`
	Patrol          int           `json:"随机巡逻(秒)"`
	EnablePatrol    bool          `json:"启用随机巡逻"`
}

func findK(key string, val interface{}, onKey func(interface{})) {
	switch value := val.(type) {
	case map[string]interface{}:
		for k, v := range value {
			if k == key {
				onKey(v)
			} else {
				findK(key, v, onKey)
			}
		}
	case []interface{}:
		for _, v := range value {
			findK(key, v, onKey)
		}
	case int32:
	}
}

func (o *IntrusionDetectSystem) Init(cfg *defines.ComponentConfig) {
	m, _ := json.Marshal(cfg.Configs)
	err := json.Unmarshal(m, o)
	if err != nil {
		panic(err)
	}
	o.k32Response, err = utils.ParseAdaptiveJsonCmd(cfg.Configs, []string{"32k手持物品反制"})
	if err != nil {
		panic(err)
	}
}

func (o *IntrusionDetectSystem) Inject(frame defines.MainFrame) {
	o.Frame = frame
	o.Frame.GetGameListener().SetOnTypedPacketCallBack(packet.IDAddPlayer, func(p packet.Packet) {
		o.onSeePlayer(p.(*packet.AddPlayer))
	})
	o.Frame.GetGameListener().SetOnTypedPacketCallBack(packet.IDMobEquipment, func(p packet.Packet) {
		o.onSeeMobItem(p.(*packet.MobEquipment))
	})
}

func (o *IntrusionDetectSystem) onSeeMobItem(pk *packet.MobEquipment) {
	if pk.EntityRuntimeID < 2 {
		// do not check bot
		return
	}
	if o.EnableK32Detect {
		nbt := pk.NewItem.Stack.NBTData
		has32K := false
		findK("lvl", nbt, func(v interface{}) {
			level := int(v.(int16))
			if level > o.K32Threshold {
				has32K = true
			}
		})
		if has32K {
			playerName := "未知玩家"
			for _, p := range o.Frame.GetUQHolder().PlayersByEntityID {
				if p.Entity != nil && p.Entity.RuntimeID == pk.EntityRuntimeID {
					playerName = p.Username
				}
			}
			o.Frame.GetBackendDisplay().Write(fmt.Sprintf("发现 32k 玩家 %v", playerName))
			marshal, _ := json.Marshal(pk)
			o.Frame.GetBackendDisplay().Write(string(marshal))
			utils.LaunchCmdsArray(o.Frame.GetGameControl(), o.k32Response, map[string]interface{}{
				"[player]": playerName,
			}, o.Frame.GetBackendDisplay())
		}
	}
}

func (o *IntrusionDetectSystem) onSeePlayer(pk *packet.AddPlayer) {
	//name := pk.Username
	if o.EnableK32Detect {
		nbt := pk.HeldItem.Stack.NBTData
		has32K := false
		findK("lvl", nbt, func(v interface{}) {
			level := int(v.(int16))
			if level > o.K32Threshold {
				has32K = true
			}
		})
		if has32K {
			o.Frame.GetBackendDisplay().Write(fmt.Sprintf("发现 32k 玩家 %v", pk.Username))
			marshal, _ := json.Marshal(pk)
			o.Frame.GetBackendDisplay().Write(string(marshal))
			utils.LaunchCmdsArray(o.Frame.GetGameControl(), o.k32Response, map[string]interface{}{
				"[player]": pk.Username,
			}, o.Frame.GetBackendDisplay())
		}
	}
}

func (o *IntrusionDetectSystem) Activate() {
	if o.EnablePatrol && o.Patrol > 0 {
		go func() {
			for {
				t := time.NewTimer(time.Second * time.Duration(o.Patrol))
				<-t.C
				//fmt.Println("巡逻")
				o.Frame.GetGameControl().SendCmd("effect @s invisibility 60 1 true")
				o.Frame.GetGameControl().SendCmd("tp @s @r ")
				o.Frame.GetGameControl().SendCmd("tp @s ~ 255 ~")
			}
		}()
	}
}
