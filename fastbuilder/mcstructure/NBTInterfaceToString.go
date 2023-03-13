package mcstructure

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func stringifyNBTInterface(input interface{}) (string, error) {
	switch reflect.TypeOf(input).Kind() {
	case reflect.Uint8:
		return fmt.Sprintf("%vb", int(input.(byte))), nil
		// byte
	case reflect.Int16:
		return fmt.Sprintf("%vs", input.(int16)), nil
		// short
	case reflect.Int32:
		return fmt.Sprintf("%v", input.(int32)), nil
		// int
	case reflect.Int64:
		return fmt.Sprintf("%vl", input.(int64)), nil
		// long
	case reflect.Float32:
		return fmt.Sprintf("%vf", strconv.FormatFloat(float64(input.(float32)), 'f', -1, 32)), nil
		// float
	case reflect.Float64:
		return fmt.Sprintf("%vf", strconv.FormatFloat(float64(input.(float64)), 'f', -1, 32)), nil
		// double
	case reflect.Array:
		ans := []string{}
		value := reflect.ValueOf(input)
		// prepare
		switch reflect.TypeOf(input).Elem().Kind() {
		case reflect.Uint8:
			for i := 0; i < value.Len(); i++ {
				ans = append(ans, fmt.Sprintf("%vb", int(value.Index(i).Interface().(byte))))
			}
			return fmt.Sprintf("[B; %v]", strings.Join(ans, ", ")), nil
			// byte_array
		case reflect.Int32:
			for i := 0; i < value.Len(); i++ {
				ans = append(ans, fmt.Sprintf("%v", value.Index(i).Interface().(int32)))
			}
			return fmt.Sprintf("[I; %v]", strings.Join(ans, ", ")), nil
			// int_array
		case reflect.Int64:
			for i := 0; i < value.Len(); i++ {
				ans = append(ans, fmt.Sprintf("%vl", value.Index(i).Interface().(int64)))
			}
			return fmt.Sprintf("[L; %v]", strings.Join(ans, ", ")), nil
			// long_array
		}
		// byte_array, int_array, long_array
	case reflect.String:
		return fmt.Sprintf("\"%v\"", strings.Replace(input.(string), "\"", "\\\"", -1)), nil
		// string
	case reflect.Slice:
		value := input.([]interface{})
		list, err := ConvertListToString(value)
		if err != nil {
			return "", fmt.Errorf("stringifyNBTInterface: Failed in %#v", value)
		}
		return list, nil
		// list
	case reflect.Map:
		value := input.(map[string]interface{})
		compound, err := ConvertCompoundToString(value, false)
		if err != nil {
			return "", fmt.Errorf("stringifyNBTInterface: Failed in %#v", value)
		}
		return compound, nil
		// compound
	}
	return "", fmt.Errorf("stringifyNBTInterface: Failed because of unknown type of the target data, occurred in %#v", input)
}

func ConvertCompoundToString(input map[string]interface{}, outputBlockStatesMode bool) (string, error) {
	ans := make([]string, 0)
	for key, value := range input {
		key = strings.Replace(key, "\"", "\\\"", -1)
		if value == nil {
			return "", fmt.Errorf("ConvertCompoundToString: Crashed in input[\"%v\"]; errorLogs = value is nil; input = %#v", key, input)
		}
		got, err := stringifyNBTInterface(value)
		if err != nil {
			return "", fmt.Errorf("ConvertCompoundToString: Crashed in input[\"%v\"]; errorLogs = %v; input = %#v", key, err, input)
		}
		if got[len(got)-1] == "b"[0] && outputBlockStatesMode {
			if got == "0b" {
				got = "false"
			} else if got == "1b" {
				got = "true"
			} else {
				return "", fmt.Errorf("ConvertCompoundToString: Crashed in input[\"%v\"]; errorLogs = outputBlockStatesModeError; input = %#v", key, input)
			}
		}
		ans = append(ans, fmt.Sprintf("\"%v\": %v", key, got))
	}
	if outputBlockStatesMode {
		return fmt.Sprintf("[%v]", strings.Join(ans, ", ")), nil
	}
	return fmt.Sprintf("{%v}", strings.Join(ans, ", ")), nil
}

func ConvertListToString(input []interface{}) (string, error) {
	ans := make([]string, 0)
	for key, value := range input {
		if value == nil {
			return "", fmt.Errorf("ConvertListToString: Crashed in input[\"%v\"]; errorLogs = value is nil; input = %#v", key, input)
		}
		got, err := stringifyNBTInterface(value)
		if err != nil {
			return "", fmt.Errorf("ConvertListToString: Crashed in input[\"%v\"]; errorLogs = %v; input = %#v", key, err, input)
		}
		ans = append(ans, got)
	}
	return fmt.Sprintf("[%v]", strings.Join(ans, ", ")), nil
}
