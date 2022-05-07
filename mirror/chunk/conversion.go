package chunk

import (
	_ "embed"

	"phoenixbuilder/minecraft/nbt"
)

// legacyBlockEntry represents a block entry used in versions prior to 1.13.
type legacyBlockEntry struct {
	Name string `nbt:"name,omitempty"`
	Meta int16  `nbt:"val,omitempty"`
}

var (
	//go:embed block_aliases.nbt
	blockAliasesData []byte
	// aliasMappings maps from a legacy block name alias to an updated name.
	aliasMappings = make(map[string]string)
)

// upgradeAliasEntry upgrades a possible alias block entry to the correct/updated block entry.
func upgradeAliasEntry(entry blockEntry) (blockEntry, bool) {
	if alias, ok := aliasMappings[entry.Name]; ok {
		entry.Name = alias
		return entry, true
	}
	return blockEntry{}, false
}

// init creates conversions for each legacy and alias entry.
func init() {
	if err := nbt.Unmarshal(blockAliasesData, &aliasMappings); err != nil {
		panic(err)
	}
}
