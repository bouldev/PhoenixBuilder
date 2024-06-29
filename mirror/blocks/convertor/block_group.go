package convertor

import (
	"fmt"
	"phoenixbuilder/mirror/blocks/describe"
	"sync"
)

type ToNEMCBaseNames struct {
	RtidUnknown uint32
	RtidAir     uint32

	legacyValuesMapping []uint32
	statesWithRtid      []struct {
		states *describe.PropsForSearch
		rtid   uint32
	}
	StatesWithRtidQuickMatch map[string]uint32
	mu                       sync.RWMutex
}

func (baseNameGroup *ToNEMCBaseNames) addAnchorByLegacyValue(legacyValue uint16, rtid uint32, overwrite bool) (exist bool, conflictErr error) {
	if int(legacyValue+1) <= len(baseNameGroup.legacyValuesMapping) {
		if baseNameGroup.legacyValuesMapping[legacyValue] == baseNameGroup.RtidUnknown || overwrite {
			baseNameGroup.legacyValuesMapping[legacyValue] = rtid
			return false, nil
		} else if baseNameGroup.legacyValuesMapping[legacyValue] != rtid && !overwrite {
			return true, fmt.Errorf("conflict runtime id ")
		} else {
			return true, nil
		}
	}
	baseNameGroup.mu.Lock()
	defer baseNameGroup.mu.Unlock()
	for int(legacyValue+1) > len(baseNameGroup.legacyValuesMapping) {
		baseNameGroup.legacyValuesMapping = append(baseNameGroup.legacyValuesMapping, baseNameGroup.RtidUnknown)
	}
	baseNameGroup.legacyValuesMapping[legacyValue] = rtid
	return false, nil
}

func (baseNameGroup *ToNEMCBaseNames) preciseMatchByLegacyValue(legacyValue uint16) (rtid uint32, found bool) {
	if int(legacyValue+1) <= len(baseNameGroup.legacyValuesMapping) {
		if rtid = baseNameGroup.legacyValuesMapping[legacyValue]; rtid == baseNameGroup.RtidUnknown {
			return uint32(baseNameGroup.RtidAir), false
		} else {
			return rtid, true
		}
	} else {
		return uint32(baseNameGroup.RtidAir), false
	}
}

func (baseNameGroup *ToNEMCBaseNames) fuzzySearchByLegacyValue(legacyValue uint16) (rtid uint32, found bool) {
	if int(legacyValue+1) <= len(baseNameGroup.legacyValuesMapping) {
		if rtid = baseNameGroup.legacyValuesMapping[legacyValue]; rtid != baseNameGroup.RtidUnknown {
			return rtid, true
		}
	}
	if int(legacyValue+1) <= len(baseNameGroup.statesWithRtid) {
		return baseNameGroup.statesWithRtid[legacyValue].rtid, true
	}
	return baseNameGroup.statesWithRtid[0].rtid, true
}

func (baseNameGroup *ToNEMCBaseNames) addAnchorByState(states *describe.PropsForSearch, runtimeID uint32, overwrite bool) (exist bool, conflictErr error) {
	quickMatchStr := "{}"
	if states != nil {
		quickMatchStr = states.InPreciseSNBT()
	}
	baseNameGroup.mu.RLock()
	if currentRuntimeID, found := baseNameGroup.StatesWithRtidQuickMatch[quickMatchStr]; found {
		if currentRuntimeID == runtimeID {
			baseNameGroup.mu.RUnlock()
			return true, nil
		} else if !overwrite {
			baseNameGroup.mu.RUnlock()
			return true, fmt.Errorf("conflict runtime id ")
		}
	}
	baseNameGroup.mu.RUnlock()
	baseNameGroup.mu.Lock()
	defer baseNameGroup.mu.Unlock()
	baseNameGroup.statesWithRtid = append(baseNameGroup.statesWithRtid, struct {
		states *describe.PropsForSearch
		rtid   uint32
	}{states: states, rtid: runtimeID})
	baseNameGroup.StatesWithRtidQuickMatch[quickMatchStr] = runtimeID
	return false, nil
}

func (baseNameGroup *ToNEMCBaseNames) preciseMatchByState(states *describe.PropsForSearch) (rtid uint32, found bool) {
	quickMatchStr := states.InPreciseSNBT()
	baseNameGroup.mu.RLock()
	defer baseNameGroup.mu.RUnlock()
	rtid, found = baseNameGroup.StatesWithRtidQuickMatch[quickMatchStr]
	return rtid, found
}

func (baseNameGroup *ToNEMCBaseNames) fuzzySearchByState(states *describe.PropsForSearch) (rtid uint32, score describe.ComparedOutput, matchAny bool) {
	quickMatchStr := states.InPreciseSNBT()
	baseNameGroup.mu.RLock()
	defer baseNameGroup.mu.RUnlock()
	rtid, found := baseNameGroup.StatesWithRtidQuickMatch[quickMatchStr]
	if found {
		sameCount := uint8(0)
		if states != nil {
			sameCount = uint8(states.NumProps())
		}
		return rtid, describe.ComparedOutput{Same: sameCount}, true
	}
	rtid = uint32(baseNameGroup.RtidAir)
	matchAny = false
	for _, anchor := range baseNameGroup.statesWithRtid {
		newScore := anchor.states.Compare(states)
		if (!matchAny) || newScore.Same > score.Same || (newScore.Same == score.Same && ((newScore.Different + newScore.Redundant + newScore.Missing) < (score.Different + score.Redundant + score.Missing))) {
			score = newScore
			rtid = anchor.rtid
		}
		matchAny = true
	}
	return rtid, score, matchAny
}
