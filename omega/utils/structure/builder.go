package structure

import (
	"fmt"
	"phoenixbuilder/mirror/chunk"
	"phoenixbuilder/mirror/define"
	"strings"
	"sync"
	"time"

	"github.com/pterm/pterm"
)

type Builder struct {
	delayBlocks     map[define.CubePos]*IOBlock
	delayBlocksMu   sync.RWMutex
	BlockCmdSender  func(cmd string)
	TpCmdSender     func(cmd string)
	ProgressUpdater func(currBlock int)
	FinalWaitTime   int
	IgnoreNbt       bool
	Stop            bool
	InitPosGetter   func() define.CubePos
}

func (o *Builder) Build(blocksIn chan *IOBlock, speed int) {
	o.delayBlocks = make(map[define.CubePos]*IOBlock)
	o.delayBlocksMu = sync.RWMutex{}
	counter := 0
	delay := time.Duration((float64(1000) / float64(speed) * float64(time.Millisecond)))
	ticker := time.NewTicker(delay)
	lastPos := o.InitPosGetter()
	pterm.Info.Printfln("DEBUG: Init Pos: %v", lastPos)
	for block := range blocksIn {
		if o.Stop {
			return
		}
		xmove := block.Pos.X() - lastPos.X()
		zmove := block.Pos.Z() - lastPos.Z()
		if counter == 0 {
			o.TpCmdSender(fmt.Sprintf("tp @s %v %v %v", block.Pos[0], 320, block.Pos[2]))
			lastPos = block.Pos
			time.Sleep(3 * time.Second)
		}
		if (xmove*xmove + zmove*zmove) > 16*16 {
			o.TpCmdSender(fmt.Sprintf("tp @s %v %v %v", block.Pos[0], 320, block.Pos[2]))
		}
		lastPos = block.Pos
		blk := chunk.RuntimeIDToLegacyBlock(block.RTID)
		if blk == nil {
			continue
		}
		cmd := fmt.Sprintf("setblock %v %v %v %v %v", block.Pos[0], block.Pos[1], block.Pos[2], strings.ReplaceAll(blk.Name, "minecraft:", ""), blk.Val)
		o.BlockCmdSender(cmd)
		o.ProgressUpdater(counter)

		counter++
		<-ticker.C
		if block.NBT != nil && !o.IgnoreNbt {
			o.delayBlocksMu.Lock()
			o.delayBlocks[block.Pos] = block
			// fmt.Println("B: ", o.delayBlocks[block.pos])
			o.delayBlocksMu.Unlock()
			if len(o.delayBlocks) > 64 {
				o.updateDelayBlocks(false)
			}
		}
	}
	o.delayBlocksMu.RLock()
	if len(o.delayBlocks) > 0 && !o.IgnoreNbt && !o.Stop {
		o.delayBlocksMu.RUnlock()
		time.Sleep(time.Duration(o.FinalWaitTime) * time.Second)
		o.updateDelayBlocks(true)
	} else {
		o.delayBlocksMu.RUnlock()
	}

	o.ProgressUpdater(-1)
}

func (o *Builder) OnBlockUpdate(pos define.CubePos, origRTID, currentRTID uint32) {
	if o.delayBlocks != nil {
		o.delayBlocksMu.RLock()
		b := o.delayBlocks[pos]
		o.delayBlocksMu.RUnlock()
		if b == nil {
			return
		}
		if b.RTID == origRTID || b.RTID == currentRTID {
			b.Hit = true
		}
	}
}

func (o *Builder) updateDelayBlocks(force bool) {
	for pos, block := range o.delayBlocks {
		if o.Stop {
			return
		}
		// fmt.Println(block)
		if block.Hit || force {
			switch block.NBT["id"] {
			case "sign":
				// nbt := block.BlockNbt
				// fmt.Println("sign: ", nbt)
				// o.Frame.GetGameControl().SendMCPacket(&packet.BlockActorData{
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
