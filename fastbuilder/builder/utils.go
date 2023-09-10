package builder

import (
	"fmt"
	"regexp"
	"phoenixbuilder/fastbuilder/mcstructure"
)

func is_block_states(str string) bool {
	matcher:=regexp.MustCompile(" {0,}\\[( {0,}\"(.*?)\" {0,}(=|:) {0,}((t|T)rue|(F|f)alse|null|(\\+|\\-)\\d+|\".*?(?<!\\\\)\") {0,},?){0,}\\] {0,}")
	return matcher.MatchString(str)
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
