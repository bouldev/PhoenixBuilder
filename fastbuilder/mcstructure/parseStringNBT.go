package mcstructure

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type stringNBT struct {
	context string
	pointer int
}

func (snbt *stringNBT) getPartOfString(stringLength int) string {
	if snbt.pointer < 0 {
		snbt.pointer = 0
	}
	end := snbt.pointer + stringLength
	if end > len(snbt.context)-1 {
		return snbt.context[snbt.pointer:]
	} else {
		return snbt.context[snbt.pointer:end]
	}
}

func (snbt *stringNBT) jumpSpace() error {
	for {
		if snbt.pointer > len(snbt.context)-1 {
			return fmt.Errorf("jumpSpace: %v out of length(%v)", snbt.pointer, len(snbt.context))
		} else if snbt.getPartOfString(1) == " " || snbt.getPartOfString(1) == "\n" || snbt.getPartOfString(1) == "\t" {
			snbt.pointer++
		} else {
			return nil
		}
	}
}

func (snbt *stringNBT) index(searchingFor string) (int, error) {
	if snbt.pointer > len(snbt.context)-1 {
		return 0, fmt.Errorf("index: %v out of length(%v)", snbt.pointer, len(snbt.context))
	}
	find := strings.Index(snbt.context[snbt.pointer:], searchingFor)
	if find == -1 {
		return 0, fmt.Errorf("index: %v not found", searchingFor)
	} else {
		return find + snbt.pointer, nil
	}
}

func (snbt *stringNBT) highSearching(input []string) (struct {
	begin int
	end   int
}, error) {
	ansSave := []struct {
		begin int
		end   int
	}{}
	for _, value := range input {
		got, err := snbt.index(value)
		if err == nil {
			ansSave = append(ansSave, struct {
				begin int
				end   int
			}{
				begin: got,
				end:   got + len(value),
			})
		}
	}
	minRecord := struct {
		begin int
		end   int
	}{
		begin: 2147483647,
	}
	success := false
	for _, value := range ansSave {
		if value.begin < minRecord.begin {
			minRecord = value
			success = true
		}
	}
	if !success {
		return struct {
			begin int
			end   int
		}{}, fmt.Errorf("highSearching: Nothing found")
	}
	return minRecord, nil
}

func (snbt *stringNBT) getRightBarrier() (int, error) {
	for {
		barrier, err := snbt.index("\"")
		if err != nil {
			return 0, fmt.Errorf("getRightBarrier: Right barrier not found")
		}
		if barrier > 0 {
			if snbt.context[barrier-1:barrier] != "\\" {
				return barrier, nil
			} else {
				snbt.pointer = barrier + 1
			}
		} else {
			return barrier, nil
		}
	}
}

func (snbt *stringNBT) getKey() (string, error) {
	err := snbt.jumpSpace()
	if err != nil {
		return "", fmt.Errorf("getKey: Incomplete key")
	}
	if snbt.getPartOfString(1) == "\"" {
		snbt.pointer++
		save := snbt.pointer
		rightBarrierLocation, err := snbt.getRightBarrier()
		if err != nil {
			return "", fmt.Errorf("getKey: Right barrier not found")
		}
		snbt.pointer = rightBarrierLocation + 1
		return strings.Replace(snbt.context[save:rightBarrierLocation], "\\\"", "\"", -1), nil
	} else {
		got, err := snbt.highSearching([]string{":", " ", "\n", "\t"})
		if err != nil {
			return "", fmt.Errorf("getKey: Boundary not found")
		}
		save := snbt.pointer
		snbt.pointer = got.begin
		return snbt.context[save:got.begin], nil
	}
}

func (snbt *stringNBT) connectome() error {
	err := snbt.jumpSpace()
	if err != nil {
		return fmt.Errorf("connectome: Incomplete connectome")
	}
	if snbt.getPartOfString(1) != ":" {
		return fmt.Errorf("connectome: \":\" not found")
	}
	snbt.pointer++
	err = snbt.jumpSpace()
	if err != nil {
		return fmt.Errorf("connectome: Incomplete connectome")
	}
	return nil
}

func (snbt *stringNBT) getValue() (interface{}, error) {
	var valueString string
	// prepare
	switch snbt.getPartOfString(1) {
	case "\"":
		snbt.pointer++
		save := snbt.pointer
		endLocation, err := snbt.getRightBarrier()
		if err != nil {
			return nil, fmt.Errorf("getValue: Right barrier '\"' not found")
		}
		valueString = snbt.context[save:endLocation]
		snbt.pointer = endLocation + 1
		return strings.Replace(valueString, "\\\"", "\"", -1), nil
		// string
	case "[":
		got, err := snbt.getListOrArray()
		if err != nil {
			return nil, fmt.Errorf("getValue: %v", err)
		}
		return got, nil
		// list
	case "{":
		got, err := snbt.getCompound()
		if err != nil {
			return nil, fmt.Errorf("getValue: %v", err)
		}
		return got, nil
		// compound
	default:
		got, err := snbt.highSearching([]string{",", "}", "]"})
		if err != nil {
			return nil, fmt.Errorf("getValue: Right barrier not found")
		}
		valueString = snbt.context[snbt.pointer:got.begin]
		snbt.pointer = got.begin
		// others
	}
	// get value
	for {
		if valueString[len(valueString)-1:] == " " || valueString[len(valueString)-1:] == "\n" || valueString[len(valueString)-1:] == "\t" {
			valueString = valueString[:len(valueString)-1]
		} else {
			break
		}
	}
	if len(valueString) <= 0 {
		return nil, fmt.Errorf("getValue: Invalid value")
	}
	switch strings.ToLower(valueString[len(valueString)-1:]) {
	case "b":
		got, err := strconv.ParseInt(valueString[:len(valueString)-1], 10, 8)
		if err != nil {
			return nil, fmt.Errorf("getValue: Invalid TAG_Byte")
		}
		return byte(got), nil
		// byte
	case "s":
		got, err := strconv.ParseInt(valueString[:len(valueString)-1], 10, 16)
		if err != nil {
			return nil, fmt.Errorf("getValue: Invalid TAG_Short")
		}
		return int16(got), nil
		// short
	case "l":
		got, err := strconv.ParseInt(valueString[:len(valueString)-1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("getValue: Invalid TAG_Long")
		}
		return got, nil
		// long
	case "f":
		got, err := strconv.ParseFloat(valueString[:len(valueString)-1], 32)
		if err != nil {
			return nil, fmt.Errorf("getValue: Invalid TAG_Float")
		}
		return float32(got), nil
		// float
	case "d":
		got, err := strconv.ParseFloat(valueString[:len(valueString)-1], 64)
		if err != nil {
			return nil, fmt.Errorf("getValue: Invalid TAG_Double")
		}
		return got, nil
		// double
	default:
		got1, err1 := strconv.ParseInt(valueString, 10, 32)
		got2, err2 := strconv.ParseFloat(valueString, 64)
		if err1 != nil && err2 != nil {
			if strings.ToLower(valueString) == "true" {
				return byte(1), nil
			}
			if strings.ToLower(valueString) == "false" {
				return byte(0), nil
			}
			// boolean(-> byte)
			if strings.Contains(valueString, " ") || strings.Contains(valueString, "\n") || strings.Contains(valueString, "\t") {
				return nil, fmt.Errorf("getValue: Invalid TAG_String")
			}
			return valueString, nil
			// string
		}
		if err1 == nil {
			if valueString[0] == "0"[0] && len(valueString) > 1 {
				return nil, fmt.Errorf("getValue: Invalid TAG_Int")
			}
			if valueString[0] == "-"[0] {
				if valueString[1] == "0"[0] && len(valueString) > 2 {
					return nil, fmt.Errorf("getValue: Invalid TAG_Int")
				}
			}
			return int32(got1), nil
		}
		// int
		return float64(got2), nil
		// double
	}
}

func (snbt *stringNBT) getListOrArray() (interface{}, error) {
	snbt.pointer++
	switch snbt.getPartOfString(2) {
	case "B;":
		snbt.pointer = snbt.pointer + 2
		ans := []byte{}
		for {
			err := snbt.jumpSpace()
			if err != nil {
				return nil, fmt.Errorf("getListOrArray: Incomplete TAG_Byte_Array")
			}
			if snbt.getPartOfString(1) == "]" {
				snbt.pointer++
				break
			} else if snbt.getPartOfString(1) == "," && len(ans) > 0 {
				snbt.pointer++
				err = snbt.jumpSpace()
				if err != nil {
					return nil, fmt.Errorf("getListOrArray: Incomplete TAG_Byte_Array")
				}
			} else if len(ans) > 0 {
				return nil, fmt.Errorf("getListOrArray: Incomplete TAG_Byte_Array")
			}
			got, err := snbt.getValue()
			if err != nil {
				return nil, fmt.Errorf("getListOrArray: Invalid TAG_Byte_Array")
			}
			GOT, normal := got.(byte)
			if !normal {
				return nil, fmt.Errorf("getListOrArray: Invalid TAG_Byte_Array")
			}
			ans = append(ans, GOT)
		}
		result := reflect.ValueOf(reflect.New(reflect.ArrayOf(len(ans), reflect.TypeOf(byte(0)))).Interface())
		for key, value := range ans {
			result.Elem().Index(key).SetUint(uint64(value))
		}
		return result.Elem().Interface(), nil
		// byte_array
	case "I;":
		snbt.pointer = snbt.pointer + 2
		ans := []int32{}
		for {
			err := snbt.jumpSpace()
			if err != nil {
				return nil, fmt.Errorf("getListOrArray: Incomplete TAG_Int_Array")
			}
			if snbt.getPartOfString(1) == "]" {
				snbt.pointer++
				break
			} else if snbt.getPartOfString(1) == "," && len(ans) > 0 {
				snbt.pointer++
				err = snbt.jumpSpace()
				if err != nil {
					return nil, fmt.Errorf("getListOrArray: Incomplete TAG_Int_Array")
				}
			} else if len(ans) > 0 {
				return nil, fmt.Errorf("getListOrArray: Incomplete TAG_Int_Array")
			}
			got, err := snbt.getValue()
			if err != nil {
				return nil, fmt.Errorf("getListOrArray: Invalid TAG_Int_Array")
			}
			GOT, normal := got.(int32)
			if !normal {
				return nil, fmt.Errorf("getListOrArray: Invalid TAG_Int_Array")
			}
			ans = append(ans, GOT)
		}
		result := reflect.ValueOf(reflect.New(reflect.ArrayOf(len(ans), reflect.TypeOf(int32(0)))).Interface())
		for key, value := range ans {
			result.Elem().Index(key).SetInt(int64(value))
		}
		return result.Elem().Interface(), nil
		// int_array
	case "L;":
		snbt.pointer = snbt.pointer + 2
		ans := []int64{}
		for {
			err := snbt.jumpSpace()
			if err != nil {
				return nil, fmt.Errorf("getListOrArray: Incomplete TAG_Long_Array")
			}
			if snbt.getPartOfString(1) == "]" {
				snbt.pointer++
				break
			} else if snbt.getPartOfString(1) == "," && len(ans) > 0 {
				snbt.pointer++
				err = snbt.jumpSpace()
				if err != nil {
					return nil, fmt.Errorf("getListOrArray: Incomplete TAG_Long_Array")
				}
			} else if len(ans) > 0 {
				return nil, fmt.Errorf("getListOrArray: Incomplete TAG_Long_Array")
			}
			got, err := snbt.getValue()
			if err != nil {
				return nil, fmt.Errorf("getListOrArray: Invalid TAG_Long_Array")
			}
			GOT, normal := got.(int64)
			if !normal {
				return nil, fmt.Errorf("getListOrArray: Invalid TAG_Long_Array")
			}
			ans = append(ans, GOT)
		}
		result := reflect.ValueOf(reflect.New(reflect.ArrayOf(len(ans), reflect.TypeOf(int64(0)))).Interface())
		for key, value := range ans {
			result.Elem().Index(key).SetInt(value)
		}
		return result.Elem().Interface(), nil
		// long_array
	default:
		ans := []interface{}{}
		for {
			err := snbt.jumpSpace()
			if err != nil {
				return nil, fmt.Errorf("getListOrArray: Incomplete TAG_List")
			}
			if snbt.getPartOfString(1) == "]" {
				snbt.pointer++
				break
			} else if snbt.getPartOfString(1) == "," && len(ans) > 0 {
				snbt.pointer++
				err = snbt.jumpSpace()
				if err != nil {
					return nil, fmt.Errorf("getListOrArray: Incomplete TAG_List")
				}
			} else if len(ans) > 0 {
				return nil, fmt.Errorf("getListOrArray: Incomplete TAG_List")
			}
			got, err := snbt.getValue()
			if err != nil {
				return nil, fmt.Errorf("getListOrArray: Invalid TAG_List")
			}
			ans = append(ans, got)
		}
		var typeOfList reflect.Kind
		for key, value := range ans {
			if key == 0 {
				typeOfList = reflect.TypeOf(value).Kind()
				continue
			}
			if reflect.TypeOf(value).Kind() != typeOfList {
				return nil, fmt.Errorf("getListOrArray: Invalid TAG_List")
			}
		}
		return ans, nil
		// list
	}
}

func (snbt *stringNBT) getCompound() (map[string]interface{}, error) {
	snbt.pointer++
	ans := map[string]interface{}{}
	for {
		err := snbt.jumpSpace()
		if err != nil {
			return map[string]interface{}{}, fmt.Errorf("getCompound: TAG_Compound")
		}
		if snbt.getPartOfString(1) == "}" {
			snbt.pointer++
			return ans, nil
		} else if snbt.getPartOfString(1) == "," && len(ans) > 0 {
			snbt.pointer++
		} else if len(ans) > 0 {
			return nil, fmt.Errorf("getCompound: Incomplete TAG_Compound")
		}
		key, err := snbt.getKey()
		if err != nil {
			return nil, fmt.Errorf("getCompound: Failed to get the key of value, and the error log is %v", err)
		}
		err = snbt.connectome()
		if err != nil {
			return nil, fmt.Errorf("getCompound: Failed to get the key of value, and the error log is %v", err)
		}
		value, err := snbt.getValue()
		if err != nil {
			return nil, fmt.Errorf("getCompound: Failed to get the value, and the error log is %v", err)
		}
		ans[key] = value
	}
}

func ParseStringNBT(SNBT string, IsParseBlockStates bool) (interface{}, error) {
	reader := stringNBT{
		context: SNBT,
		pointer: 0,
	}
	// prepare
	err := reader.jumpSpace()
	if err != nil {
		reader.pointer = reader.pointer - 5
		return nil, fmt.Errorf("ParseStringNBT: Failed to parse the target string-nbt, and the error may occurred in >>>%v<<<; SNBT = %#v", reader.getPartOfString(10), SNBT)
	}
	// prepare
	if IsParseBlockStates {
		if reader.getPartOfString(1) == "[" {
			reader.context = fmt.Sprintf("{%v", reader.context[1:])
		} else {
			return nil, fmt.Errorf("ParseStringNBT: Failed to parse the target string-nbt, and the error may occurred in >>>%v<<<; SNBT = %#v", reader.getPartOfString(10), SNBT)
		}
		for {
			if reader.context[len(reader.context)-1:] == " " || reader.context[len(reader.context)-1:] == "\n" || reader.context[len(reader.context)-1:] == "\t" {
				reader.context = reader.context[:len(reader.context)-1]
			} else {
				break
			}
		}
		if reader.context[len(reader.context)-1:] == "]" {
			reader.context = fmt.Sprintf("%v}", reader.context[:len(reader.context)-1])
		} else {
			reader.pointer = len(reader.context) - 5
			return nil, fmt.Errorf("ParseStringNBT: Failed to parse the target string-nbt, and the error may occurred in >>>%v<<<; SNBT = %#v", reader.getPartOfString(10), SNBT)
		}
	}
	// prepare to parse block states
	switch reader.getPartOfString(1) {
	case "{":
		compound, err := reader.getCompound()
		if err != nil {
			reader.pointer = reader.pointer - 5
			return nil, fmt.Errorf("ParseStringNBT: Failed to parse the target string-nbt, and the error may occurred in >>>%v<<<; SNBT = %#v", reader.getPartOfString(10), SNBT)
		}
		return compound, nil
		// compound
	case "[":
		list, err := reader.getListOrArray()
		if err != nil {
			reader.pointer = reader.pointer - 5
			return nil, fmt.Errorf("ParseStringNBT: Failed to parse the target string-nbt, and the error may occurred in >>>%v<<<; SNBT = %#v", reader.getPartOfString(10), SNBT)
		}
		return list, nil
		// list
	default:
		reader.context = fmt.Sprintf("%v,", reader.context)
		value, err := reader.getValue()
		if err != nil {
			reader.context = reader.context[:len(reader.context)-1]
			reader.pointer = reader.pointer - 5
			return nil, fmt.Errorf("ParseStringNBT: Failed to parse the target string-nbt, and the error may occurred in >>>%v<<<; SNBT = %#v", reader.getPartOfString(10), SNBT)
		}
		return value, nil
		// value
	}
}
