package utils

import (
	"fmt"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/mirror/define"
)

func GenCommandBlockUpdateFromNbt(pos define.CubePos, blockName string, nbt map[string]interface{}) (cfg *packet.CommandBlockUpdate, err error) {
	var mode uint32
	defer func() {
		r := recover()
		if r != nil {
			err = fmt.Errorf("cannot gen block update %v", r)
			cfg = nil
		}
	}()
	if blockName == "command_block" {
		mode = packet.CommandBlockImpulse
	} else if blockName == "repeating_command_block" {
		mode = packet.CommandBlockRepeating
	} else if blockName == "chain_command_block" {
		mode = packet.CommandBlockChain
	} else {
		return nil, fmt.Errorf("not command block")
	}
	cmd, _ := nbt["Command"].(string)
	cusname, _ := nbt["CustomName"].(string)
	exeft, _ := nbt["ExecuteOnFirstTick"].(uint8)
	tickdelay, _ := nbt["TickDelay"].(int32)     //*/
	aut, _ := nbt["auto"].(uint8)                //!needrestone
	trackoutput, _ := nbt["TrackOutput"].(uint8) //
	lo, _ := nbt["LastOutput"].(string)
	conditionalmode := nbt["conditionalMode"].(uint8)
	var exeftb bool
	if exeft == 0 {
		exeftb = false
	} else {
		exeftb = true
	}
	var tob bool
	if trackoutput == 1 {
		tob = true
	} else {
		tob = false
	}
	var nrb bool
	if aut == 1 {
		nrb = false
		//REVERSED!!
	} else {
		nrb = true
	}
	var conb bool
	if conditionalmode == 1 {
		conb = true
	} else {
		conb = false
	}
	return &packet.CommandBlockUpdate{
		Block:              true,
		Position:           protocol.BlockPos{int32(pos.X()), int32(pos.Y()), int32(pos.Z())},
		Mode:               mode,
		NeedsRedstone:      nrb,
		Conditional:        conb,
		Command:            cmd,
		LastOutput:         lo,
		Name:               cusname,
		TickDelay:          tickdelay,
		ExecuteOnFirstTick: exeftb,
		ShouldTrackOutput:  tob,
	}, nil
}
