package builder

import (
	"fmt"
	"phoenixbuilder/fastbuilder/mcstructure"
)

// 测定 str 是否是方块状态。
// 不会检查其正确性
func test_block_states_string(str string) bool {
	reader := mcstructure.NewStringReader(str)
	location, exist := reader.GetCharacterWithNoSpace()
	if !exist {
		return false
	}
	reader.Pointer = location
	current, _ := reader.GetCurrentCharacter()
	return current == "["
}

// 将 blockStates 格式化为标准形式
func format_block_states(blockStates string) (string, error) {
	blockStatesMap, err := mcstructure.UnMarshalBlockStates(blockStates)
	if err != nil {
		return "", fmt.Errorf("format_block_states: %v", err)
	}
	blockStatesString, err := mcstructure.MarshalBlockStates(blockStatesMap)
	if err != nil {
		return "", fmt.Errorf("format_block_states: %v", err)
	}
	return blockStatesString, nil
}
