package chunk

import (
	"bytes"
	"fmt"
	"phoenixbuilder/mirror/define"

	"phoenixbuilder/minecraft/nbt"
)

func NEMCNetworkDecode(data []byte, count int) (*Chunk, []map[string]interface{}, error) {
	air, ok := StateToRuntimeID("minecraft:air", nil)
	if !ok {
		panic("cannot find air runtime ID")
	}
	var (
		c       = New(air, define.Range{-64, 319})
		buf     = bytes.NewBuffer(data)
		err     error
		encoder = &nemcNetworkEncoding{}
	)
	encoder.isChunkDecoding = true
	for i := 0; i < count; i++ {
		index := uint8(i)
		// decodeSubChunk(buf, c, &index, NetworkEncoding)
		c.sub[index+4], err = decodeSubChunk(buf, c, &index, encoder)
		if err != nil {
			return nil, nil, err
		}
	}
	encoder.isChunkDecoding = false
	fakeBiomes := make([]byte, 256)
	buf.Read(fakeBiomes)
	_, _ = buf.ReadByte()

	// it seems netease add something after biomes info,
	// e.g. [13 45 77 109 141 173 205 237]
	//      [5 13 37 45 69 77 85 109 141 173 205 237]
	// the following 14 lines try to get rid of it, but i don't know what
	// is missed
	for _, b := range buf.Bytes() {
		// Nbt should start with a Nbt TAG_Compound
		if b != uint8(10) {
			buf.ReadByte()
		} else {
			dec := nbt.NewDecoder(bytes.NewBuffer(buf.Bytes()))
			var m map[string]interface{}
			if err := dec.Decode(&m); err != nil {
				buf.ReadByte()
			} else {
				break
			}
		}
	}

	dec := nbt.NewDecoder(buf)
	nbtBlocks := []map[string]interface{}{}
	for buf.Len() != 0 {
		var m map[string]interface{}
		if err := dec.Decode(&m); err != nil {
			// the rest of buf is also effect, so we stop decoding and return immediately
			fmt.Printf("error decoding block entity: %v\n", err)
			return c, nbtBlocks, nil
			// return nil, fmt.Errorf("error decoding block entity: %w", err)
		}
		// c.SetBlockNBT(cube.Pos{int(m["x"].(int32)), int(m["y"].(int32)), int(m["z"].(int32))}, m)
		//id:Bed isMovable:1 x:81 y:64 z:163
		nbtBlocks = append(nbtBlocks, m)
	}
	return c, nbtBlocks, nil
}

// DiskDecode decodes the data from a SerialisedData object into a chunk and returns it. If the data was
// invalid, an error is returned.
func DiskDecode(data SerialisedData, r define.Range) (*Chunk, error) {
	air, ok := StateToRuntimeID("minecraft:air", nil)
	if !ok {
		panic("cannot find air runtime ID")
	}

	c := New(air, r)

	var err error
	for i, sub := range data.SubChunks {
		if len(sub) == 0 {
			// No data for this sub chunk.
			continue
		}
		index := uint8(i)
		if c.sub[index], err = decodeSubChunk(bytes.NewBuffer(sub), c, &index, DiskEncoding); err != nil {
			return nil, err
		}
	}
	return c, nil
}

// decodeSubChunk decodes a SubChunk from a bytes.Buffer. The Encoding passed defines how the block storages of the
// SubChunk are decoded.
func decodeSubChunk(buf *bytes.Buffer, c *Chunk, index *byte, e Encoding) (*SubChunk, error) {
	ver, err := buf.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("error reading version: %w", err)
	}
	sub := NewSubChunk(c.air)
	switch ver {
	default:
		return nil, fmt.Errorf("unknown sub chunk version %v: can't decode", ver)
	case 1:
		// Version 1 only has one layer for each sub chunk, but uses the format with palettes.
		storage, err := decodePalettedStorage(buf, e, BlockPaletteEncoding)
		if err != nil {
			return nil, err
		}
		sub.storages = append(sub.storages, storage)
	case 8, 9:
		// Version 8 allows up to 256 layers for one sub chunk.
		storageCount, err := buf.ReadByte()
		if err != nil {
			return nil, fmt.Errorf("error reading storage count: %w", err)
		}
		if ver == 9 {
			uIndex, err := buf.ReadByte()
			if err != nil {
				return nil, fmt.Errorf("error reading subchunk index: %w", err)
			}
			// The index as written here isn't the actual index of the subchunk within the chunk. Rather, it is the Y
			// value of the subchunk. This means that we need to translate it to an index.
			*index = uint8(int8(uIndex) - int8(c.r[0]>>4))
		}
		sub.storages = make([]*PalettedStorage, storageCount)

		for i := byte(0); i < storageCount; i++ {
			sub.storages[i], err = decodePalettedStorage(buf, e, BlockPaletteEncoding)
			if err != nil {
				return nil, err
			}
		}
	}
	return sub, nil
}

// decodePalettedStorage decodes a PalettedStorage from a bytes.Buffer. The Encoding passed is used to read either a
// network or disk block storage.
func decodePalettedStorage(buf *bytes.Buffer, e Encoding, pe paletteEncoding) (*PalettedStorage, error) {
	blockSize, err := buf.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("error reading block size: %w", err)
	}
	blockSize >>= 1
	if blockSize == 0x7f {
		return nil, nil
	}

	size := paletteSize(blockSize)
	uint32Count := size.uint32s()

	uint32s := make([]uint32, uint32Count)
	byteCount := uint32Count * 4

	data := buf.Next(byteCount)
	if len(data) != byteCount {
		return nil, fmt.Errorf("cannot read paletted storage (size=%v) %T: not enough block data present: expected %v bytes, got %v", blockSize, pe, byteCount, len(data))
	}
	for i := 0; i < uint32Count; i++ {
		// Explicitly don't use the binary package to greatly improve performance of reading the uint32s.
		uint32s[i] = uint32(data[i*4]) | uint32(data[i*4+1])<<8 | uint32(data[i*4+2])<<16 | uint32(data[i*4+3])<<24
	}
	p, err := e.decodePalette(buf, paletteSize(blockSize), pe)
	return newPalettedStorage(uint32s, p), err
}
