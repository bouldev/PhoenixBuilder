package block_actors

import (
	"phoenixbuilder/minecraft/protocol"
	general "phoenixbuilder/minecraft/protocol/block_actors/general_actors"
)

type CommandBlock struct {
	general.BlockActor `mapstructure:",squash"`
	Command            string `mapstructure:"Command"`            // TAG_String(8) = ""
	CustomName         string `mapstructure:"CustomName"`         // TAG_String(8) = ""
	ExecuteOnFirstTick byte   `mapstructure:"ExecuteOnFirstTick"` // TAG_Byte(1) = 0
	LPCommandMode      int32  `mapstructure:"LPCommandMode"`      // TAG_Int(4) = 0
	LPCondionalMode    byte   `mapstructure:"LPCondionalMode"`    // TAG_Byte(1) = 0
	LPRedstoneMode     byte   `mapstructure:"LPRedstoneMode"`     // TAG_Byte(1) = 0
	LastExecution      int64  `mapstructure:"LastExecution"`      // TAG_Long(5) = 0
	LastOutput         string `mapstructure:"LastOutput"`         // TAG_Byte(8) = ""
	LastOutputParams   []any  `mapstructure:"LastOutputParams"`   // TAG_List[TAG_String] (9[8]) = []
	SuccessCount       int32  `mapstructure:"SuccessCount"`       // TAG_Int(4) = 0
	TickDelay          int32  `mapstructure:"TickDelay"`          // TAG_Int(4) = 0
	TrackOutput        byte   `mapstructure:"TrackOutput"`        // TAG_Byte(1) = 1
	Version            int32  `mapstructure:"Version"`            // TAG_Int(4) = 35
	Auto               byte   `mapstructure:"auto"`               // TAG_Byte(1) = 1
	ConditionMet       byte   `mapstructure:"conditionMet"`       // TAG_Byte(1) = 0
	Powered            byte   `mapstructure:"powered"`            // TAG_Byte(1) = 0

	/*
		Types for command blocks are checked by their names.
		Whether a command block is conditional is checked through its data value.
		SINCE IT IS NOT INCLUDED IN NBT DATA!!!

		(Wrote by LNSSPsd/Ruphane, added by Happy2018new)
	*/
	ConditionalMode byte `mapstructure:"conditionalMode"` // Not used; TAG_Byte(1) = 0
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
	protocol.NBTSlice(io, &c.LastOutputParams, func(t *[]string) { protocol.FuncSliceVarint32Length(io, t, io.String) })
	io.Uint8(&c.TrackOutput)
	io.Varint64(&c.LastExecution)
	io.Varint32(&c.TickDelay)
	io.Uint8(&c.ExecuteOnFirstTick)
}
