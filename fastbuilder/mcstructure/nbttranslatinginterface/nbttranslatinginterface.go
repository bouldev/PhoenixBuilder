package nbttranslatinginterface

import (
	"fmt"
	"strconv"
	"strings"
)

// 判断 nbt 中 value 的数据类型
func GetData(input interface{}) (string, error) {
	value1, result := input.(byte)
	if result {
		return fmt.Sprintf("%vb", int(value1)), nil
	}
	// byte
	value2, result := input.(int16)
	if result {
		return fmt.Sprintf("%vs", value2), nil
	}
	// short
	value3, result := input.(int32)
	if result {
		return fmt.Sprintf("%v", value3), nil
	}
	// int
	value4, result := input.(int64)
	if result {
		return fmt.Sprintf("%vl", value4), nil
	}
	// long
	value5, result := input.(float32)
	if result {
		return fmt.Sprintf("%vf", strconv.FormatFloat(float64(value5), 'f', 16, 32)), nil
	}
	// float
	value6, result := input.(float64)
	if result {
		return fmt.Sprintf("%vd", strconv.FormatFloat(float64(value6), 'f', 16, 64)), nil
	}
	// double
	value, result := input.([]interface{})
	if result {
		if len(value) > 0 {
			_, result = value[0].(byte)
			if result {
				ans := make([]string, 0)
				for _, i := range value {
					got, err := i.(byte)
					if err {
						ans = append(ans, fmt.Sprintf("%vb", int(got)))
					} else {
						return "", fmt.Errorf("GetData: Failed")
					}
				}
				return fmt.Sprintf("[B; %v]", strings.Join(ans, ", ")), nil
			}
			// byte_array
			_, result = value[0].(int32)
			if result {
				ans := make([]string, 0)
				for _, i := range value {
					got, err := i.(int32)
					if err {
						ans = append(ans, fmt.Sprintf("%v", got))
					} else {
						return "", fmt.Errorf("GetData: Failed")
					}
				}
				return fmt.Sprintf("[I; %v]", strings.Join(ans, ", ")), nil
			}
			// int_array
			_, result = value[0].(int64)
			if result {
				ans := make([]string, 0)
				for _, i := range value {
					got, err := i.(int64)
					if err {
						ans = append(ans, fmt.Sprintf("%v", got))
					} else {
						return "", fmt.Errorf("GetData: Failed")
					}
				}
				return fmt.Sprintf("[L; %v]", strings.Join(ans, ", ")), nil
			}
			// long_array
		}
		got, err := List(value)
		if err != nil {
			return "", fmt.Errorf("GetData: Failed")
		} else {
			return got, nil
		}
		// list
	}
	// byte_array, int_array, long_array, list
	value7, result := input.(string)
	if result {
		return fmt.Sprintf("\"%v\"", value7), nil
	}
	// string
	value8, result := input.(map[string]interface{})
	if result {
		compound, err := Compound(value8, false)
		if err != nil {
			return "", fmt.Errorf("GetData: Failed")
		} else {
			return compound, nil
		}
	}
	// compound
	return "", fmt.Errorf("GetData: Failed")
}

func Compound(input map[string]interface{}, outputBlockStatesMode bool) (string, error) {
	ans := make([]string, 0)
	for key, value := range input {
		if value == nil {
			return "", fmt.Errorf("Compound: Crashed in input[\"%v\"]", key)
		}
		got, err := GetData(value)
		if err != nil {
			return "", fmt.Errorf("Compound: Crashed in input[\"%v\"]", key)
		} else {
			if got[len(got)-1] == "b"[0] && outputBlockStatesMode {
				if got == "0b" {
					got = "false"
				} else if got == "1b" {
					got = "true"
				} else {
					return "", fmt.Errorf("Compound: Crashed in input[\"%v\"]", key)
				}
			}
			ans = append(ans, fmt.Sprintf("\"%v\": %v", key, got))
		}
	}
	if outputBlockStatesMode {
		return fmt.Sprintf("[%v]", strings.Join(ans, ", ")), nil
	}
	return fmt.Sprintf("{%v}", strings.Join(ans, ", ")), nil
}

func List(input []interface{}) (string, error) {
	ans := make([]string, 0)
	for key, value := range input {
		if value == nil {
			return "", fmt.Errorf("List: Crashed in input[\"%v\"]", key)
		}
		got, err := GetData(value)
		if err != nil {
			return "", fmt.Errorf("List: Crashed in input[\"%v\"]", key)
		} else {
			ans = append(ans, got)
		}
	}
	return fmt.Sprintf("[%v]", strings.Join(ans, ", ")), nil
}
