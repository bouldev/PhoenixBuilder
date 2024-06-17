package describe

import (
	"fmt"
	"strings"
)

type Block struct {
	rtid            uint32
	name            string
	nameForSearch   BaseWithNameSpace
	statsSnbt       string
	legacyValue     uint16
	states          Props
	statesForSearch *PropsForSearch
}

func (b *Block) String() string {
	if b == nil {
		return "[empty]"
	}
	return fmt.Sprintf("%v %v (Value: %v) (RuntimeID: %v)", b.name, b.states.SNBTString(), b.legacyValue, b.rtid)
}

func (b *Block) BedrockString() string {
	return b.nameForSearch.BaseName() + " " + b.states.BedrockString(true)
}

func (b *Block) Rtid() uint32 {
	return b.rtid
}

func (b *Block) ShortName() string {
	return b.nameForSearch.BaseName()
}

func (b *Block) LongName() string {
	return b.nameForSearch.LongName()
}

// func (b *Block) Name() BaseWithNameSpace {
// 	return b.nameForSearch
// }

func (b *Block) NameForSearch() BaseWithNameSpace {
	return b.nameForSearch
}

func (b *Block) States() Props {
	return b.states
}

func (b *Block) StatesForSearch() *PropsForSearch {
	return b.statesForSearch
}

func (b *Block) LegacyValue() uint16 {
	return b.legacyValue
}

func NewBlockFromSnbt(blockName, statesSnbt string, value uint16, rtid uint32) *Block {
	blockName = strings.TrimSpace(blockName)
	blockName = strings.TrimPrefix(blockName, "minecraft:")
	blockName = strings.TrimSpace(blockName)
	statesSnbt = strings.TrimSpace(statesSnbt)
	statesSnbt = strings.ReplaceAll(statesSnbt, "minecraft:", "")
	statesSnbt = strings.TrimSpace(statesSnbt)
	propsForSearch, err := PropsForSearchFromStr(statesSnbt)
	if err != nil {
		panic(err)
	}
	return &Block{
		rtid:            rtid,
		name:            blockName,
		nameForSearch:   BlockNameForSearch(blockName),
		statsSnbt:       statesSnbt,
		legacyValue:     value,
		states:          PropsFromSNBT(statesSnbt),
		statesForSearch: propsForSearch,
	}
}
