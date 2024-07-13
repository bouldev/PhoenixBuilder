/*
PhoenixBuilder specific fields.
Author: Happy2018new
*/
package protocol

// PistonAttachedBlocks reads some attached blocks as a int32 slice from the underlying buffer.
func (r *Reader) PistonAttachedBlocks(m *[]int32) {
	var blocks []BlockPos
	var nbtBlocks []int32

	FuncSliceVarint16Length(r, &blocks, r.BlockPos)

	nbtBlocks = make([]int32, len(blocks)*3)
	for key, value := range blocks {
		nbtBlocks[key*3], nbtBlocks[key*3+1], nbtBlocks[key*3+2] = value[0], value[1], value[2]
	}
}

// PistonAttachedBlocks writes some attached blocks as a int32 slice to the underlying buffer.
func (w *Writer) PistonAttachedBlocks(m *[]int32) {
	var blocks []BlockPos
	var length int16

	if m == nil || len(*m) == 0 {
		w.Varint16(&length)
		return
	}

	length = int16(len(*m) / 3)
	blocks = make([]BlockPos, length)

	for i := 0; i < int(length); i++ {
		blocks[i] = BlockPos{(*m)[i*3], (*m)[i*3+1], (*m)[i*3+2]}
	}

	FuncSliceVarint16Length(w, &blocks, w.BlockPos)
}
