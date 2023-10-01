package mcstructure

import (
	"fmt"
	"phoenixbuilder/fastbuilder/mc_command_parser"
	"strings"
)

func MarshalBlockStates(blockStates map[string]interface{}) (string, error) {
	temp := []string{}
	separator := mc_command_parser.BlockStatesDefaultSeparator
	for key, value := range blockStates {
		switch val := value.(type) {
		case string:
			temp = append(temp, fmt.Sprintf(
				"%#v%s%#v", key, separator, val,
			))
			// e.g. "color"="orange"
		case byte:
			switch val {
			case 0:
				temp = append(temp, fmt.Sprintf("%#v%sfalse", key, separator))
			case 1:
				temp = append(temp, fmt.Sprintf("%#v%strue", key, separator))
			default:
				return "", fmt.Errorf("MarshalBlockStates: Unexpected value %d(expect = 0 or 1) was found", val)
			}
			// e.g. "open_bit"=true
		case int32:
			temp = append(temp, fmt.Sprintf("%#v%s%d", key, separator, val))
			// e.g. "facing_direction"=0
		default:
			return "", fmt.Errorf("MarshalBlockStates: Unexpected data type of blockStates[%#v]; blockStates[%#v] = %#v", key, key, value)
		}
	}
	return fmt.Sprintf("[%s]", strings.Join(temp, ",")), nil
}

func UnmarshalBlockStates(blockStates string) (m map[string]interface{}, err error) {
	func() {
		defer func() {
			if errMessage := recover(); errMessage != nil {
				err = fmt.Errorf("UnmarshalBlockStates: %v", errMessage)
			}
		}()
		m = mc_command_parser.ParseBlockStates(blockStates)
	}()
	return
}
