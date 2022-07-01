package components

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/mirror"
	"phoenixbuilder/mirror/chunk"
	"phoenixbuilder/mirror/define"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"regexp"
)

type RemoveBlockComplex struct {
	targetRTID []uint32
	DenyCmdsIn interface{} `json:"处理指令"`
	DenyCmds   []defines.Cmd
}

type RemoveBlock struct {
	*BasicComponent
	BlocksToRemove map[string]*RemoveBlockComplex `json:"清除规则"`
	fastFilter     map[uint32]*RemoveBlockComplex
}

func (o *RemoveBlock) Init(cfg *defines.ComponentConfig) {
	m, _ := json.Marshal(cfg.Configs)
	err := json.Unmarshal(m, o)
	if err != nil {
		panic(err)
	}
	o.fastFilter = make(map[uint32]*RemoveBlockComplex)
	for RegString, complex := range o.BlocksToRemove {
		complex.targetRTID = make([]uint32, 0)
		reg := regexp.MustCompile(RegString)
		for rtid, block := range chunk.Blocks {
			if reg.Find([]byte(block.Name)) != nil {
				fmt.Printf("%v => %v\n", RegString, rtid)
				complex.targetRTID = append(complex.targetRTID, uint32(rtid))
				o.fastFilter[uint32(rtid)] = complex
			}
		}
		if complex.DenyCmds, err = utils.ParseAdaptiveCmd(complex.DenyCmdsIn); err != nil {
			panic(err)
		}
	}
}

//TODO Check if remove block is affected by 0 -> -64
func (o *RemoveBlock) onLevelChunk(cd *mirror.ChunkData) {
	for sub_i, sub := range cd.Chunk.Sub() {
		palette := sub.Layer(0).Palette()
		flag := false
		for palette_i := 0; palette_i < palette.Len(); palette_i++ {
			rtid := palette.Value(uint16(palette_i))
			if _, hasK := o.fastFilter[rtid]; hasK {
				// fmt.Println("HasK!")
				flag = true
			}
		}
		if !flag {
			continue
		}
		for x := uint8(0); x < 16; x++ {
			for z := uint8(0); z < 16; z++ {
				for y := uint8(0); y < 24; y++ {
					rtid := sub.Block(x, y, z, 0)
					if complex, hasK := o.fastFilter[rtid]; hasK {
						wy := int16(sub_i)*16 + int16(y) + int16(define.WorldRange[0])
						wx := int(cd.ChunkPos.X())*16 + int(x)
						wz := int(cd.ChunkPos.Z())*16 + int(z)
						// fmt.Println(wx, wy, wz)
						go utils.LaunchCmdsArray(o.Frame.GetGameControl(), complex.DenyCmds, map[string]interface{}{
							"[x]": wx, "[y]": wy, "[z]": wz,
						}, o.Frame.GetBackendDisplay())
					}
				}
			}
		}
	}
}

func (o *RemoveBlock) onBlockUpdate(pos define.CubePos, origRTID, currentRTID uint32) {
	if complex, hasK := o.fastFilter[currentRTID]; hasK {
		go utils.LaunchCmdsArray(o.Frame.GetGameControl(), complex.DenyCmds, map[string]interface{}{
			"[x]": pos.X(), "[y]": pos.Y(), "[z]": pos.Z(),
		}, o.Frame.GetBackendDisplay())
	}
}

func (o *RemoveBlock) Inject(frame defines.MainFrame) {
	o.Frame = frame
	o.Frame.GetGameListener().AppendOnBlockUpdateInfoCallBack(o.onBlockUpdate)
	o.Frame.GetGameListener().SetOnLevelChunkCallBack(o.onLevelChunk)
}
