package block_actors

import (
	"phoenixbuilder/fastbuilder/utils"
	"phoenixbuilder/minecraft/protocol"
	general "phoenixbuilder/minecraft/protocol/block_actors/general_actors"
)

// 讲台
type Lectern struct {
	general.BlockActor
	Book       protocol.Item `nbt:"book"`       // TAG_Compound(10)
	HasBook    byte          `nbt:"hasBook"`    // TAG_Byte(1) = 0
	Page       int32         `nbt:"page"`       // TAG_Int(4) = 0
	TotalPages int32         `nbt:"totalPages"` // TAG_Int(4) = 1
}

// ID ...
func (*Lectern) ID() string {
	return IDLectern
}

// 检查 x 是否存在 Lectern 中记录的所有数据
func (l *Lectern) CheckExist(x map[string]any) (exist bool) {
	_, exist1 := x["book"]
	_, exist2 := x["hasBook"]
	_, exist3 := x["page"]
	_, exist4 := x["totalPages"]
	return exist1 && exist2 && exist3 && exist4
}

func (l *Lectern) Marshal(io protocol.IO) {
	protocol.Single(io, &l.BlockActor)
	io.Uint8(&l.HasBook)

	if l.HasBook == 1 {
		io.Varint32(&l.Page)
		io.Varint32(&l.TotalPages)
		protocol.Single(io, &l.Book)
	}
}

func (l *Lectern) ToNBT() map[string]any {
	globalMap := l.BlockActor.ToNBT()
	if l.HasBook == 1 {
		return utils.MergeMaps(
			globalMap,
			map[string]any{
				"book":       l.Book.ToNBT(),
				"hasBook":    l.HasBook,
				"page":       l.Page,
				"totalPages": l.TotalPages,
			},
		)
	}
	return globalMap
}

func (l *Lectern) FromNBT(x map[string]any) {
	l.BlockActor.FromNBT(x)

	if l.CheckExist(x) {
		l.Book.FromNBT(x["book"].(map[string]any))
		l.HasBook = x["hasBook"].(byte)
		l.Page = x["page"].(int32)
		l.TotalPages = x["totalPages"].(int32)
	}
}
