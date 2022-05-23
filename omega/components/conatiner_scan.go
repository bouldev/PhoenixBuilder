package components

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/mirror"
	"phoenixbuilder/mirror/define"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
)

type ContainerScan struct {
	*BasicComponent
	EnableK32Detect bool          `json:"启用32容器检测"`
	K32Threshold    int           `json:"32k物品附魔等级阈值"`
	k32Response     []defines.Cmd `json:"32k容器反制"`
}

func (o *ContainerScan) Init(cfg *defines.ComponentConfig) {
	m, _ := json.Marshal(cfg.Configs)
	err := json.Unmarshal(m, o)
	if err != nil {
		panic(err)
	}
	o.k32Response, err = utils.ParseAdaptiveJsonCmd(cfg.Configs, []string{"32k容器反制"})
	if err != nil {
		panic(err)
	}
}

func (o *ContainerScan) checkNbt(x, y, z int, nbt map[string]interface{}, getStr func() string) {
	has32K := false
	findK("lvl", nbt, func(v interface{}) {
		level := int(v.(int16))
		if level > o.K32Threshold {
			has32K = true
		}
	})
	if has32K {
		o.Frame.GetBackendDisplay().Write(fmt.Sprintf("位于 %v %v %v 的32k方块:"+getStr(), x, y, z))
		utils.LaunchCmdsArray(o.Frame.GetGameControl(), o.k32Response, map[string]interface{}{
			"[x]": x,
			"[y]": y,
			"[z]": z,
		}, o.Frame.GetBackendDisplay())
	}
}

func (o *ContainerScan) onLevelChunk(cd *mirror.ChunkData) {
	if o.EnableK32Detect {
		for _, nbt := range cd.BlockNbts {
			x, y, z := define.GetPosFromNBT(nbt)
			o.checkNbt(int(x), int(y), int(z), nbt, func() string {
				marshal, _ := json.Marshal(nbt)
				return string(marshal)
			})
		}
	}
}

func (o *ContainerScan) onBlockActorData(pk *packet.BlockActorData) {
	if o.EnableK32Detect {
		nbt := pk.NBTData
		x, y, z := pk.Position.X(), pk.Position.Y(), pk.Position.Z()
		o.checkNbt(int(x), int(y), int(z), nbt, func() string {
			marshal, _ := json.Marshal(nbt)
			return string(marshal)
		})
	}
}

func (o *ContainerScan) Inject(frame defines.MainFrame) {
	o.Frame = frame
	o.Frame.GetGameListener().SetOnTypedPacketCallBack(packet.IDBlockActorData, func(p packet.Packet) {
		o.onBlockActorData(p.(*packet.BlockActorData))
	})
	o.Frame.GetGameListener().SetOnLevelChunkCallBack(o.onLevelChunk)
}
