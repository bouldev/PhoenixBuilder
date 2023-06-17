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
	"sync"
	"time"

	"phoenixbuilder/omega/utils/structure"

	"github.com/df-mc/goleveldb/leveldb/opt"
)

type DifferRecover struct {
	*defines.BasicComponent
	Triggers                      []string `json:"触发词"`
	Speed                         int      `json:"修复速度"`
	BackUpName                    string   `json:"备份存档名"`
	Operators                     []string `json:"授权使用者"`
	currentProvider, ckptProvider mirror.ChunkReader
	delayBlocks                   map[define.CubePos]*structure.IOBlockForBuilder
	delayBlocksMu                 sync.Mutex
}

func (o *DifferRecover) Init(cfg *defines.ComponentConfig, storage defines.StorageAndLogProvider) {
	m, _ := json.Marshal(cfg.Configs)
	err := json.Unmarshal(m, o)
	if err != nil {
		panic(err)
	}
	o.BackUpName = "backup"
}

// TODO Check if differ recover is affected by 0 -> -64
func (o *DifferRecover) GetBlocksPipe(currentProvider, ckptProvider mirror.ChunkReader, pos define.CubePos, distance int) chan *structure.IOBlockForBuilder {
	computeRequiredChunks := func(pos define.CubePos, distance int) (requiredChunks []define.ChunkPos) {
		// chunkX, ChunkZ := int(math.Floor(float64(pos[0]/16))), int(math.Floor(float64(pos[2]/16)))
		chunkSX, chunkSZ := (pos[0]-distance-1)/16, (pos[2]-distance-1)/16
		chunkEX, chunkEZ := (pos[0]+distance+1)/16, (pos[2]+distance+1)/16
		requiredChunks = make([]define.ChunkPos, 0)
		for cx := chunkSX; cx <= chunkEX; cx++ {
			for cz := chunkSZ; cz <= chunkEZ; cz++ {
				requiredChunks = append(requiredChunks, define.ChunkPos{int32(cx), int32(cz)})
			}
		}
		return requiredChunks
	}
	chunksToFix := computeRequiredChunks(pos, distance)
	repairChan := make(chan *structure.IOBlockForBuilder, 10240)
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
			nbts := ckpt.BlockNbts
			for x := uint8(0); x < 16; x++ {
				for z := uint8(0); z < 16; z++ {
					for subChunk := int16(0); subChunk < 24; subChunk++ {
						for sy := int16(0); sy < 16; sy++ {
							y := subChunk*16 + sy + int16(define.WorldRange[0])
							targetBlock := ckpt.Chunk.Block(x, y, z, 0)
							realBlock := current.Chunk.Block(x, y, z, 0)
							if targetBlock != realBlock {
								p := define.CubePos{int(x) + int(chunkPos[0])*16, int(y), int(z) + int(chunkPos[1])*16}
								b := &structure.IOBlockForBuilder{Pos: p, RTID: targetBlock}
								if nbt, hasK := nbts[p]; hasK {
									b.NBT = nbt
									// fmt.Println("A: ", b.blockNbt)
								}
								repairChan <- b
							}
						}
					}
				}
			}
		}
		close(repairChan)
	}()
	return repairChan
}

func (o *DifferRecover) updateDelayBlocks(force bool) {
	// fmt.Println("C: ", o.delayBlocks)
	for pos, block := range o.delayBlocks {
		fmt.Println(block)
		if block.Hit || force {
			switch block.NBT["id"] {
			case "sign":
				// nbt := block.BlockNbt
				// fmt.Println("sign: ", nbt)
				// o.Frame.GetGameControl().GetInteraction().WritePacket(&packet.BlockActorData{
				// 	Position: protocol.BlockPos{int32(pos[0]), int32(pos[1]), int32(pos[2])},
				// 	NBTData:  nbt,
				// })
			case "CommandBlock":

			}
			o.delayBlocksMu.Lock()
			delete(o.delayBlocks, pos)
			o.delayBlocksMu.Unlock()
		}
	}
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
	o.Frame.GetBotTaskScheduler().CommitNormalTask(&defines.BasicBotTaskPauseAble{
		BasicBotTask: defines.BasicBotTask{
			Name: fmt.Sprintf("Fix"),
			ActivateFn: func() {
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
				o.delayBlocks = make(map[define.CubePos]*structure.IOBlockForBuilder)
				o.delayBlocksMu = sync.Mutex{}
				t := time.NewTicker(time.Millisecond * time.Duration(o.Speed))
				sender := o.Frame.GetGameControl().SendWOCmd
				for block := range blocksToFix {
					counter++
					if counter%100 == 99 {
						pk.ActionBar(fmt.Sprintf("current %v blocks\n", counter+1))
						sender(fmt.Sprintf("tp @s %v %v %v\n", block.Pos[0], block.Pos[1], block.Pos[2]))
					}
					blk, found := chunk.RuntimeIDToLegacyBlock(block.RTID)
					if !found {
						continue
					}
					cmd := fmt.Sprintf("setblock %v %v %v %v %v", block.Pos[0], block.Pos[1], block.Pos[2], strings.ReplaceAll(blk.Name, "minecraft:", ""), blk.Val)
					sender(cmd)
					if block.NBT != nil {
						o.delayBlocksMu.Lock()
						o.delayBlocks[block.Pos] = block
						// fmt.Println("B: ", o.delayBlocks[block.pos])
						o.delayBlocksMu.Unlock()
						if len(o.delayBlocks) > 64 {
							o.updateDelayBlocks(false)
						}
					}
					<-t.C
				}
				time.Sleep(3 * time.Second)
				o.updateDelayBlocks(true)
				pk.Say(fmt.Sprintf("完成，总计 %v 方块\n", counter))
				o.delayBlocks = nil
			},
		},
	})
	return true
}

func (o *DifferRecover) onBlockUpdate(pos define.CubePos, origRTID, currentRTID uint32) {
	if o.delayBlocks != nil {
		b := o.delayBlocks[pos]
		if b == nil {
			return
		}
		if b.RTID == origRTID || b.RTID == currentRTID {
			b.Hit = true
		}
	}
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
	o.Frame.GetGameListener().AppendOnBlockUpdateInfoCallBack(o.onBlockUpdate)
}
