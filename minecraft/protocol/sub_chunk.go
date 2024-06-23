package protocol

import "math"

const (
	HeightMapDataNone = iota
	HeightMapDataHasData
	HeightMapDataTooHigh
	HeightMapDataTooLow
)

const (
	SubChunkRequestModeLimitless = math.MaxUint32 - iota
	SubChunkRequestModeLimited
)

const (
	SubChunkResultSuccess = iota + 1
	SubChunkResultChunkNotFound
	SubChunkResultInvalidDimension
	SubChunkResultPlayerNotFound
	SubChunkResultIndexOutOfBounds
	SubChunkResultSuccessAllAir
)

// SubChunkEntry contains the data of a sub-chunk entry relative to a center sub chunk position, used for the sub-chunk
// requesting system introduced in v1.18.10.
type SubChunkEntry struct {
	// Offset contains the offset between the sub-chunk position and the center position.
	Offset SubChunkOffset
	// Result is always one of the constants defined in the SubChunkResult constants.
	Result byte
	// RawPayload contains the serialized sub-chunk data.
	RawPayload []byte
	// HeightMapType is always one of the constants defined in the HeightMapData constants.
	HeightMapType byte

	/*
		PhoenixBuilder specific changes.
		Changes Maker: Liliya233
		Committed by Happy2018new.

		HeightMapData is the data for the height map.

		For netease, the data type of this field is []uint8,
		but on standard minecraft, this is []int8.
	*/
	HeightMapData []uint8
	// HeightMapData []int8

	// BlobHash is the hash of the blob.
	BlobHash uint64
}

// Marshal encodes/decodes a SubChunkEntry assuming the blob cache is enabled.
func (x *SubChunkEntry) Marshal(r IO) {
	Single(r, &x.Offset)
	r.Uint8(&x.Result)
	if x.Result != SubChunkResultSuccessAllAir {
		r.ByteSlice(&x.RawPayload)
	}
	r.Uint8(&x.HeightMapType)
	if x.HeightMapType == HeightMapDataHasData {
		// PhoenixBuilder specific changes.
		// Changes Maker: Liliya233
		// Committed by Happy2018new.
		{
			FuncSliceOfLen(r, 256, &x.HeightMapData, r.Uint8)
			// FuncSliceOfLen(r, 256, &x.HeightMapData, r.Int8)
		}
	}
	r.Uint64(&x.BlobHash)
}

// SubChunkEntryNoCache encodes/decodes a SubChunkEntry assuming the blob cache is not enabled.
func SubChunkEntryNoCache(r IO, x *SubChunkEntry) {
	Single(r, &x.Offset)
	r.Uint8(&x.Result)
	r.ByteSlice(&x.RawPayload)
	r.Uint8(&x.HeightMapType)
	if x.HeightMapType == HeightMapDataHasData {
		// PhoenixBuilder specific changes.
		// Changes Maker: Liliya233
		// Committed by Happy2018new.
		{
			FuncSliceOfLen(r, 256, &x.HeightMapData, r.Uint8)
			// FuncSliceOfLen(r, 256, &x.HeightMapData, r.Int8)
		}
	}
}

/*
PhoenixBuilder specific changes.
Changes Maker: Liliya233
Committed by Happy2018new.

SubChunkOffset represents an offset from the base position of another sub chunk.

For netease, the data type of this field is [3]uint8,
but on standard minecraft, this is [3]int8.
*/
type SubChunkOffset [3]uint8 // [3]int8

// Marshal encodes/decodes a SubChunkOffset.
func (x *SubChunkOffset) Marshal(r IO) {
	// PhoenixBuilder specific changes.
	// Changes Maker: Liliya233
	// Committed by Happy2018new.
	{
		r.Uint8(&x[0])
		r.Uint8(&x[1])
		r.Uint8(&x[2])
		/*
			r.Int8(&x[0])
			r.Int8(&x[1])
			r.Int8(&x[2])
		*/
	}
}
