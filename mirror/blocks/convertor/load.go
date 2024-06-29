package convertor

import (
	"fmt"
	"phoenixbuilder/mirror/blocks/describe"
)

func (c *ToNEMCConvertor) LoadConvertRecord(r *ConvertRecord, overwrite bool, strict bool) {
	if val, ok := r.GetLegacyValue(); ok {
		if exist, err := c.AddAnchorByLegacyValue(describe.BlockNameForSearch(r.Name), val, r.RTID, overwrite); err != nil || exist {
			if strict {
				panic(fmt.Errorf("fail to add translation: %v %v %v", r.Name, val, r.RTID))
			}
		}
	} else {
		props, err := describe.PropsForSearchFromStr(r.SNBTStateOrValue)
		if err != nil {
			// continue
			panic(err)
		}
		if exist, err := c.AddAnchorByState(describe.BlockNameForSearch(r.Name), props, uint32(r.RTID), overwrite); err != nil || exist {
			if strict {
				panic(fmt.Errorf("fail to add translation: %v %v %v", r.Name, props.InPreciseSNBT(), r.RTID))
			}
		}
	}
}

func (c *ToNEMCConvertor) LoadTargetBlock(block *describe.Block) {
	if exist, err := c.AddAnchorByLegacyValue(block.NameForSearch(), block.LegacyValue(), block.Rtid(), false); err != nil {
		panic(err)
	} else if exist {
		panic("should not happen")
	}
	if exist, err := c.AddAnchorByState(block.NameForSearch(), block.StatesForSearch(), block.Rtid(), false); err != nil {
		panic(err)
	} else if exist {
		panic("should not happen")
	}
}
