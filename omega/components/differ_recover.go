package components

import (
	"encoding/json"
	"fmt"
	"path"
	"phoenixbuilder/mirror"
	"phoenixbuilder/mirror/chunk"
	"phoenixbuilder/mirror/define"
	"phoenixbuilder/mirror/io/mcdb"
	"phoenixbuilder/omega/defines"
	"strconv"
	"strings"
	"time"

	"github.com/df-mc/goleveldb/leveldb/opt"
)

type DifferRecover struct {
	*BasicComponent
	Triggers                      []string `json:"触发词"`
	Speed                         int      `json:"修复速度"`
	BackUpName                    string   `json:"备份存档名"`
	Operators                     []string `json:"授权使用者"`
	currentProvider, ckptProvider mirror.ChunkReader
}

func (o *DifferRecover) Init(cfg *defines.ComponentConfig) {
	m, _ := json.Marshal(cfg.Configs)
	err := json.Unmarshal(m, o)
	if err != nil {
		panic(err)
	}
	o.BackUpName = "backup"
}

type blockToRepair struct {
	pos   define.CubePos
	block uint32
}

func (o *DifferRecover) GetBlocksPipe(currentProvider, ckptProvider mirror.ChunkReader, pos define.CubePos, distance int) chan blockToRepair {
	computeRequiredChunks := func(pos define.CubePos, distance int) (requiredChunks []define.ChunkPos) {
		// chunkX, ChunkZ := int(math.Floor(float64(pos[0]/16))), int(math.Floor(float64(pos[2]/16)))
		chunkSX, chunkSZ := (pos[0]-distance)/16, (pos[2]-distance)/16
		chunkEX, chunkEZ := (pos[0]+distance)/16, (pos[2]+distance)/16
		requiredChunks = make([]define.ChunkPos, 0)
		for cx := chunkSX; cx <= chunkEX; cx++ {
			for cz := chunkSZ; cz <= chunkEZ; cz++ {
				requiredChunks = append(requiredChunks, define.ChunkPos{int32(cx), int32(cz)})
			}
		}
		return requiredChunks
	}
	chunksToFix := computeRequiredChunks(pos, distance)
	repairChan := make(chan blockToRepair, 10240)
	go func() {
		for _, chunkPos := range chunksToFix {
			ckpt := ckptProvider.Get(chunkPos)
			current := currentProvider.Get(chunkPos)
			if ckpt == nil {
				fmt.Printf("missing backup chunk @ %v, skipping\n", chunkPos)
				continue
			}
			if current == nil {
				fmt.Printf("missing current chunk @ %v, skipping\n", chunkPos)
				continue
			}
			for x := uint8(0); x < 16; x++ {
				for z := uint8(0); z < 16; z++ {
					for y := int16(256); y != 0; y-- {
						targetBlock := ckpt.Chunk.Block(x, y, z, 0)
						realBlock := current.Chunk.Block(x, y, z, 0)
						if targetBlock != realBlock {
							p := define.CubePos{int(x) + int(chunkPos[0])*16, int(y), int(z) + int(chunkPos[1])*16}
							repairChan <- blockToRepair{pos: p, block: targetBlock}
						}
					}
				}
			}
		}
		close(repairChan)
	}()
	return repairChan
}

func (o *DifferRecover) onTrigger(chat *defines.GameChat) (stop bool) {
	flag := false
	for _, name := range o.Operators {
		if name == chat.Name {
			flag = true
		}
	}
	if !flag {
		o.Frame.GetGameControl().SayTo(chat.Name, "你没有权限使用这个功能")
		return true
	} else {
		o.Frame.GetGameControl().SayTo(chat.Name, fmt.Sprintf("修复速度为 %v, 预计每秒修复 %.1f 个方块", o.Speed, float32(1000)/float32(o.Speed)))
		o.Frame.GetGameControl().SayTo(chat.Name, fmt.Sprintf("注意！ 过快的速度可能导致机器人崩溃或租赁服崩溃"))
	}
	distance := 1
	if len(chat.Msg) > 0 {
		_d, err := strconv.Atoi(chat.Msg[0])
		if err == nil {
			distance = _d
		} else {
			o.Frame.GetGameControl().SayTo(chat.Name, "输入的修复半径无效，将修复当前所在区块")
		}
	} else {
		o.Frame.GetGameControl().SayTo(chat.Name, "未输入修复半径，将修复当前所在区块")
	}
	go func() {
		pk := o.Frame.GetGameControl().GetPlayerKit(chat.Name)
		_currentPos := <-pk.GetPos("@a[name=[player]]")
		if _currentPos == nil || len(_currentPos) == 0 {
			pk.Say("位置无效，请退出租赁服重试")
		}
		fmt.Println(_currentPos)
		o.Frame.GetGameControl().SendCmd(fmt.Sprintf("tp @s %v %v %v", _currentPos[0], _currentPos[1], _currentPos[2]))
		currentPos := define.CubePos{_currentPos[0], _currentPos[1], _currentPos[2]}
		blocksToFix := o.GetBlocksPipe(o.currentProvider, o.ckptProvider, currentPos, distance)
		counter := 0
		t := time.NewTicker(time.Millisecond * time.Duration(o.Speed))
		sender := o.Frame.GetGameControl().SendWOCmd
		for block := range blocksToFix {
			counter++
			if counter%100 == 99 {
				pk.ActionBar(fmt.Sprintf("current %v blocks\n", counter+1))
			}
			blk := chunk.RuntimeIDToLegacyBlock(block.block)
			if blk == nil {
				continue
			}
			cmd := fmt.Sprintf("setblock %v %v %v %v %v\n", block.pos[0], block.pos[1], block.pos[2], strings.ReplaceAll(blk.Name, "minecraft:", ""), blk.Val)
			// fmt.Println(cmd)
			sender(cmd)
			<-t.C
		}
		pk.Say(fmt.Sprintf("完成，总计 %v 方块\n", counter))
	}()
	return true
}

// func (o *DifferRecover) onChunk(chunk *mirror.ChunkData) {
// 	ckpt := o.ckptProvider.Get(chunk.ChunkPos)
// 	ckptNbts := chunk.BlockNbts
// 	if ckpt == nil {
// 		ckptNbts = make(map[define.CubePos]map[string]interface{})
// 	} else {
// 		ckptNbts = ckpt.BlockNbts
// 	}
// 	realNbts := chunk.BlockNbts
// 	allNbts := make(map[define.CubePos]bool)
// 	for k := range ckptNbts {
// 		allNbts[k] = true
// 	}
// 	for k := range realNbts {
// 		allNbts[k] = true
// 	}
// 	for k := range allNbts {
// 		ckptNBT := ckptNbts[k]
// 		realNBT := realNbts[k]
// 		if (ckptNBT == nil || ckptNBT["id"] != "CommandBlock") && (realNBT == nil || realNBT["id"] != "CommandBlock") {
// 			continue
// 		}
// 		cS := "no_data"
// 		rS := "no_data"
// 		if ckptNBT != nil {
// 			cS = ckptNBT["Command"].(string)
// 		}
// 		if realNBT != nil {
// 			rS = realNBT["Command"].(string)
// 			if strings.Contains(rS, "lightning") {
// 				pterm.Error.Printf("%v: ckpt %v != current %v\n", k, ckptNBT, rS)
// 			}
// 		}

// 		if cS != rS {
// 			fmt.Printf("%v: ckpt %v != current %v\n", k, ckptNBT, rS)
// 		}
// 	}
// }

func (o *DifferRecover) Inject(frame defines.MainFrame) {
	o.Frame = frame
	ckptPath := path.Join(o.Frame.GetWorldsDir(), o.BackUpName)
	ckptProvider, err := mcdb.New(ckptPath, opt.FlateCompression)
	if err != nil {
		panic(err)
	}
	if ckptProvider == nil {
		panic(fmt.Errorf("找不到指定的存档文件，配置文件中的文件夹应该位于 %v", ckptPath))
	}
	currentProvider := o.Frame.GetWorldProvider()
	o.ckptProvider = ckptProvider
	o.currentProvider = currentProvider
	o.Frame.GetGameListener().SetGameMenuEntry(&defines.GameMenuEntry{
		MenuEntry: defines.MenuEntry{
			Triggers:     o.Triggers,
			ArgumentHint: "[修复半径]",
			FinalTrigger: false,
			Usage:        "根据备份档中信息修复所在区域",
		},
		OptionalOnTriggerFn: o.onTrigger,
	})
	// o.Frame.GetGameListener().SetOnLevelChunkCallBack(o.onChunk)
}
