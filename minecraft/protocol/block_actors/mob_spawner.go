package block_actors

import (
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/block_actors/fields"
	general "phoenixbuilder/minecraft/protocol/block_actors/general_actors"
)

// 刷怪笼
type MobSpawner struct {
	general.BlockActor  `mapstructure:",squash"`
	Delay               int16             `mapstructure:"Delay"`                     // TAG_Short(3) = 20
	DisplayEntityHeight float32           `mapstructure:"DisplayEntityHeight"`       // TAG_Float(6) = 1.7999999523162842
	DisplayEntityScale  float32           `mapstructure:"DisplayEntityScale"`        // TAG_Float(6) = 1
	DisplayEntityWidth  float32           `mapstructure:"DisplayEntityWidth"`        // TAG_Float(6) = 0.800000011920929
	EntityIdentifier    string            `mapstructure:"EntityIdentifier"`          // TAG_String(8) = ""
	MaxNearbyEntities   int16             `mapstructure:"MaxNearbyEntities"`         // TAG_Short(3) = 6
	MaxSpawnDelay       int16             `mapstructure:"MaxSpawnDelay"`             // TAG_Short(3) = 800
	MinSpawnDelay       int16             `mapstructure:"MinSpawnDelay"`             // TAG_Short(3) = 200
	RequiredPlayerRange int16             `mapstructure:"RequiredPlayerRange"`       // TAG_Short(3) = 16
	SpawnCount          int16             `mapstructure:"SpawnCount"`                // TAG_Short(3) = 4
	SpawnRange          int16             `mapstructure:"SpawnRange"`                // TAG_Short(3) = 4
	SpawnData           *fields.SpawnData `mapstructure:"SpawnData,omitempty"`       // TAG_Compound(10)
	SpawnPotentials     []any             `mapstructure:"SpawnPotentials,omitempty"` // TAG_List[TAG_Compound] (9[10])
}

// ID ...
func (*MobSpawner) ID() string {
	return IDMobSpawner
}

func (m *MobSpawner) Marshal(io protocol.IO) {
	f := func() *fields.SpawnData {
		if m.SpawnData == nil {
			m.SpawnData = new(fields.SpawnData)
		}
		return m.SpawnData
	}

	protocol.Single(io, &m.BlockActor)
	io.String(&m.EntityIdentifier)
	io.Varint16(&m.Delay)
	io.Varint16(&m.MinSpawnDelay)
	io.Varint16(&m.MaxSpawnDelay)
	io.Varint16(&m.SpawnCount)
	io.Varint16(&m.MaxNearbyEntities)
	io.Varint16(&m.RequiredPlayerRange)
	io.Varint16(&m.SpawnRange)
	io.Float32(&m.DisplayEntityWidth)
	io.Float32(&m.DisplayEntityHeight)
	io.Float32(&m.DisplayEntityScale)
	protocol.NBTOptionalSliceVarint16Length(io, &m.SpawnPotentials, &[]fields.SpawnData{})
	protocol.NBTOptionalMarshaler(io, m.SpawnData, f, true)
}
