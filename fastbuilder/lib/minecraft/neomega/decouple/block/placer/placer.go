package placer

import (
	"fmt"
	"phoenixbuilder/fastbuilder/lib/minecraft/mirror/define"
	"phoenixbuilder/fastbuilder/lib/minecraft/neomega/omega"
	"phoenixbuilder/minecraft/protocol/packet"
	"strings"
	"sync"
	"time"
)

func init() {
	if false {
		func(omega.BlockPlacer) {}(&BlockPlacer{})
	}
}

type BlockPlacer struct {
	onBlockActorCbs map[define.CubePos]func(define.CubePos, *packet.BlockActorData)
	blockActorLock  sync.Mutex
	omega.CmdSender
	omega.GameIntractable
}

func NewBlockPlacer(reactable omega.ReactCore, cmdSender omega.CmdSender, packetSender omega.GameIntractable) omega.BlockPlacer {
	c := &BlockPlacer{
		onBlockActorCbs: make(map[define.CubePos]func(define.CubePos, *packet.BlockActorData)),
		blockActorLock:  sync.Mutex{},
		CmdSender:       cmdSender,
		GameIntractable: packetSender,
	}

	reactable.SetOnTypedPacketCallBack(packet.IDBlockActorData, func(p packet.Packet) {
		c.onBlockActor(p.(*packet.BlockActorData))
	})
	return c
}

func (c *BlockPlacer) onBlockActor(p *packet.BlockActorData) {
	pos := define.CubePos{int(p.Position.X()), int(p.Position.Y()), int(p.Position.Z())}
	c.blockActorLock.Lock()
	if cb, found := c.onBlockActorCbs[pos]; found {
		delete(c.onBlockActorCbs, pos)
		c.blockActorLock.Unlock()
		cb(pos, p)
	} else {
		c.blockActorLock.Unlock()
	}
}

func (g *BlockPlacer) PlaceCommandBlock(pos define.CubePos, commandBlockName string, blockDataOrStateStr string,
	withMove, withAirPrePlace bool, updatePacket *packet.CommandBlockUpdate,
	onDone func(done bool), timeOut time.Duration) {
	done := make(chan bool)
	go func() {
		select {
		case <-time.NewTimer(timeOut).C:
			onDone(false)
		case <-done:
		}
	}()
	go func() {
		if withMove {
			g.SendWSCmd(fmt.Sprintf("tp @s %v %v %v", pos.X(), pos.Y(), pos.Z()))
			time.Sleep(100 * time.Millisecond)
		}
		if withAirPrePlace {
			cmd := fmt.Sprintf("setblock %v %v %v %v %v", pos[0], pos[1], pos[2], "air", 0)
			g.SendWOCmd(cmd)
			time.Sleep(100 * time.Millisecond)
		} else {
			g.SendPacket(updatePacket)
		}
		cmd := fmt.Sprintf("setblock %v %v %v %v %v", pos[0], pos[1], pos[2], strings.Replace(commandBlockName, "minecraft:", "", 1), blockDataOrStateStr)
		g.SendWOCmd(cmd)
		g.blockActorLock.Lock()
		g.onBlockActorCbs[pos] = func(cp define.CubePos, bad *packet.BlockActorData) {
			go func() {
				g.blockActorLock.Lock()
				// g.SendWSCmd(fmt.Sprintf("tp @s %v %v %v", pos.X(), pos.Y(), pos.Z()))
				// time.Sleep(50 * time.Millisecond)
				g.SendPacket(updatePacket)
				g.onBlockActorCbs[pos] = func(cp define.CubePos, bad *packet.BlockActorData) {
					g.blockActorLock.Lock()
					delete(g.onBlockActorCbs, pos)
					g.blockActorLock.Unlock()
					g.SendPacket(updatePacket)
					onDone(true)
					done <- true
				}
				g.blockActorLock.Unlock()
			}()
		}
		g.blockActorLock.Unlock()

	}()
}

func (g *BlockPlacer) PlaceSignBlock(pos define.CubePos, signBlockName string, blockDataOrStateStr string, withMove, withAirPrePlace bool, updatePacket *packet.BlockActorData, onDone func(done bool), timeOut time.Duration) {
	done := make(chan bool)
	go func() {
		select {
		case <-time.NewTimer(timeOut).C:
			onDone(false)
		case <-done:
		}
	}()
	go func() {
		if withMove {
			g.SendWSCmd(fmt.Sprintf("tp @s %v %v %v", pos.X(), pos.Y(), pos.Z()))
			time.Sleep(100 * time.Millisecond)
		}
		if withAirPrePlace {
			cmd := fmt.Sprintf("setblock %v %v %v %v %v", pos[0], pos[1], pos[2], "air", 0)
			g.SendWOCmd(cmd)
			time.Sleep(100 * time.Millisecond)
		}
		cmd := fmt.Sprintf("setblock %v %v %v %v %v", pos[0], pos[1], pos[2], strings.Replace(signBlockName, "minecraft:", "", 1), blockDataOrStateStr)
		g.SendWOCmd(cmd)
		g.blockActorLock.Lock()
		g.onBlockActorCbs[pos] = func(cp define.CubePos, bad *packet.BlockActorData) {
			go func() {
				g.blockActorLock.Lock()
				g.SendWSCmd(fmt.Sprintf("tp @s %v %v %v", pos.X(), pos.Y(), pos.Z()))
				time.Sleep(50 * time.Millisecond)
				g.SendPacket(updatePacket)
				g.onBlockActorCbs[pos] = func(cp define.CubePos, bad *packet.BlockActorData) {
					onDone(true)
					done <- true
				}
				g.blockActorLock.Unlock()
			}()
		}
		g.blockActorLock.Unlock()
	}()
}
