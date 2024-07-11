package block_actors

import (
	"phoenixbuilder/fastbuilder/utils"
	"phoenixbuilder/minecraft/protocol"
	general "phoenixbuilder/minecraft/protocol/block_actors/general_actors"
)

// 命令方块
type CommandBlock struct {
	general.BlockActor
	Command            string   `nbt:"Command"`            // TAG_String(8) = ""
	CustomName         string   `nbt:"CustomName"`         // TAG_String(8) = ""
	ExecuteOnFirstTick byte     `nbt:"ExecuteOnFirstTick"` // TAG_Byte(1) = 0
	LPCommandMode      int32    `nbt:"LPCommandMode"`      // TAG_Int(4) = 0
	LPCondionalMode    byte     `nbt:"LPCondionalMode"`    // TAG_Byte(1) = 0
	LPRedstoneMode     byte     `nbt:"LPRedstoneMode"`     // TAG_Byte(1) = 0
	LastExecution      int64    `nbt:"LastExecution"`      // TAG_Long(5) = 0
	LastOutput         string   `nbt:"LastOutput"`         // TAG_Byte(8) = ""
	LastOutputParams   []string `nbt:"LastOutputParams"`   // TAG_List[TAG_String] (9[8]) = []
	SuccessCount       int32    `nbt:"SuccessCount"`       // TAG_Int(4) = 0
	TickDelay          int32    `nbt:"TickDelay"`          // TAG_Int(4) = 0
	TrackOutput        byte     `nbt:"TrackOutput"`        // TAG_Byte(1) = 1
	Version            int32    `nbt:"Version"`            // TAG_Int(4) = 35
	Auto               byte     `nbt:"auto"`               // TAG_Byte(1) = 1
	ConditionMet       byte     `nbt:"conditionMet"`       // TAG_Byte(1) = 0
	ConditionalMode    byte     `nbt:"conditionalMode"`    // Not used; TAG_Byte(1) = 0
	Powered            byte     `nbt:"powered"`            // TAG_Byte(1) = 0
}

// ID ...
func (*CommandBlock) ID() string {

	return IDCommandBlock
}

func (c *CommandBlock) Marshal(io protocol.IO) {
	protocol.Single(io, &c.BlockActor)
	io.Uint8(&c.Powered)
	io.Uint8(&c.Auto)
	io.Uint8(&c.ConditionMet)
	io.Uint8(&c.LPCondionalMode)
	io.Uint8(&c.LPRedstoneMode)
	io.Varint32(&c.LPCommandMode)
	io.String(&c.Command)
	io.Varint32(&c.Version)
	io.Varint32(&c.SuccessCount)
	io.String(&c.CustomName)
	io.String(&c.LastOutput)
	protocol.FuncSliceVarint32Length(io, &c.LastOutputParams, io.String)
	io.Uint8(&c.TrackOutput)
	io.Varint64(&c.LastExecution)
	io.Varint32(&c.TickDelay)
	io.Uint8(&c.ExecuteOnFirstTick)
}

func (c *CommandBlock) ToNBT() map[string]any {
	return utils.MergeMaps(
		c.BlockActor.ToNBT(),
		map[string]any{
			"powered":            c.Powered,
			"Command":            c.Command,
			"CustomName":         c.CustomName,
			"ExecuteOnFirstTick": c.ExecuteOnFirstTick,
			"LPCommandMode":      c.LPCommandMode,
			"LPCondionalMode":    c.LPCondionalMode,
			"LPRedstoneMode":     c.LPRedstoneMode,
			"LastExecution":      c.LastExecution,
			"LastOutput":         c.LastOutput,
			"LastOutputParams":   utils.ToAnyList(c.LastOutputParams),
			"SuccessCount":       c.SuccessCount,
			"TickDelay":          c.TickDelay,
			"TrackOutput":        c.TrackOutput,
			"Version":            c.Version,
			"auto":               c.Auto,
			"conditionMet":       c.ConditionMet,
			"conditionalMode":    c.ConditionalMode,
		},
	)
}

func (c *CommandBlock) FromNBT(x map[string]any) {
	c.BlockActor.FromNBT(x)
	c.Powered = x["powered"].(byte)
	c.Command = x["Command"].(string)
	c.CustomName = x["CustomName"].(string)
	c.ExecuteOnFirstTick = x["ExecuteOnFirstTick"].(byte)
	c.LPCommandMode = x["LPCommandMode"].(int32)
	c.LPCondionalMode = x["LPCondionalMode"].(byte)
	c.LPRedstoneMode = x["LPRedstoneMode"].(byte)
	c.LastExecution = x["LastExecution"].(int64)
	c.LastOutput = x["LastOutput"].(string)
	c.LastOutputParams = utils.FromAnyList[string](x["LastOutputParams"].([]any))
	c.SuccessCount = x["SuccessCount"].(int32)
	c.TickDelay = x["TickDelay"].(int32)
	c.TrackOutput = x["TrackOutput"].(byte)
	c.Version = x["Version"].(int32)
	c.Auto = x["auto"].(byte)
	c.ConditionMet = x["conditionMet"].(byte)
	c.ConditionalMode = x["conditionalMode"].(byte)
}
