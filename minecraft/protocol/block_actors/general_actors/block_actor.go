package general

import (
	"phoenixbuilder/minecraft/protocol"
)

// 描述一个方块实体的通用字段
type BlockActor struct {
	IsMovable      byte   `mapstructure:"isMovable"`      // TAG_Byte(1) = 1
	CustomName     string `mapstructure:"CustomName"`     // TAG_String(8) = ""
	UserCustomData string `mapstructure:"UserCustomData"` // TAG_String(8) = ""
}

func (g *BlockActor) Marshal(r protocol.IO) {
	r.Uint8(&g.IsMovable)
	r.String(&g.CustomName)
	r.String(&g.UserCustomData)
}
