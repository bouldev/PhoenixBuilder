package convertor

import (
	"phoenixbuilder/mirror/blocks/describe"
	"sync"
)

type ToNEMCConvertor struct {
	rtidUnknown uint32
	rtidAir     uint32
	baseNames   map[string]*ToNEMCBaseNames
	mu          sync.RWMutex
}

func (c *ToNEMCConvertor) ensureBaseNameGroup(name string) *ToNEMCBaseNames {
	c.mu.RLock()
	if to, found := c.baseNames[name]; found {
		c.mu.RUnlock()
		return to
	}
	c.mu.RUnlock()
	c.mu.Lock()
	defer c.mu.Unlock()
	to := &ToNEMCBaseNames{
		RtidUnknown:         c.rtidUnknown,
		RtidAir:             c.rtidAir,
		legacyValuesMapping: make([]uint32, 0),
		statesWithRtid: make([]struct {
			states *describe.PropsForSearch
			rtid   uint32
		}, 0),
		StatesWithRtidQuickMatch: make(map[string]uint32),
		mu:                       sync.RWMutex{},
	}
	c.baseNames[name] = to
	return to
}

func (c *ToNEMCConvertor) getBaseNameGroup(name string) (baseGroup *ToNEMCBaseNames, found bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	group, found := c.baseNames[name]
	return group, found
}

func (c *ToNEMCConvertor) AddAnchorByLegacyValue(name describe.BaseWithNameSpace, legacyValue uint16, nemcRTID uint32, overwrite bool) (exist bool, conflictErr error) {
	baseNameGroup := c.ensureBaseNameGroup(name.BaseName())
	return baseNameGroup.addAnchorByLegacyValue(legacyValue, nemcRTID, overwrite)
}

func (c *ToNEMCConvertor) PreciseMatchByLegacyValue(name describe.BaseWithNameSpace, legacyValue uint16) (rtid uint32, found bool) {
	baseGroup, found := c.getBaseNameGroup(name.BaseName())
	if !found {
		return uint32(c.rtidAir), false
	}
	return baseGroup.preciseMatchByLegacyValue(legacyValue)
}

func (c *ToNEMCConvertor) TryBestSearchByLegacyValue(name describe.BaseWithNameSpace, legacyValue uint16) (rtid uint32, found bool) {
	baseGroup, found := c.getBaseNameGroup(name.BaseName())
	if !found {
		return uint32(c.rtidAir), false
	}
	return baseGroup.fuzzySearchByLegacyValue(legacyValue)
}

func (c *ToNEMCConvertor) AddAnchorByState(name describe.BaseWithNameSpace, states *describe.PropsForSearch, runtimeID uint32, overwrite bool) (exist bool, conflictErr error) {
	baseNameGroup := c.ensureBaseNameGroup(name.BaseName())
	return baseNameGroup.addAnchorByState(states, runtimeID, overwrite)
}

func (c *ToNEMCConvertor) PreciseMatchByState(name describe.BaseWithNameSpace, states *describe.PropsForSearch) (rtid uint32, found bool) {
	baseGroup, found := c.getBaseNameGroup(name.BaseName())
	if !found {
		return uint32(c.rtidAir), false
	}
	return baseGroup.preciseMatchByState(states)
}

func (c *ToNEMCConvertor) TryBestSearchByState(name describe.BaseWithNameSpace, states *describe.PropsForSearch) (rtid uint32, score describe.ComparedOutput, matchAny bool) {
	baseGroup, found := c.getBaseNameGroup(name.BaseName())
	if !found {
		return uint32(c.rtidAir), describe.ComparedOutput{}, false
	}
	return baseGroup.fuzzySearchByState(states)
}

func NewToNEMCConverter(rtidUnknown, ritdAir uint32) *ToNEMCConvertor {
	return &ToNEMCConvertor{
		rtidUnknown: rtidUnknown,
		rtidAir:     ritdAir,
		baseNames:   map[string]*ToNEMCBaseNames{},
		mu:          sync.RWMutex{},
	}
}
