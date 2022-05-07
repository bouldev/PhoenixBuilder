package mirror

import (
	"phoenixbuilder/mirror/chunk"
	"phoenixbuilder/mirror/define"
	"time"
)

// 为和国际版MC保持统一，世界范围被定义为 -64~319,
// 接受网易版数据包时(NEMCNetwork Decode) 会将 0~256 扩张到 -64~319
var WorldRange = define.Range{-64, 319}
var TimeStampNotFound = time.Unix(0, 0).Unix()

// ChunkData 包含一个区块的方块数据，Nbt信息，
// 收到/保存/读取该区块时区块所在的位置 ChunkX/ChunkZ (ChunkX=X>>4)
// 以及区块收到/保存的时间 (Unix Second)
type ChunkData struct {
	Chunk     *chunk.Chunk
	BlockNbts []map[string]interface{}
	TimeStamp int64
	ChunkPos  define.ChunkPos
}

func (cd *ChunkData) GetTime() time.Time {
	return time.Unix(cd.TimeStamp, 0)
}

func (cd *ChunkData) SetTime(t time.Time) {
	cd.TimeStamp = t.Unix()
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
	SetOffset(offset define.Pos)
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
	GetWithDeadline(pos define.ChunkPos, deadline time.Time) (data *ChunkData)
}

// 可以读写区块
type ChunkProvider interface {
	ChunkReader
	ChunkWriter
}

// 可以读写区块，但是容量有限制
// 因此每次写之后都必须检查因超过限制被 Drop 的部分
// 例如，区块容量 4096 块区块, 当达到该限制时，Drop 出 2048 区块以释放空间
// 若每次Wirte后未检查容量限制，应该在容量不足时 panic
type ChunkCacher interface {
	ChunkProvider
	GetDroppedChunks() []*ChunkData
}

type WorldDumper interface {
	DumpAll() chan RidBlockWithNbt
}

type WorldFeeder interface {
	Add(block RidBlockWithNbt) error
}

// ChunkCacher 和 ChunkProvider 构成 MirrorWorld 的存储体系
// offset 描述 Offset Pos (Outside Pos-Inside Pos)
type MirrorWorld interface {
	SetChunkCacher(cacher ChunkCacher)
	SetChunkProvider(provier ChunkProvider)
	SetOffSet(offset define.Pos)
	GetDumper() WorldDumper
	GetFeeder() WorldFeeder
}
