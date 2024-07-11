package fields

import (
	"phoenixbuilder/fastbuilder/utils"
	"phoenixbuilder/minecraft/protocol"
)

// ------------------------- SpawnData -------------------------

// 描述 刷怪笼 中的一个复用字段
type SpawnData struct {
	Properties map[string]any `nbt:"Properties"` // TAG_Compound(10)
	TypeID     string         `nbt:"TypeId"`     // TAG_String(8)
	Weight     int32          `nbt:"Weight"`     // TAG_Int(4)
}

func (s *SpawnData) Marshal(r protocol.IO) {
	r.NBTWithLength(&s.Properties)
	r.String(&s.TypeID)
	r.Varint32(&s.Weight)
}

func (s *SpawnData) ToNBT() map[string]any {
	return map[string]any{
		"Properties": s.Properties,
		"TypeId":     s.TypeID,
		"Weight":     s.Weight,
	}
}

func (s *SpawnData) FromNBT(x map[string]any) {
	s.Properties = x["Properties"].(map[string]any)
	s.TypeID = x["TypeId"].(string)
	s.Weight = x["Weight"].(int32)
}

// ------------------------- MultiSpawnData -------------------------

// 描述多个 SpawnData
type MultiSpawnData struct {
	Data []SpawnData // TAG_List[TAG_Compound] (9[10])
}

func (m *MultiSpawnData) Marshal(r protocol.IO) {
	protocol.SliceVarint16Length(r, &m.Data)
}

func (m *MultiSpawnData) ToNBT() []any {
	return utils.ToAnyList(m.Data)
}

func (m *MultiSpawnData) FromNBT(x []any) {
	m.Data = utils.FromAnyList[SpawnData](x)
}
