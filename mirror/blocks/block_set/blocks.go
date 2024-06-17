package block_set

import "neo-omega-kernel/neomega/blocks/describe"

// describe blocks of a specific game version
// rtids must start from 0 and consecutive
type BlockSet struct {
	unknownRuntimeID uint32
	airRuntimeID     uint32
	version          uint32
	blocks           []*describe.Block
}

func (bs *BlockSet) Version() uint32 {
	return bs.version
}

func (bs *BlockSet) UnknownRitd() uint32 {
	return bs.unknownRuntimeID
}

func (bs *BlockSet) AirRuntimeID() uint32 {
	return bs.airRuntimeID
}

func (bs *BlockSet) BlockByRtid(rtid uint32) *describe.Block {
	if int(rtid) >= len(bs.blocks) {
		return nil
	}
	return bs.blocks[int(rtid)]
}

func NewBlockSet(unknownRuntimeID, airRuntimeID, version uint32) *BlockSet {
	return &BlockSet{
		unknownRuntimeID: unknownRuntimeID,
		airRuntimeID:     airRuntimeID,
		version:          version,
		blocks:           []*describe.Block{},
	}
}

func (bs *BlockSet) AddBlock(b *describe.Block) {
	if int(b.Rtid()) != len(bs.blocks) {
		panic("rtid mismatch")
	}
	bs.blocks = append(bs.blocks, b)
}
