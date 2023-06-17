package structure

import (
	"fmt"
	"phoenixbuilder/mirror"
	"phoenixbuilder/mirror/chunk"
	"phoenixbuilder/mirror/define"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/pterm/pterm"
)

type Builder struct {
	delayBlocks     map[define.CubePos]*IOBlockForBuilder
	delayBlocksMu   sync.Mutex
	BlockCmdSender  func(cmd string)
	NormalCmdSender func(cmd string)
	ProgressUpdater func(currBlock int)
	FinalWaitTime   int
	IgnoreNbt       bool
	Stop            bool
	InitPosGetter   func() define.CubePos
	Ctrl            defines.GameControl
}

type deferAction struct {
	time   time.Time
	action func()
}

func (o *Builder) Build(blocksIn chan *IOBlockForBuilder, speed int, boostSleepTime time.Duration) {
	o.delayBlocks = make(map[define.CubePos]*IOBlockForBuilder)
	// o.delayBlocksMu = sync.Mutex{}
	counter := 0
	var doDelay func()
	if runtime.GOOS == "windows" {
		delay := time.Duration((float64(100*1000) / float64(speed) * float64(time.Millisecond)))
		ticker := time.NewTicker(delay)
		oneHunredCounter := 0
		doDelay = func() {
			if oneHunredCounter == 100 {
				<-ticker.C
				oneHunredCounter = 0
			}
			oneHunredCounter++
		}
	} else {
		delay := time.Duration((float64(1000) / float64(speed) * float64(time.Millisecond)))
		ticker := time.NewTicker(delay)
		doDelay = func() {
			<-ticker.C
		}
	}

	lastPos := o.InitPosGetter()
	pterm.Info.Printfln("DEBUG: Init Pos: %v", lastPos)
	// deferCmd := make(chan deferAction, 10)
	// defer close(deferCmd)
	// go func() {
	// 	for cmd := range deferCmd {
	// 		time.Sleep(time.Until(cmd.time))
	// 		cmd.action()
	// 	}
	// }()
	// deferSendBlock := func(block *IOBlockForBuilder) {
	// 	deferCmd <- deferAction{
	// 		time: time.Now().Add(time.Millisecond * 100),
	// 		action: func() {
	// 			// blk := chunk.RuntimeIDToLegacyBlock(block.RTID)
	// 			o.BlockCmdSender(fmt.Sprintf("testforblock %v %v %v air 0", block.Pos.X(), block.Pos.Y(), block.Pos.Z()))
	// 		},
	// 	}

	// }
	// var lastDelayBlock *IOBlockForBuilder
	fallBackActions := make(map[define.CubePos]func())
	fallBackActionsMu := sync.Mutex{}
	moveTicker := time.NewTicker(time.Millisecond * 50)
	for block := range blocksIn {
		if o.Stop {
			return
		}
		xmove := block.Pos.X() - lastPos.X()
		zmove := block.Pos.Z() - lastPos.Z()
		if counter == 0 {
			o.NormalCmdSender(fmt.Sprintf("tp @s %v %v %v", block.Pos[0], 320, block.Pos[2]))
			lastPos = block.Pos
			time.Sleep(3 * time.Second)
		}
		if (xmove*xmove) > 16*16 || (zmove*zmove) > 16*16 {
			o.NormalCmdSender(fmt.Sprintf("tp @s %v %v %v", block.Pos[0], 320, block.Pos[2]))
			lastPos = block.Pos
			<-moveTicker.C

			// 	if !o.IgnoreNbt && lastDelayBlock != nil {
			// 		counter := 0
			// 		isAir := true
			// 		for {
			// 			time.Sleep(time.Millisecond * 100)
			// 			counter++
			// 			if counter == 30 || !isAir {
			// 				break
			// 			}
			// 			utils.GetBlockAt(o.Ctrl, fmt.Sprintf("%v %v %v", lastDelayBlock.Pos.X(), lastDelayBlock.Pos.Y(), lastDelayBlock.Pos.Z()), func(outOfWorld, _isAir bool, name string, pos define.CubePos) {
			// 				isAir = _isAir
			// 			})
			// 		}
			// 		if !isAir {
			// 			fmt.Println("Wait OK")
			// 		} else {
			// 			fmt.Println("time out")
			// 		}
			// 		for pos, blk := range o.delayBlocks {
			// 			o.SetBlock(pos, blk)
			// 		}
			// 		lastDelayBlock = nil
			// 	}
			// }
			fallBackActionsMu.Lock()
			if !o.IgnoreNbt && len(fallBackActions) > 0 {
				fallBackActionsMu.Unlock()
				time.Sleep(time.Second * 3)
				fallBackActionsMu.Lock()
				forceActions := make(map[define.CubePos]func())
				for pos, actions := range fallBackActions {
					forceActions[pos] = actions
				}
				fallBackActions = make(map[define.CubePos]func())
				fallBackActionsMu.Unlock()
				for pos, action := range fallBackActions {
					pterm.Warning.Printfln("Force Execute Fallback Actions @ %v", pos)
					action()
				}
			} else {
				fallBackActionsMu.Unlock()
			}
		}
		blk, found := chunk.RuntimeIDToLegacyBlock(block.RTID)
		if !found {
			continue
		}
		o.ProgressUpdater(counter)
		if block.Expand16 {
			cmd := fmt.Sprintf("fill %v %v %v %v %v %v %v %v", block.Pos[0], block.Pos[1], block.Pos[2], block.Pos[0]+15, block.Pos[1]+15, block.Pos[2]+15, strings.Replace(blk.Name, "minecraft:", "", 1), blk.Val)
			// fmt.Println("fast fill")
			o.NormalCmdSender(cmd)
			counter += 4096
			time.Sleep(boostSleepTime)
		} else {
			if block.NBT != nil && !o.IgnoreNbt && block.NBT["id"] == "CommandBlock" {
				// cmd := fmt.Sprintf("setblock %v %v %v %v %v", block.Pos[0], block.Pos[1], block.Pos[2], strings.Replace(blk.Name, "minecraft:", "", 1), blk.Val)
				// o.BlockCmdSender(cmd)
				// lastDelayBlock = block
				// deferSendBlock(block)
				placeStart := time.Now()
				quickDone := false
				if cfg, err := utils.GenCommandBlockUpdateFromNbt(block.Pos, blk.Name, block.NBT); err == nil {
					fallBackActionsMu.Lock()
					fallBackActions[block.Pos] = func() {
						pterm.Warning.Printfln("命令方块放置超时 @ %v time out!", block.Pos)
						pterm.Warning.Printfln("重新尝试放置命令方块: 坐标: %v 名称: %v %v 信息: %v", block.Pos, blk.Name, blk.Val, block.NBT)
						o.Ctrl.PlaceCommandBlock(block.Pos, blk.Name, int(blk.Val), false, true, cfg, func(done bool) {
							if !done {
								pterm.Error.Printfln("命令方块放置失败: 坐标: %v 名称: %v %v 信息: %v", block.Pos, blk.Name, blk.Val, block.NBT)
							}
						}, time.Second*3)
					}
					fallBackActionsMu.Unlock()
					o.Ctrl.PlaceCommandBlock(block.Pos, blk.Name, int(blk.Val), false, true, cfg, func(done bool) {
						if done {
							quickDone = true
							fallBackActionsMu.Lock()
							pterm.Success.Printfln("命令方块放置成功 @ %v", block.Pos)
							delete(fallBackActions, block.Pos)
							fallBackActionsMu.Unlock()
						}
					}, time.Second*3)
					for time.Since(placeStart) < time.Millisecond*50 {
						if !quickDone {
							time.Sleep(time.Millisecond)
						} else {
							break
						}
					}
				} else {
					pterm.Error.Println("无法从NBT: %v 获得命令方块数据 %v", block.NBT, err)
				}

			} else {
				cmd := fmt.Sprintf("setblock %v %v %v %v %v", block.Pos[0], block.Pos[1], block.Pos[2], strings.Replace(blk.Name, "minecraft:", "", 1), blk.Val)
				o.BlockCmdSender(cmd)
			}
			counter++
		}
		doDelay()
		// if block.NBT != nil && !o.IgnoreNbt {
		// 	o.delayBlocksMu.Lock()
		// 	o.delayBlocks[block.Pos] = block
		// 	o.delayBlocksMu.Unlock()
		// }
	}
	fallBackActionsMu.Lock()
	if !o.IgnoreNbt && len(fallBackActions) > 0 {
		fallBackActionsMu.Unlock()
		time.Sleep(time.Second * 3)
		fallBackActionsMu.Lock()
		forceActions := make(map[define.CubePos]func())
		for pos, actions := range fallBackActions {
			forceActions[pos] = actions
		}
		fallBackActions = make(map[define.CubePos]func())
		fallBackActionsMu.Unlock()
		for pos, action := range fallBackActions {
			pterm.Warning.Printfln("Force Execute Fallback Actions @ %v", pos)
			action()
		}
	} else {
		fallBackActionsMu.Unlock()
	}
	// o.delayBlocksMu.Lock()
	// if len(o.delayBlocks) > 0 && !o.IgnoreNbt && !o.Stop {
	// 	counter := 0
	// 	for len(o.delayBlocks) > 0 {
	// 		o.delayBlocksMu.Unlock()
	// 		time.Sleep(time.Duration(o.FinalWaitTime) * time.Second)
	// 		o.delayBlocksMu.Lock()
	// 		counter++
	// 		if counter == 10 {
	// 			break
	// 		}
	// 	}
	// 	o.delayBlocksMu.Unlock()
	// 	if len(o.delayBlocks) > 0 {
	// 		pterm.Warning.Println("强制写入方块 NBT 信息")
	// 	}
	// 	o.updateDelayBlocks(true)
	// } else {
	// 	o.delayBlocksMu.Unlock()

	// o.ProgressUpdater(-1)
}

func (o *Builder) OnBlockUpdate(pos define.CubePos, origRTID, currentRTID uint32) {
	// fmt.Println("update ", pos, origRTID, currentRTID)
	// if o.delayBlocks != nil {
	// 	o.delayBlocksMu.Lock()
	// 	b := o.delayBlocks[pos]
	// 	o.delayBlocksMu.Unlock()
	// 	if b == nil {
	// 		return
	// 	}
	// 	if b.RTID == currentRTID {
	// 		fmt.Println("hit ", pos, origRTID, currentRTID)
	// 		b.Hit = true
	// 		o.delayBlocksMu.Lock()
	// 		o.SetBlock(pos, b)
	// 		o.delayBlocksMu.Unlock()
	// 	}
	// }
}
func (o *Builder) OnLevelChunk(cd *mirror.ChunkData) {
	// if o.delayBlocks != nil {
	// 	blocksToRemove := make(map[define.CubePos]*IOBlockForBuilder)
	// 	o.delayBlocksMu.Lock()
	// 	for pos, b := range o.delayBlocks {
	// 		if b.Hit {
	// 			continue
	// 		}
	// 		chunkPos := define.ChunkPos{int32(pos[0] >> 4), int32(pos[2] >> 4)}
	// 		if chunkPos != cd.ChunkPos {
	// 			continue
	// 		}
	// 		newRTID := cd.Chunk.Block(uint8(pos.X()), int16(pos.Y()), uint8(pos.Z()), 0)
	// 		// fmt.Println(chunk.RuntimeIDToLegacyBlock(newRTID), chunk.RuntimeIDToLegacyBlock(b.RTID))
	// 		if b.RTID == newRTID {
	// 			// fmt.Println("hit ", pos, newRTID)
	// 			blocksToRemove[pos] = b
	// 			b.Hit = true
	// 		}
	// 	}
	// 	for pos, blk := range blocksToRemove {
	// 		o.SetBlock(pos, blk)
	// 	}
	// 	o.delayBlocksMu.Unlock()
	// }
}

// func (o *Builder) SetBlock(pos define.CubePos, block *IOBlockForBuilder) {
// 	delete(o.delayBlocks, pos)
// 	switch block.NBT["id"] {
// 	case "sign":
// 		// nbt := block.BlockNbt
// 		// fmt.Println("sign: ", nbt)
// 		// o.Frame.GetGameControl().GetInteraction().WritePacket(&packet.BlockActorData{
// 		// 	Position: protocol.BlockPos{int32(pos[0]), int32(pos[1]), int32(pos[2])},
// 		// 	NBTData:  nbt,
// 		// })
// 	case "CommandBlock":
// 		pterm.Info.Println("DEBUG: CommandBlock: ", block.Pos)
// 		item := block.NBT
// 		blk := chunk.RuntimeIDToLegacyBlock(block.RTID)
// 		curblockname := blk.Name
// 		var mode uint32
// 		if curblockname == "command_block" {
// 			mode = packet.CommandBlockImpulse
// 		} else if curblockname == "repeating_command_block" {
// 			mode = packet.CommandBlockRepeating
// 		} else if curblockname == "chain_command_block" {
// 			mode = packet.CommandBlockChain
// 		}
// 		cmd, _ := item["Command"].(string)
// 		cusname, _ := item["CustomName"].(string)
// 		exeft, _ := item["ExecuteOnFirstTick"].(uint8)
// 		tickdelay, _ := item["TickDelay"].(int32)     //*/
// 		aut, _ := item["auto"].(uint8)                //!needrestone
// 		trackoutput, _ := item["TrackOutput"].(uint8) //
// 		lo, _ := item["LastOutput"].(string)
// 		conditionalmode := item["conditionalMode"].(uint8)
// 		var exeftb bool
// 		if exeft == 0 {
// 			exeftb = false
// 		} else {
// 			exeftb = true
// 		}
// 		var tob bool
// 		if trackoutput == 1 {
// 			tob = true
// 		} else {
// 			tob = false
// 		}
// 		var nrb bool
// 		if aut == 1 {
// 			nrb = false
// 			//REVERSED!!
// 		} else {
// 			nrb = true
// 		}
// 		var conb bool
// 		if conditionalmode == 1 {
// 			conb = true
// 		} else {
// 			conb = false
// 		}
// 		o.Ctrl.SendCmd(fmt.Sprintf("tp @s %v %v %v", block.Pos.X(), block.Pos.Y(), block.Pos.Z()))
// 		time.Sleep(50 * time.Millisecond)
// 		o.Frame.GetGameControl().GetInteraction().WritePacket(&packet.CommandBlockUpdate{
// 			Block:              true,
// 			Position:           protocol.BlockPos{int32(block.Pos.X()), int32(block.Pos.Y()), int32(block.Pos.Z())},
// 			Mode:               mode,
// 			NeedsRedstone:      nrb,
// 			Conditional:        conb,
// 			Command:            cmd,
// 			LastOutput:         lo,
// 			Name:               cusname,
// 			TickDelay:          tickdelay,
// 			ExecuteOnFirstTick: exeftb,
// 			ShouldTrackOutput:  tob,
// 		})
// 		time.Sleep(50 * time.Millisecond)
// 	}

// }

// func (o *Builder) updateDelayBlocks(force bool) {
// 	// fmt.Println("DEBUG: DO UPDATE DelayBlocks ", force)
// 	// for pos, block := range o.delayBlocks {
// 	// 	// fmt.Println("DEBUG: ", block.Pos, block.Hit)
// 	// 	if o.Stop {
// 	// 		return
// 	// 	}
// 	// 	// fmt.Println(block)
// 	// 	if block.Hit || force {
// 	// 		if force {
// 	// 			pterm.Warning.Printfln("强制向方块 %v 写入信息", pos)
// 	// 		}
// 	// 		o.SetBlock(pos, block)
// 	// 	}
// 	// }
// }
