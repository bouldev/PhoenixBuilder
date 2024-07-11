package block_actors

import (
	"phoenixbuilder/fastbuilder/utils"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/block_actors/fields"
	general "phoenixbuilder/minecraft/protocol/block_actors/general_actors"
)

// 刷怪笼
type MobSpawner struct {
	general.BlockActor
	Delay               int16                                    `nbt:"Delay"`               // TAG_Short(3) = 20
	DisplayEntityHeight float32                                  `nbt:"DisplayEntityHeight"` // TAG_Float(6) = 1.7999999523162842
	DisplayEntityScale  float32                                  `nbt:"DisplayEntityScale"`  // TAG_Float(6) = 1
	DisplayEntityWidth  float32                                  `nbt:"DisplayEntityWidth"`  // TAG_Float(6) = 0.800000011920929
	EntityIdentifier    string                                   `nbt:"EntityIdentifier"`    // TAG_String(8) = ""
	MaxNearbyEntities   int16                                    `nbt:"MaxNearbyEntities"`   // TAG_Short(3) = 6
	MaxSpawnDelay       int16                                    `nbt:"MaxSpawnDelay"`       // TAG_Short(3) = 800
	MinSpawnDelay       int16                                    `nbt:"MinSpawnDelay"`       // TAG_Short(3) = 200
	RequiredPlayerRange int16                                    `nbt:"RequiredPlayerRange"` // TAG_Short(3) = 16
	SpawnCount          int16                                    `nbt:"SpawnCount"`          // TAG_Short(3) = 4
	SpawnRange          int16                                    `nbt:"SpawnRange"`          // TAG_Short(3) = 4
	SpawnData           protocol.Optional[fields.SpawnData]      `nbt:"SpawnData"`           // TAG_Compound(10)
	SpawnPotentials     protocol.Optional[fields.MultiSpawnData] `nbt:"SpawnPotentials"`     // TAG_List(9)
}

// ID ...
func (*MobSpawner) ID() string {
	return IDMobSpawner
}

func (m *MobSpawner) Marshal(io protocol.IO) {
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
	protocol.OptionalMarshaler(io, &m.SpawnPotentials)
	protocol.OptionalMarshaler(io, &m.SpawnData)
}

func (m *MobSpawner) ToNBT() map[string]any {
	var temp map[string]any
	if spawnData, has := m.SpawnData.Value(); has {
		temp = map[string]any{
			"SpawnData": spawnData.ToNBT(),
		}
	}
	if spawnPotentials, has := m.SpawnPotentials.Value(); has {
		utils.MergeMaps(
			map[string]any{
				"SpawnPotentials": spawnPotentials.ToNBT(),
			}, temp,
		)
	}
	return utils.MergeMaps(
		m.BlockActor.ToNBT(),
		map[string]any{
			"Delay":               m.Delay,
			"DisplayEntityHeight": m.DisplayEntityHeight,
			"DisplayEntityScale":  m.DisplayEntityScale,
			"DisplayEntityWidth":  m.DisplayEntityWidth,
			"EntityIdentifier":    m.EntityIdentifier,
			"MaxNearbyEntities":   m.MaxNearbyEntities,
			"MaxSpawnDelay":       m.MaxSpawnDelay,
			"MinSpawnDelay":       m.MinSpawnDelay,
			"RequiredPlayerRange": m.RequiredPlayerRange,
			"SpawnCount":          m.SpawnCount,
			"SpawnRange":          m.SpawnRange,
		},
		temp,
	)
}

func (m *MobSpawner) FromNBT(x map[string]any) {
	m.BlockActor.FromNBT(x)
	m.Delay = x["Delay"].(int16)
	m.DisplayEntityHeight = x["DisplayEntityHeight"].(float32)
	m.DisplayEntityScale = x["DisplayEntityScale"].(float32)
	m.DisplayEntityWidth = x["DisplayEntityWidth"].(float32)
	m.EntityIdentifier = x["EntityIdentifier"].(string)
	m.MaxNearbyEntities = x["MaxNearbyEntities"].(int16)
	m.MaxSpawnDelay = x["MaxSpawnDelay"].(int16)
	m.MinSpawnDelay = x["MinSpawnDelay"].(int16)
	m.RequiredPlayerRange = x["RequiredPlayerRange"].(int16)
	m.SpawnCount = x["SpawnCount"].(int16)
	m.SpawnRange = x["SpawnRange"].(int16)

	if spawnData, has := x["SpawnData"].(map[string]any); has {
		new := fields.SpawnData{}
		new.FromNBT(spawnData)
		m.SpawnData = protocol.Optional[fields.SpawnData]{Set: true, Val: new}
	}
	if spawnPotentials, has := x["SpawnPotentials"].([]any); has {
		new := fields.MultiSpawnData{}
		new.FromNBT(spawnPotentials)
		m.SpawnPotentials = protocol.Optional[fields.MultiSpawnData]{Set: true, Val: new}
	}
}
