package chunk

import (
	"bytes"
	"fmt"
	"phoenixbuilder/fastbuilder/lib/minecraft/mirror/define"
	"phoenixbuilder/minecraft/nbt"

	"github.com/pterm/pterm"
)

// func NEMCNetworkDecode(data []byte, count int) (c *Chunk, nbtBlocks []map[string]interface{}, err error) {
// 	air, ok := StateToRuntimeID("minecraft:air", nil)
// 	if !ok {
// 		panic("cannot find air runtime ID")
// 	}
// 	var (
// 		buf     = bytes.NewBuffer(data)
// 		encoder = &nemcNetworkEncoding{}
// 	)
// 	c = New(air, define.Range{-64, 319})
// 	nbtBlocks = []map[string]interface{}{}
// 	defer func() {
// 		r := recover()
// 		if r != nil {
// 			c = nil
// 			err = fmt.Errorf("%v", r)
// 		}
// 	}()
// 	encoder.isChunkDecoding = true
// 	// offset := uint8((int(define.NEMCWorldStart) - define.WorldRange[0]) / 16)
// 	// fmt.Println(offset)
// 	for i := 0; i < count; i++ {
// 		index := uint8(i)
// 		// decodeSubChunk(buf, c, &index, NetworkEncoding)
// 		c.sub[index], err = decodeSubChunk(buf, c, &index, encoder)
// 		if err != nil {
// 			return nil, nil, err
// 		}
// 	}
// 	encoder.isChunkDecoding = false
// 	subChunkCount := 25
// 	for i := 0; i < subChunkCount; i++ {
// 		_, err := decodePalettedStorage(buf, encoder, biomePaletteEncoding{})
// 		if err != nil {
// 			fmt.Println(err)
// 			return nil, nil, err
// 		}
// 	}
// 	return c, nbtBlocks, nil
// }

func NEMCSubChunkDecode(data []byte) (int8, *SubChunk, []map[string]interface{}, error) {
	var (
		buf     = bytes.NewBuffer(data)
		encoder = &nemcNetworkEncoding{}
	)
	encoder.isChunkDecoding = true
	subChunkIndex, subChunk, err := decodeSubChunkNEMC(buf, encoder, AirRID)
	// palette := subChunk.storages[0].Palette()
	// for i := 0; i < palette.Len(); i++ {
	// 	rid := palette.Value(uint16(i))
	// 	b, _ := RuntimeIDToBlock(rid)
	// 	fmt.Printf("%v\t", b.Name)
	// }

	nbts := make([]map[string]interface{}, 0)
	if buf.Len() > 0 {
		nbtDocoder := nbt.NewDecoder(buf)
		blockData := make(map[string]interface{})
		for buf.Len() != 0 {
			if err := nbtDocoder.Decode(&blockData); err != nil {
				pterm.Printfln("decode chunk nbt error %v", err)
				break
			}
			//fmt.Println(blockData)
			nbts = append(nbts, blockData)
		}
	}
	return subChunkIndex, subChunk, nbts, err
}

// DiskDecode decodes the data from a SerialisedData object into a chunk and returns it. If the data was
// invalid, an error is returned.
func DiskDecode(data SerialisedData, r define.Range) (*Chunk, error) {
	c := New(AirRID, r)

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
		sub.Storages = append(sub.Storages, storage)
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
		sub.Storages = make([]*PalettedStorage, storageCount)

		for i := byte(0); i < storageCount; i++ {
			sub.Storages[i], err = decodePalettedStorage(buf, e, BlockPaletteEncoding)
			if err != nil {
				return nil, err
			}

		}
		//if storageCount > 1 {
		//	for x := uint8(0); x < 16; x++ {
		//		for y := uint8(0); y < 16; y++ {
		//			for z := uint8(0); z < 16; z++ {
		//				for i := 0; i < int(storageCount); i++ {
		//					rtid := sub.storages[i].At(x, y, z)
		//					if rtid != AirRID {
		//						if i != 0 {
		//							fmt.Printf("%v, %v\n", i, rtid)
		//						}
		//					}
		//				}
		//			}
		//		}
		//	}
		//}

	}
	return sub, nil
}

func decodeSubChunkNEMC(buf *bytes.Buffer, e Encoding, airRID uint32) (int8, *SubChunk, error) {
	ver, err := buf.ReadByte()
	Index := int8(127)
	if err != nil {
		return Index, nil, fmt.Errorf("error reading version: %w", err)
	}
	sub := NewSubChunk(airRID)
	switch ver {
	default:
		return Index, nil, fmt.Errorf("unknown sub chunk version %v: can't decode", ver)
	case 1:
		// Version 1 only has one layer for each sub chunk, but uses the format with palettes.
		storage, err := decodePalettedStorage(buf, e, BlockPaletteEncoding)
		if err != nil {
			return Index, nil, err
		}
		sub.Storages = append(sub.Storages, storage)
	case 8, 9:
		// Version 8 allows up to 256 layers for one sub chunk.
		storageCount, err := buf.ReadByte()
		if err != nil {
			return Index, nil, fmt.Errorf("error reading storage count: %w", err)
		}
		if ver == 9 {
			uIndex, err := buf.ReadByte()
			Index = int8(uIndex)
			if err != nil {
				return Index, nil, fmt.Errorf("error reading subchunk index: %w", err)
			}
			// The index as written here isn't the actual index of the subchunk within the chunk. Rather, it is the Y
			// value of the subchunk. This means that we need to translate it to an index.
		}
		sub.Storages = make([]*PalettedStorage, storageCount)

		for i := byte(0); i < storageCount; i++ {
			sub.Storages[i], err = decodePalettedStorage(buf, e, BlockPaletteEncoding)
			if err != nil {
				return Index, nil, err
			}
		}
	}
	return Index, sub, nil
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

	size := PaletteSize(blockSize)
	uint32Count := size.Uint32s()

	uint32s := make([]uint32, uint32Count)
	byteCount := uint32Count * 4

	data := buf.Next(byteCount)
	if len(data) != byteCount {
		return nil, fmt.Errorf("cannot read paletted storage (size=%v) %T: not enough block data present: expected %v bytes, got %v", blockSize, pe, byteCount, len(data))
	}
	for i := 0; i < uint32Count; i++ {
		// Explicitly don't use the binary package to greatly improve performance of reading the Uint32s.
		uint32s[i] = uint32(data[i*4]) | uint32(data[i*4+1])<<8 | uint32(data[i*4+2])<<16 | uint32(data[i*4+3])<<24
	}
	p, err := e.decodePalette(buf, PaletteSize(blockSize), pe)
	return NewPalettedStorage(uint32s, p), err
}
