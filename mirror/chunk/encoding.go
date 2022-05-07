package chunk

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"phoenixbuilder/minecraft/nbt"

	"phoenixbuilder/minecraft/protocol"
)

type (
	// Encoding is an encoding type used for Chunk encoding. Implementations of this interface are DiskEncoding and
	// NetworkEncoding, which can be used to encode a Chunk to an intermediate disk or network representation respectively.
	Encoding interface {
		encodePalette(buf *bytes.Buffer, p *Palette, e paletteEncoding)
		decodePalette(buf *bytes.Buffer, blockSize paletteSize, e paletteEncoding) (*Palette, error)
		network() byte
	}
	// paletteEncoding is an encoding type used for Chunk encoding. It is used to encode different types of palettes
	// (for example, blocks or biomes) differently.
	paletteEncoding interface {
		encode(buf *bytes.Buffer, v uint32)
		decode(buf *bytes.Buffer) (uint32, error)
	}
)

type SerialisedData struct {
	// sub holds the data of the serialised sub chunks in a chunk. Sub chunks that are empty or that otherwise
	// don't exist are represented as an empty slice (or technically, nil).
	SubChunks [][]byte
	// BlockNBT is an encoded NBT array of all blocks that carry additional NBT, such as chests, with all
	// their contents.
	BlockNBT []byte
}

var (
	// DiskEncoding is the Encoding for writing a Chunk to disk. It writes block palettes using NBT and does not use
	// varints.
	DiskEncoding diskEncoding
	// BlockPaletteEncoding is the paletteEncoding used for encoding a palette of block states encoded as NBT.
	BlockPaletteEncoding blockPaletteEncoding
)

// blockPaletteEncoding implements the encoding of block palettes to disk.
type blockPaletteEncoding struct{}

func (blockPaletteEncoding) encode(buf *bytes.Buffer, v uint32) {
	// Get the block state registered with the runtime IDs we have in the palette of the block storage
	// as we need the name and data value to store.
	name, props, _ := RuntimeIDToState(v)
	_ = nbt.NewEncoderWithEncoding(buf, nbt.LittleEndian).Encode(blockEntry{Name: name, State: props, Version: CurrentBlockVersion})
}
func (blockPaletteEncoding) decode(buf *bytes.Buffer) (uint32, error) {
	var e blockEntry
	if err := nbt.NewDecoderWithEncoding(buf, nbt.LittleEndian).Decode(&e); err != nil {
		return 0, fmt.Errorf("error decoding block palette entry: %w", err)
	}
	// As of 1.18.30, many common block state names have been renamed for consistency and the old names are now aliases.
	// This function checks if the entry has an alias and if so, returns the updated entry.
	if updatedEntry, ok := upgradeAliasEntry(e); ok {
		e = updatedEntry
	}

	v, ok := StateToRuntimeID(e.Name, e.State)
	if !ok {
		return 0, fmt.Errorf("cannot get runtime ID of block state %v{%+v}", e.Name, e.State)
	}
	return v, nil
}

// diskEncoding implements the Chunk encoding for writing to disk.
type diskEncoding struct{}

func (diskEncoding) network() byte { return 0 }
func (diskEncoding) encodePalette(buf *bytes.Buffer, p *Palette, e paletteEncoding) {
	if p.size != 0 {
		_ = binary.Write(buf, binary.LittleEndian, uint32(p.Len()))
	}
	for _, v := range p.values {
		e.encode(buf, v)
	}
}
func (diskEncoding) decodePalette(buf *bytes.Buffer, blockSize paletteSize, e paletteEncoding) (*Palette, error) {
	paletteCount := uint32(1)
	if blockSize != 0 {
		if err := binary.Read(buf, binary.LittleEndian, &paletteCount); err != nil {
			return nil, fmt.Errorf("error reading palette entry count: %w", err)
		}
	}

	var err error
	palette := newPalette(blockSize, make([]uint32, paletteCount))
	for i := uint32(0); i < paletteCount; i++ {
		palette.values[i], err = e.decode(buf)
		if err != nil {
			return nil, err
		}
	}
	return palette, nil
}

// networkEncoding implements the Chunk encoding for sending over network.
type nemcNetworkEncoding struct {
	isChunkDecoding bool
}

func (*nemcNetworkEncoding) network() byte { return 1 }
func (*nemcNetworkEncoding) translate(nemcRID uint32) (mcRid uint32) {
	return NEMCRuntimeIDToStandardRuntimeID(nemcRID)
}
func (*nemcNetworkEncoding) encodePalette(buf *bytes.Buffer, p *Palette, _ paletteEncoding) {
	panic("nemcNetworkEncoding.encodePalette not implement")
}
func (o *nemcNetworkEncoding) decodePalette(buf *bytes.Buffer, blockSize paletteSize, _ paletteEncoding) (*Palette, error) {
	var paletteCount int32 = 1
	if blockSize != 0 {
		if err := protocol.Varint32(buf, &paletteCount); err != nil {
			return nil, fmt.Errorf("error reading palette entry count: %w", err)
		}
		if paletteCount <= 0 {
			return nil, fmt.Errorf("invalid palette entry count %v", paletteCount)
		}
	}

	blocks, temp := make([]uint32, paletteCount), int32(0)
	for i := int32(0); i < paletteCount; i++ {
		if err := protocol.Varint32(buf, &temp); err != nil {
			return nil, fmt.Errorf("error decoding palette entry: %w", err)
		}
		if o.isChunkDecoding {
			blocks[i] = o.translate(uint32(temp))
		} else {
			blocks[i] = uint32(temp)
		}

	}
	return &Palette{values: blocks, size: blockSize}, nil
}
