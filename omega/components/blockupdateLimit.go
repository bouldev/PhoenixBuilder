package components

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/mirror/chunk"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"strings"
	"sync"
	"time"
)

type RedStoneUpdateLimit struct {
	*defines.BasicComponent
	MaxUpdatePer10Second int           `json:"10s内最多允许的变化次数"`
	execeedResponse      []defines.Cmd `json:"刷新过快的反制"`
	BlockNames           []string      `json:"方块名里包含这些关键词时即检查"`
	redstoneRidCache     map[uint32]bool
	mu                   sync.Mutex
	updateRecord         map[protocol.BlockPos]int
}

func (o *RedStoneUpdateLimit) Init(cfg *defines.ComponentConfig) {
	m, _ := json.Marshal(cfg.Configs)
	err := json.Unmarshal(m, o)
	if err != nil {
		panic(err)
	}
	o.execeedResponse, err = utils.ParseAdaptiveJsonCmd(cfg.Configs, []string{"刷新过快的反制"})
	if err != nil {
		panic(err)
	}
	o.redstoneRidCache = map[uint32]bool{}
	o.updateRecord = make(map[protocol.BlockPos]int)
}

func (o *RedStoneUpdateLimit) doResponse(pos protocol.BlockPos) {
	x, y, z := pos.X(), pos.Y(), pos.Z()
	o.Frame.GetBackendDisplay().Write(fmt.Sprintf("位于 %v %v %v 的红石相关方块刷新过快", x, y, z))
	go utils.LaunchCmdsArray(o.Frame.GetGameControl(), o.execeedResponse, map[string]interface{}{
		"[x]": x,
		"[y]": y,
		"[z]": z,
	}, o.Frame.GetBackendDisplay())
}

func (o *RedStoneUpdateLimit) recordUpdate(pos protocol.BlockPos) {
	defer func() {
		r := recover()
		if r != nil {
			fmt.Println(r)
		}
	}()
	if c, hasK := o.updateRecord[pos]; hasK {
		if c > o.MaxUpdatePer10Second {
			o.doResponse(pos)
			o.updateRecord[pos] = 0
		} else {
			o.updateRecord[pos]++
		}
	} else {
		o.updateRecord[pos] = 1
	}
}

func (o *RedStoneUpdateLimit) onBlockUpdate(pk *packet.UpdateBlock) {
	nemcRid := pk.NewBlockRuntimeID
	rid := chunk.NEMCRuntimeIDToStandardRuntimeID(nemcRid)
	legacyBlock := chunk.RuntimeIDToLegacyBlock(rid)
	// fmt.Println(nemcRid, rid, legacyBlock)
	isRedstone := false
	if _isRedstone, hasK := o.redstoneRidCache[rid]; hasK {
		isRedstone = _isRedstone
	} else {
		for _, bk := range o.BlockNames {
			if strings.Contains(legacyBlock.Name, bk) {
				// fmt.Println(legacyBlock.Name, " is redstone")
				isRedstone = true
				o.redstoneRidCache[rid] = isRedstone
				break
			}
		}
	}
	if !isRedstone {
		return
	}
	o.recordUpdate(pk.Position)
}

func (o *RedStoneUpdateLimit) Inject(frame defines.MainFrame) {
	o.Frame = frame
	o.Frame.GetGameListener().SetOnTypedPacketCallBack(packet.IDUpdateBlock, func(p packet.Packet) {
		o.onBlockUpdate(p.(*packet.UpdateBlock))
	})
}

func (o *RedStoneUpdateLimit) Activate() {
	for {
		<-time.NewTimer(10 * time.Second).C
		o.updateRecord = make(map[protocol.BlockPos]int, 0)
	}
}
