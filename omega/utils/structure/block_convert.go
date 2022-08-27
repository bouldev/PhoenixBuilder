package structure

import (
	"phoenixbuilder/mirror/chunk"
	"strings"
)

type DoubleValueLegacyBlockToRuntimeIDMapper struct {
	palatteIDToBlockNameMapping map[uint16]string
	quickCache                  map[uint32]uint32
	// AuxHash                     map[chunk.LegacyBlockHash]uint32
	// AuxOffset                   int
}

func NewDoubleValueLegacyBlockToRuntimeIDMapper() *DoubleValueLegacyBlockToRuntimeIDMapper {
	return &DoubleValueLegacyBlockToRuntimeIDMapper{
		palatteIDToBlockNameMapping: map[uint16]string{},
		quickCache:                  map[uint32]uint32{},
		// AuxHash:                     map[chunk.LegacyBlockHash]uint32{},
		// AuxOffset:                   len(chunk.Blocks) + 1,
	}

}

func (o *DoubleValueLegacyBlockToRuntimeIDMapper) AddBlockNamePalette(paletteID uint16, blockName string) {
	blockName = strings.ReplaceAll(blockName, "minecraft:", "")
	o.palatteIDToBlockNameMapping[paletteID] = blockName
}

func (o *DoubleValueLegacyBlockToRuntimeIDMapper) GetRTID(blockPaletteID uint16, data uint16) (rtid uint32) {
	// pterm.Info.Println(blockPaletteID, data)
	quickCacheID := uint32(blockPaletteID)<<16 | uint32(data)
	if rtid, ok := o.quickCache[quickCacheID]; ok {
		// pterm.Info.Println(rtid)
		return rtid
	}
	blockName, found := o.palatteIDToBlockNameMapping[blockPaletteID]
	if !found {
		o.quickCache[quickCacheID] = chunk.AirRID
		// pterm.Info.Println("Air")
		return chunk.AirRID
	}
	if rtid, found := chunk.LegacyBlockToRuntimeID(blockName, data); found {
		// pterm.Info.Println("Put", rtid)
		o.quickCache[quickCacheID] = rtid
		return rtid
	} else {
		if rtid, found := chunk.LegacyBlockToRuntimeID(blockName, 0); found {
			o.quickCache[quickCacheID] = rtid
			return rtid
		}
		o.quickCache[quickCacheID] = chunk.AirRID
		// pterm.Info.Println("Air2")
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

type RuntimeIDToPaletteConvertor struct {
	quickCache       map[uint32]uint32
	Palette          []string
	paletteLookUp    map[string]uint32
	AcquirePaletteFN func(uint32) string
}

func NewRuntimeIDToPaletteConvertor() *RuntimeIDToPaletteConvertor {
	return &RuntimeIDToPaletteConvertor{
		quickCache:       map[uint32]uint32{},
		Palette:          make([]string, 0),
		paletteLookUp:    make(map[string]uint32),
		AcquirePaletteFN: nil,
	}
}

func (o *RuntimeIDToPaletteConvertor) Convert(rtid uint32) uint32 {
	if paletteI, found := o.quickCache[rtid]; found {
		return paletteI
	} else {
		name := o.AcquirePaletteFN(rtid)
		paletteI := uint32(0)
		if _paletteI, found := o.paletteLookUp[name]; found {
			paletteI = _paletteI
		} else {
			paletteI = uint32(len(o.Palette))
			o.Palette = append(o.Palette, name)
			o.paletteLookUp[name] = paletteI
		}
		o.quickCache[rtid] = paletteI
		return paletteI
	}
}
