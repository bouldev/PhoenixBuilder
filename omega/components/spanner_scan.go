package components

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/mirror"
	"phoenixbuilder/mirror/define"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"strings"
)

type MobSpawnerScan struct {
	*defines.BasicComponent
	FilterHas      []string      `json:"如果包含以下关键词则清除"`
	FilterHasnt    []string      `json:"如果不包含以下关键词之一则清除"`
	cleanUpActions []defines.Cmd `json:"违规刷怪笼反制"`
}

func (o *MobSpawnerScan) needRemove(l string) bool {
	if o.FilterHas != nil && len(o.FilterHas) != 0 {
		for _, h := range o.FilterHas {
			if strings.Contains(l, h) {
				return true
			}
		}
	}
	if o.FilterHasnt != nil && len(o.FilterHasnt) != 0 {
		for _, h := range o.FilterHasnt {
			if strings.Contains(l, h) {
				return false
			}
		}
		return true
	}
	return false
}

func (o *MobSpawnerScan) Init(cfg *defines.ComponentConfig) {
	m, _ := json.Marshal(cfg.Configs)
	err := json.Unmarshal(m, o)
	if err != nil {
		panic(err)
	}
	o.cleanUpActions, err = utils.ParseAdaptiveJsonCmd(cfg.Configs, []string{"违规刷怪笼反制"})
	if err != nil {
		panic(err)
	}
}

func (o *MobSpawnerScan) checkNbt(x, y, z int, nbt map[string]interface{}) {
	illegal := false
	reason := ""
	findK("EntityIdentifier", nbt, func(v interface{}) {
		if EntityIdentifier, success := v.(string); success {
			if o.needRemove(EntityIdentifier) {
				reason = fmt.Sprintf("非法的刷怪物: %v", EntityIdentifier)
				illegal = true
			}
		}
	})
	findK("MinSpawnDelay", nbt, func(v interface{}) {
		if deley, success := v.(int16); success {
			if deley < 200 {
				reason = fmt.Sprintf("刷怪速度过快: %v", deley)
				illegal = true
			}
		}
	})
	findK("MaxSpawnDelay", nbt, func(v interface{}) {
		if deley, success := v.(int16); success {
			if deley < 200 {
				reason = fmt.Sprintf("刷怪速度过快: %v", deley)
				illegal = true
			}
		}
	})
	//fmt.Println(nbt)
	if illegal {
		o.Frame.GetBackendDisplay().Write(fmt.Sprintf("位于 %v %v %v 的违规刷怪笼: %v", x, y, z, reason))
		go utils.LaunchCmdsArray(o.Frame.GetGameControl(), o.cleanUpActions, map[string]interface{}{
			"[x]": x,
			"[y]": y,
			"[z]": z,
		}, o.Frame.GetBackendDisplay())
	}
}

func (o *MobSpawnerScan) onLevelChunk(cd *mirror.ChunkData) {
	for _, nbt := range cd.BlockNbts {
		if x, y, z, success := define.GetPosFromNBT(nbt); success {
			o.checkNbt(int(x), int(y), int(z), nbt)
		}
	}
}

func (o *MobSpawnerScan) Inject(frame defines.MainFrame) {
	o.Frame = frame
	o.Frame.GetGameListener().SetOnLevelChunkCallBack(o.onLevelChunk)
}
