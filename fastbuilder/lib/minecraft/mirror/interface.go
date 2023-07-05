package mirror

import (
	"phoenixbuilder/fastbuilder/lib/minecraft/mirror/chunk"
	"phoenixbuilder/fastbuilder/lib/minecraft/mirror/define"
	"time"
)

var TimeStampNotFound = time.Unix(0, 0).Unix()

// ChunkData 包含一个区块的方块数据，Nbt信息，
// 收到/保存/读取该区块时区块所在的位置 ChunkX/ChunkZ (ChunkX=X>>4)
// 以及区块收到/保存的时间 (Unix Second)
type ChunkData struct {
	Chunk     *chunk.Chunk
	BlockNbts map[define.CubePos]map[string]interface{}
	SyncTime  int64
	ChunkPos  define.ChunkPos
}

func (cd *ChunkData) GetSyncTime() time.Time {
	return time.Unix(cd.SyncTime, 0)
}

func (cd *ChunkData) SetSyncTime(t time.Time) {
	cd.SyncTime = t.Unix()
}

type RidBlockWithNbt struct {
	Rid uint32
	Nbt map[string]interface{}
}

type LegacyBlockWithNbt struct {
	block *chunk.LegacyBlock
	Nbt   map[string]interface{}
}

// 考虑到 Chunk 是一个结构化的，空间受限的，16对齐的数据结构
// 因此，不同格式(特别是非序列化的格式中)，不同世界的转换很不方便
// WorldChunkBasic 通过提供一个 Offset Pos (Outside Pos-Inside Pos)
// 允许序列化的 Outside Blocks 与 Chunk 结构的 Inside Blocks 转换
type WorldChunkBasic interface {
	SetOffset(offset define.CubePos)
	DumpAll() chan RidBlockWithNbt
}

// WorldChunkAdvanced 和 WorldChunkBasic 概念类似
// 只是增加了可被导入，导出的数据类型
type WorldChunkAdvanced interface {
	WorldChunkBasic
	DumpAllAsLegacyBlock() chan LegacyBlockWithNbt
}

type ChunkWriter interface {
	Write(data *ChunkData) error
}

// 没有该数据时应该返回 nil
// GetWithDeadline(pos ChunkPos, deadline time.Time) 若在 deadline 前无法获得数据，那么应该返回 nil
type ChunkReader interface {
	Get(ChunkPos define.ChunkPos) (data *ChunkData)
	// GetWithNoFallBack(ChunkPos define.ChunkPos) (data *ChunkData)
}

// ChunkRequester 在指定deadline时间之前获得目标区块
type ChunkRequester interface {
	GetWithDeadline(pos define.ChunkPos, deadline time.Time) (data *ChunkData)
}

// 可以读写区块
type ChunkProvider interface {
	ChunkReader
	ChunkWriter
}

// type WorldDumper interface {
// 	DumpAll() chan RidBlockWithNbt
// }

// type WorldFeeder interface {
// 	Add(block RidBlockWithNbt) error
// }

// // ChunkCacher 和 ChunkProvider 构成 MirrorWorld 的存储体系
// // offset 描述 Offset Pos (Outside Pos-Inside Pos)
// type MirrorWorld interface {
// 	SetChunkRequester(requester ChunkRequester)
// 	SetChunkProvider(provier ChunkProvider)
// 	SetOffSet(offset define.Pos)
// 	GetDumper() WorldDumper
// 	GetFeeder() WorldFeeder
// }
