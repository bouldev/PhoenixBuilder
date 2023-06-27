package chunk

import (
	"bytes"
	"sync"
)

var pool = sync.Pool{
	New: func() any {
		return bytes.NewBuffer(make([]byte, 0, 1024))
	},
}

// Encode encodes Chunk to an intermediate representation SerialisedData. An Encoding may be passed to encode either for
// network or disk purposed, the most notable difference being that the network encoding generally uses varints and no
// NBT.
// 2401PT: Biomes are removed
func Encode(c *Chunk, e Encoding) SerialisedData {
	buf := pool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		pool.Put(buf)
	}()
	return encodeSubChunks(buf, c, e)
}

// encodeSubChunks encodes the sub chunks of the Chunk passed into the bytes.Buffer buf. It uses the encoding passed to
// encode the block storages and returns the resulting SerialisedData.
func encodeSubChunks(buf *bytes.Buffer, c *Chunk, e Encoding) (d SerialisedData) {
	d.SubChunks = make([][]byte, len(c.sub))
	for i, sub := range c.sub {
		_, _ = buf.Write([]byte{SubChunkVersion, byte(len(sub.Storages)), uint8(i + (c.r[0] >> 4))})
		for _, storage := range sub.Storages {
			encodePalettedStorage(buf, storage, e, BlockPaletteEncoding)
		}
		d.SubChunks[i] = make([]byte, buf.Len())
		_, _ = buf.Read(d.SubChunks[i])
	}
	return
}

// encodePalettedStorage encodes a PalettedStorage into a bytes.Buffer. The Encoding passed is used to write the Palette
// of the PalettedStorage.
func encodePalettedStorage(buf *bytes.Buffer, storage *PalettedStorage, e Encoding, pe paletteEncoding) {
	b := make([]byte, len(storage.indices)*4+1)
	b[0] = byte(storage.bitsPerIndex<<1) | e.network()

	for i, v := range storage.indices {
		// Explicitly don't use the binary package to greatly improve performance of writing the Uint32s.
		b[i*4+1], b[i*4+2], b[i*4+3], b[i*4+4] = byte(v), byte(v>>8), byte(v>>16), byte(v>>24)
	}
	_, _ = buf.Write(b)

	e.encodePalette(buf, storage.palette, pe)
}
