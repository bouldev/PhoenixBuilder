package builder

import (
	"fmt"
	"phoenixbuilder/fastbuilder/mcstructure"
	"phoenixbuilder/fastbuilder/string_reader"
)

func is_block_states(str string) bool {
	reader := string_reader.NewStringReader(&str)
	reader.JumpSpace()
	return reader.Next(true) == "["
}

func format_block_states(blockStates string) (string, error) {
	blockStatesMap, err := mcstructure.UnmarshalBlockStates(blockStates)
	if err != nil {
		return "", fmt.Errorf("format_block_states: %v", err)
	}
	blockStatesString, err := mcstructure.MarshalBlockStates(blockStatesMap)
	if err != nil {
		return "", fmt.Errorf("format_block_states: %v", err)
	}
	return blockStatesString, nil
}
