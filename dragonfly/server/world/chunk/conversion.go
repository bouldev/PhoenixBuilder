package chunk

// legacyBlockEntry represents a block entry used in versions prior to 1.13.
type legacyBlockEntry struct {
	Name string `nbt:"name,omitempty"`
	Meta int16  `nbt:"val,omitempty"`
}

var (
	// legacyMappings allows simple conversion from a legacy block entry to a new one.
	legacyMappings = make(map[legacyBlockEntry]blockEntry)
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

// upgradeLegacyEntry upgrades a legacy block entry to a new one.
func upgradeLegacyEntry(name string, meta int16) (blockEntry, bool) {
	entry, ok := legacyMappings[legacyBlockEntry{Name: name, Meta: meta}]
	if !ok {
		// Also try cases where the meta should be disregarded.
		entry, ok = legacyMappings[legacyBlockEntry{Name: name}]
	}
	return entry, ok
}

// init creates conversions for each legacy and alias entry.
func init() {
}
