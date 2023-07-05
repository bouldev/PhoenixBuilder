package chunk

// SubChunk is a cube of blocks located in a chunk. It has a size of 16x16x16 blocks and forms part of a stack
// that forms a Chunk.
type SubChunk struct {
	air      uint32
	Storages []*PalettedStorage
}

// NewSubChunk creates a new sub chunk. All sub chunks should be created through this function
func NewSubChunk(air uint32) *SubChunk {
	return &SubChunk{air: air}
}

// Empty checks if the SubChunk is considered empty. This is the case if the SubChunk has 0 block storages or if it has
// a single one that is completely filled with air.
func (sub *SubChunk) Empty() bool {
	return len(sub.Storages) == 0 || (len(sub.Storages) == 1 && len(sub.Storages[0].palette.Values) == 1 && sub.Storages[0].palette.Values[0] == sub.air)
}

func (sub *SubChunk) Invalid() bool {
	return sub.Storages == nil
}

func (sub *SubChunk) Validate() {
	sub.Storages = []*PalettedStorage{}
}

// Layer returns a certain block storage/layer from a sub chunk. If no storage at the layer exists, the layer
// is created, as well as all layers between the current highest layer and the new highest layer.
func (sub *SubChunk) Layer(layer uint8) *PalettedStorage {
	for uint8(len(sub.Storages)) <= layer {
		// Keep appending to storages until the requested layer is achieved. Makes working with new layers
		// much easier.
		sub.Storages = append(sub.Storages, emptyStorage(sub.air))
	}
	return sub.Storages[layer]
}

// Layers returns all layers in the sub chunk. This method may also return an empty slice.
func (sub *SubChunk) Layers() []*PalettedStorage {
	return sub.Storages
}

// Block returns the runtime ID of the block located at the given X, Y and Z. X, Y and Z must be in a
// range of 0-15.
func (sub *SubChunk) Block(x, y, z byte, layer uint8) uint32 {
	if uint8(len(sub.Storages)) <= layer {
		return sub.air
	}
	return sub.Storages[layer].At(x, y, z)
}

// SetBlock sets the given block runtime ID at the given X, Y and Z. X, Y and Z must be in a range of 0-15.
func (sub *SubChunk) SetBlock(x, y, z byte, layer uint8, block uint32) {
	sub.Layer(layer).Set(x, y, z, block)
}

// Compact cleans the garbage from all block storages that sub chunk contains, so that they may be
// cleanly written to a database.
// func (sub *SubChunk) compact() {
// 	newStorages := make([]*PalettedStorage, 0, len(sub.storages))
// 	for _, storage := range sub.storages {
// 		storage.compact()
// 		if len(storage.palette.values) == 1 && storage.palette.values[0] == sub.air {
// 			// If the palette has only air in it, it means the storage is empty, so we can ignore it.
// 			continue
// 		}
// 		newStorages = append(newStorages, storage)
// 	}
// 	sub.storages = newStorages
// }
