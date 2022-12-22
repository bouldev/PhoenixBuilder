package structure

import (
	"errors"
	"phoenixbuilder/mirror/define"
)

type IOBlockForBuilder struct {
	Pos         define.CubePos
	BlockName   string // for operation 13 which named `PlaceBlockWithBlockStates`
	BlockStates string // for operation 13 which named `PlaceBlockWithBlockStates`
	BlockData   uint16 // for operation 7 which named `PlaceBlock`
	RTID        uint32
	NBT         map[string]interface{}
	Hit         bool
	Expand16    bool
}

type IOBlockForDecoder struct {
	Pos         define.CubePos
	BlockName   string // for operation 13 which named `PlaceBlockWithBlockStates`
	BlockStates string // for operation 13 which named `PlaceBlockWithBlockStates`
	BlockData   uint16 // for operation 7 which named `PlaceBlock`
	RTID        uint32
	NBT         map[string]interface{}
}

type CommandBlockNBT struct {
	Command            string
	CustomName         string
	ExecuteOnFirstTick uint8
	TickDelay          int32
	Auto               uint8 `nbt:"auto"` // need redstone
	TrackOutput        uint8
	LastOutput         string
	ConditionalMode    uint8 `nbt:"conditionalMode"`
	Data               int32 `nbt:"data"`
}

var ErrImportFormatNotSupport = errors.New("format unsupported")
