package mcstructure

import (
	"encoding/json"
	"fmt"
	"strings"
)

// "color":"orange" [current]
// or
// "color"="orange"
const default_separator string = ":"

// 将 blockStates 编码为字符串形式。
// 该形式可直接被 setblock 命令所使用
func MarshalBlockStates(blockStates map[string]interface{}) (string, error) {
	temp := []string{}
	// 初始化
	for key, value := range blockStates {
		switch val := value.(type) {
		case string:
			temp = append(temp, fmt.Sprintf(
				"%#v%s%#v", key, default_separator, val,
			))
			// e.g. "color"="orange"
		case byte:
			switch val {
			case 0:
				temp = append(temp, fmt.Sprintf("%#v%sfalse", key, default_separator))
			case 1:
				temp = append(temp, fmt.Sprintf("%#v%strue", key, default_separator))
			default:
				return "", fmt.Errorf("MarshalBlockStates: Unexpected value %d(expect = 0 or 1) was found", val)
			}
			// e.g. "open_bit"=true
		case int32:
			temp = append(temp, fmt.Sprintf("%#v%s%d", key, default_separator, val))
			// e.g. "facing_direction"=0
		default:
			return "", fmt.Errorf("MarshalBlockStates: Unexpected data type of blockStates[%#v]; blockStates[%#v] = %#v", key, key, value)
		}
	}
	// 编码
	return fmt.Sprintf("[%s]", strings.Join(temp, ",")), nil
	// 返回值
}

// 将 blockStates 解码为 map[string]interface{} 。
// blockStates 应当可直接被 setblock 命令所使用
func UnMarshalBlockStates(blockStates string) (map[string]interface{}, error) {
	var version int
	reader := NewStringReader(blockStates)
	result := map[string]interface{}{}
	// 初始化
	temp, exist := reader.GetCharacterWithNoSpace()
	if !exist {
		return nil, fmt.Errorf(`UnMarshalBlockStates: EOF; blockStates = %#v`, blockStates)
	}
	reader.Pointer = temp
	// 跳过空格
	left_boundary, _ := reader.GetCurrentCharacter()
	if left_boundary != "[" {
		return nil, fmt.Errorf(`UnMarshalBlockStates: Unexpected first character %#v(expect = "[") was found; blockStates = %#v`, left_boundary, blockStates)
	}
	reader.Pointer++
	// 检查左边界 "[" 的正确性
	temp, exist = reader.GetCharacterWithNoSpace()
	if !exist {
		return nil, fmt.Errorf(`UnMarshalBlockStates: EOF; blockStates = %#v`, blockStates)
	}
	// 跳过空格
	reader.Pointer = temp
	current, _ := reader.GetCurrentCharacter()
	if current == "]" {
		return result, nil
	}
	// 如果提供的方块状态是一个空列表
	for {
		temp, exist := reader.GetCharacterWithNoSpace()
		if !exist {
			return nil, fmt.Errorf(`UnMarshalBlockStates: EOF; blockStates = %#v`, blockStates)
		}
		reader.Pointer = temp
		// 跳过空格
		left_boundary, _ := reader.GetCurrentCharacter()
		if left_boundary != `"` {
			return nil, fmt.Errorf("UnMarshalBlockStates: Unexpected left boundary %#v(expect = `\"`) was found; blockStates = %#v", left_boundary, blockStates)
		}
		reader.Pointer++
		// 检查左边界 `"` 的正确性
		right_boundary_location, exist := reader.GetRightBundary()
		if !exist {
			return nil, fmt.Errorf("UnMarshalBlockStates: EOF; blockStates = %#v", blockStates)
		}
		key := reader.Context[temp : right_boundary_location+1]
		json.Unmarshal([]byte(key), &key)
		reader.Pointer = right_boundary_location + 1
		// 获取键名。
		// e.g. "te\"st" -> te"st
		temp, exist = reader.GetCharacterWithNoSpace()
		if !exist {
			return nil, fmt.Errorf(`UnMarshalBlockStates: EOF; blockStates = %#v`, blockStates)
		}
		reader.Pointer = temp
		// 跳过空格
		separator, _ := reader.GetCurrentCharacter()
		switch separator {
		case ":":
			if version == 0 {
				version = 1
			} else if version != 1 {
				return nil, fmt.Errorf(`UnMarshalBlockStates: Unexpected separator %#v(expect = "=") was found; blockStates = %#v`, separator, blockStates)
			}
		case "=":
			if version == 0 {
				version = 2
			} else if version != 2 {
				return nil, fmt.Errorf(`UnMarshalBlockStates: Unexpected separator %#v(expect = ":") was found; blockStates = %#v`, separator, blockStates)
			}
		default:
			return nil, fmt.Errorf(`UnMarshalBlockStates: Unexpected separator %#v(expect = ":" or "=") was found; blockStates = %#v`, separator, blockStates)
		}
		reader.Pointer++
		// 检查连接符 ":" 或 "=" 的正确性
		temp, exist = reader.GetCharacterWithNoSpace()
		if !exist {
			return nil, fmt.Errorf(`UnMarshalBlockStates: EOF; blockStates = %#v`, blockStates)
		}
		reader.Pointer = temp
		// 跳过空格
		left_boundary, _ = reader.GetCurrentCharacter()
		if left_boundary == `"` {
			reader.Pointer++
			right_boundary_location, exist = reader.GetRightBundary()
			if !exist {
				return nil, fmt.Errorf("UnMarshalBlockStates: EOF; blockStates = %#v", blockStates)
			}
			value := reader.Context[temp : right_boundary_location+1]
			json.Unmarshal([]byte(value), &value)
			result[key] = value
			reader.Pointer = right_boundary_location + 1
			// e.g. "orange\"" -> orange"
		} else if left_boundary == "+" || left_boundary == "-" || left_boundary == "0" || left_boundary == "1" || left_boundary == "2" || left_boundary == "3" || left_boundary == "4" || left_boundary == "5" || left_boundary == "6" || left_boundary == "7" || left_boundary == "8" || left_boundary == "9" {
			value, _ := reader.GetInt()
			result[key] = int32(value)
			// e.g. +3 -> int32(3)
		} else if left_boundary == "t" || left_boundary == "f" || left_boundary == "T" || left_boundary == "F" {
			value, _ := reader.GetBool()
			if value {
				result[key] = byte(1)
			} else {
				result[key] = byte(0)
			}
			// e.g. FaLsE -> byte(0)
		} else {
			return nil, fmt.Errorf("UnMarshalBlockStates: Unexpected left boundary %#v was found; blockStates = %#v", left_boundary, blockStates)
		}
		// 获取该方块状态的值
		temp, exist = reader.GetCharacterWithNoSpace()
		if !exist {
			return nil, fmt.Errorf(`UnMarshalBlockStates: EOF; blockStates = %#v`, blockStates)
		}
		reader.Pointer = temp
		// 跳过空格
		current, _ = reader.GetCurrentCharacter()
		switch current {
		case ",":
			reader.Pointer++
		case "]":
			return result, nil
		}
		// 返回值
	}
	// 以逐个读入的方式解析该方块状态的字符串形式
}
