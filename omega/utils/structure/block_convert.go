package structure

import (
	"phoenixbuilder/mirror/chunk"
	"strings"
)

type DoubleValueLegacyBlockToRuntimeIDMapper struct {
	palatteIDToBlockNameMapping map[uint16]string
	quickCache                  map[uint32]uint32
}

func NewDoubleValueLegacyBlockToRuntimeIDMapper() *DoubleValueLegacyBlockToRuntimeIDMapper {
	return &DoubleValueLegacyBlockToRuntimeIDMapper{
		palatteIDToBlockNameMapping: map[uint16]string{},
		quickCache:                  map[uint32]uint32{},
	}

}

func (o *DoubleValueLegacyBlockToRuntimeIDMapper) AddBlockNamePalette(paletteID uint16, blockName string) {
	blockName = strings.ReplaceAll(blockName, "minecraft:", "")
	o.palatteIDToBlockNameMapping[paletteID] = blockName
}

func (o *DoubleValueLegacyBlockToRuntimeIDMapper) GetRTID(blockPaletteID uint16, data uint16) (rtid uint32) {
	quickCacheID := uint32(blockPaletteID)<<16 | uint32(data)
	if rtid, ok := o.quickCache[quickCacheID]; ok {
		return rtid
	}
	blockName, found := o.palatteIDToBlockNameMapping[blockPaletteID]
	if !found {
		o.quickCache[quickCacheID] = chunk.AirRID
		return chunk.AirRID
	}
	if rtid, found := chunk.LegacyBlockToRuntimeID(blockName, data); found {
		o.quickCache[quickCacheID] = rtid
		return rtid
	} else {
		o.quickCache[quickCacheID] = chunk.AirRID
		return chunk.AirRID
	}
}

type RuntimeIDConvertor struct {
	ConvertFN  func(uint32) uint32
	quickCache map[uint32]uint32
}

func NewRuntimeIDConvertor() *RuntimeIDConvertor {
	return &RuntimeIDConvertor{
		quickCache: map[uint32]uint32{},
	}
}

func (o *RuntimeIDConvertor) Convert(orig uint32) uint32 {
	if rtid, found := o.quickCache[orig]; found {
		return rtid
	} else {
		rtid = o.ConvertFN(orig)
		o.quickCache[orig] = rtid
		return rtid
	}
}
