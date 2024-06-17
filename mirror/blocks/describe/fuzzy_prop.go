package describe

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Fuzzy Prop Val 的目的是解决以下情况:
// 当一个属性的 key 相同时, 两个值为:
//
//  1. uint8 <-> uint8: true/false==true/false? -> true
//  2. uint8 <-> int32: true/false==(1,2,3...)/0 ? -> true
//  3. uint8 <-> string:
//     true/false == "true"/"false" ? -> true
//     true/false == "0b"/"1b" ? -> true
//     true/false == "0"/"1" ? true
//  4. int32 <-> int32: 0,1,2,3==0,1,2,3 ? -> true
//  5. int32 <-> string:
//     0,1,2,3...=="0"/"1"/"2","3"... ? -> true
//     0,1=="0b"/"1b" ? -> true
//     0,1=="true"/"false" ? -> true
// 6. string <-> string: "val"=="val" -> true

// 模糊类型转换和 Type
// uint8 	-> 	0/1 	; 0/1 		; -
// int32 	-> 	0/1/-	; 0,1,2,...	; "0","1","2"
// string	->  0/1		; 0,1,2,...	; "val"

// 这种接口是为了转换时能以最大概率匹配成功而设计的
type PropValForSearch interface {
	// 对于 FuzzyPropVal, 其 type 可能为 PropValUint8/PropValInt32/PropValString 中的多种
	Type() PropValType
	// 返回结果为 (Type() & val)!=0
	HasType(val PropValType) bool
	// 若不为 PropValUint8 则 panic
	Uint8Val() uint8
	// 若不为 PropValInt32 则 panic
	Int32Val() int32
	// 若不为 PropValString 则 panic
	StringVal() string
	// fuzzy equal, e.g. PrecisePropVal(true).Equal(FuzzyPropVal("1")) -> true
	FuzzyEqual(val2 PropValForSearch) bool
	// uint8 -> 0b/1b int32 -> 1,2,3,... string => "true"->1b/"false"->0b/"0"->0/"1"->1/"2"->2/...
	InPreciseSNBT() string
}

type fuzzyUint8 struct {
	val bool
}

func (v fuzzyUint8) Type() PropValType {
	return PropValTypeUint8 | PropValTypeInt32 | PropValTypeString
}

func (v fuzzyUint8) HasType(val PropValType) bool {
	return (v.Type() & val) != 0
}

func (v fuzzyUint8) Uint8Val() uint8 {
	if v.val {
		return uint8(1)
	}
	return uint8(0)
}

func (v fuzzyUint8) Int32Val() int32 {
	if v.val {
		return int32(1)
	}
	return int32(0)
}

func (v fuzzyUint8) StringVal() string {
	if v.val {
		return "true"
	}
	return "false"
}

func (v fuzzyUint8) FuzzyEqual(val2 PropValForSearch) bool {
	// 1. uint8 <-> uint8: true/false==true/false? -> true
	if val2.HasType(PropValTypeUint8) {
		return val2.Uint8Val() == v.Uint8Val()
	}
	// uint8 <-> int32: true/false==(1,2,3...)/0 ? -> true
	if val2.HasType(PropValTypeInt32) {
		if val2.Int32Val() == 0 {
			return !v.val
		}
		return v.val
	}
	//  3. uint8 <-> string:
	//     true/false == "true"/"false" ? -> true
	//     true/false == "0b"/"1b" ? -> true
	//     true/false == "0"/"1" ? true
	v2 := val2.StringVal()
	if v.val {
		if v2 == "true" || v2 == "1b" || v2 == "1" {
			return true
		}
		return false
	}
	if v2 == "false" || v2 == "0b" || v2 == "0" {
		return true
	}
	return false
}

func (v fuzzyUint8) InPreciseSNBT() string {
	if v.val {
		return "1b"
	} else {
		return "0b"
	}
}

var FuzzyPropVal0 = fuzzyUint8{val: false}
var FuzzyPropVal1 = fuzzyUint8{val: true}

func FuzzyPropValForSearchFromBool(b bool) PropValForSearch {
	if b {
		return FuzzyPropVal1
	}
	return FuzzyPropVal0
}

type fuzzyInt32 struct {
	val    int32
	strVal string
}

func (v fuzzyInt32) Type() PropValType {
	t := PropValTypeInt32 | PropValTypeString
	if v.val == 0 || v.val == 1 {
		t = t | PropValTypeUint8
	}
	return t
}

func (v fuzzyInt32) HasType(val PropValType) bool {
	return (v.Type() & val) != 0
}

func (v fuzzyInt32) Uint8Val() uint8 {
	if v.val == 0 {
		return uint8(0)
	} else if v.val == 1 {
		return uint8(1)
	}
	panic("not uint8")
}

func (v fuzzyInt32) Int32Val() int32 {
	return v.val
}

func (v fuzzyInt32) StringVal() string {
	return v.strVal
}

func (v fuzzyInt32) FuzzyEqual(val2 PropValForSearch) bool {
	if val2.HasType(PropValTypeUint8) {
		if v.HasType(PropValTypeUint8) {
			return v.Uint8Val() == val2.Uint8Val()
		}
		return val2.FuzzyEqual(v)
	}
	//  4. int32 <-> int32: 0,1,2,3==0,1,2,3 ? -> true
	if val2.HasType(PropValTypeInt32) {
		return v.val == val2.Int32Val()
	}
	//  5. int32 <-> string:
	//     0,1,2,3...=="0"/"1"/"2","3"... ? -> true
	//     0,1=="0b"/"1b" ? -> true
	//     0,1=="true"/"false" ? -> true
	v2 := val2.StringVal()
	if v.val == 1 {
		if v2 == "true" || v2 == "1b" || v2 == "1" {
			return true
		}
		return false
	} else if v.val == 0 {
		if v2 == "false" || v2 == "0b" || v2 == "0" {
			return true
		}
		return false
	} else {
		return v.strVal == v2
	}
}

func (v fuzzyInt32) InPreciseSNBT() string {
	return v.strVal
}

var pregenererateFuzzyPropValInt32 []fuzzyInt32

func init() {
	pregenererateFuzzyPropValInt32 = make([]fuzzyInt32, 128)
	for i := int32(0); i < 128; i++ {
		pregenererateFuzzyPropValInt32[i] = fuzzyInt32{i, fmt.Sprintf("%v", i)}
	}
}

func FuzzyPropValForSearchFromInt32(val int32) PropValForSearch {
	if val < 128 {
		return pregenererateFuzzyPropValInt32[val]
	}
	return fuzzyInt32{val, fmt.Sprintf("%v", val)}
}

type fuzzyString struct {
	val    int32
	strVal string
}

func (v fuzzyString) Type() PropValType {
	if v.val == -1 {
		return PropValTypeString
	}
	if v.val == 0 || v.val == 1 {
		return PropValTypeString | PropValTypeInt32 | PropValTypeUint8
	}
	return PropValTypeString | PropValTypeInt32
}

func (v fuzzyString) HasType(val PropValType) bool {
	return (v.Type() & val) != 0
}

func (v fuzzyString) Uint8Val() uint8 {
	if v.val == 0 {
		return uint8(0)
	} else if v.val == 1 {
		return uint8(1)
	}
	panic("not uint8")
}

func (v fuzzyString) Int32Val() int32 {
	if v.val == -1 {
		panic("not int32")
	}
	return v.val
}

func (v fuzzyString) StringVal() string {
	return v.strVal
}

func (v fuzzyString) FuzzyEqual(val2 PropValForSearch) bool {
	if val2.HasType(PropValTypeUint8) {
		if v.HasType(PropValTypeUint8) {
			return v.Uint8Val() == val2.Uint8Val()
		}
		return val2.FuzzyEqual(v)
	}
	if val2.HasType(PropValTypeInt32) {
		if v.HasType(PropValTypeInt32) {
			return v.Int32Val() == val2.Int32Val()
		}
		return val2.FuzzyEqual(v)
	}
	return v.strVal == val2.StringVal()
}

func (v fuzzyString) InPreciseSNBT() string {
	if v.strVal == "true" || v.strVal == "1b" {
		return "1b"
	} else if v.strVal == "false" || v.strVal == "0b" {
		return "0b"
	}
	return v.strVal
}

func FuzzyPropValForSearchFromString(val string) fuzzyString {
	val = strings.TrimSuffix(strings.TrimPrefix(val, "\""), "\"")
	fs := fuzzyString{strVal: val, val: -1}
	if val == "true" || val == "1b" || val == "1" {
		fs.val = 1
		return fs
	} else if val == "false" || val == "0b" || val == "0" {
		fs.val = 0
		return fs
	}
	intVal, err := strconv.Atoi(val)
	if err == nil {
		fs.val = int32(intVal)
	}
	return fs
}

// func assertPropSame(v1, v2 PropValForSearch) {
// 	if !v1.FuzzyEqual(v2) {
// 		panic(fmt.Errorf("%v!=%v", v1, v2))
// 	}
// 	if !v2.FuzzyEqual(v1) {
// 		panic(fmt.Errorf("%v!=%v", v1, v2))
// 	}
// }

// func init() {
// 	uint80 := FuzzyPropValForSearchFromBool(false)
// 	uint81 := FuzzyPropValForSearchFromBool(true)
// 	int320 := FuzzyPropValForSearchFromInt32(0)
// 	int321 := FuzzyPropValForSearchFromInt32(1)
// 	int32240 := FuzzyPropValForSearchFromInt32(240)
// 	string0b := FuzzyPropValForSearchFromString("0b")
// 	string1b := FuzzyPropValForSearchFromString("1b")
// 	stringfalse := FuzzyPropValForSearchFromString("false")
// 	stringtrue := FuzzyPropValForSearchFromString("true")
// 	string0 := FuzzyPropValForSearchFromString("0")
// 	string1 := FuzzyPropValForSearchFromString("1")
// 	string240 := FuzzyPropValForSearchFromString("240")
// 	assertPropSame(uint80, int320)
// 	assertPropSame(uint81, int321)
// 	assertPropSame(string0b, uint80)
// 	assertPropSame(string0b, int320)
// 	assertPropSame(string1b, uint81)
// 	assertPropSame(string1b, int321)
// 	assertPropSame(stringfalse, uint80)
// 	assertPropSame(stringfalse, int320)
// 	assertPropSame(stringtrue, uint81)
// 	assertPropSame(stringtrue, int321)
// 	assertPropSame(stringfalse, string0b)
// 	assertPropSame(stringtrue, string1b)
// 	assertPropSame(string0, uint80)
// 	assertPropSame(string0, int320)
// 	assertPropSame(string1, uint81)
// 	assertPropSame(string1, int321)
// 	assertPropSame(string0, string0b)
// 	assertPropSame(string1, string1b)
// 	assertPropSame(stringfalse, string0)
// 	assertPropSame(stringtrue, string1)
// 	assertPropSame(int32240, string240)
// }

type PropsForSearch struct {
	props []struct {
		Name  string
		Value PropValForSearch
	}
	propsByName map[string]PropValForSearch
}

func (ps *PropsForSearch) NumProps() int {
	return len(ps.props)
}

func (ps *PropsForSearch) InPreciseSNBT() string {
	if ps == nil {
		return "{}"
	}
	if ps.props == nil {
		return "{}"
	}
	if len(ps.props) == 0 {
		return "{}"
	}
	props := make([]string, 0, len(ps.props))
	for _, p := range ps.props {
		props = append(props, p.Name+":"+p.Value.InPreciseSNBT())
	}
	stateStr := strings.Join(props, ",")
	return "{" + stateStr + "}"
}

type ComparedOutput struct {
	Same      uint8
	Different uint8
	Redundant uint8
	Missing   uint8
}

func (ps *PropsForSearch) Compare(compare *PropsForSearch) (o ComparedOutput) {
	if ps == nil || compare == nil || compare.propsByName == nil {
		return
	}
	for n, p := range compare.propsByName {
		if actualP, found := ps.propsByName[n]; found {
			if actualP.FuzzyEqual(p) {
				o.Same++
			} else {
				o.Different++
			}
		} else {
			o.Redundant++
		}
	}
	for an := range ps.propsByName {
		if _, found := compare.propsByName[an]; !found {
			o.Missing++
		}
	}
	return o
}

func PropsForSearchFromMap(mapProps map[string]PropValForSearch) *PropsForSearch {
	// clean
	cleanMapProps := map[string]PropValForSearch{}
	for k, v := range mapProps {
		cleanMapProps[strings.ReplaceAll(strings.ReplaceAll(k, "\"", ""), "minecraft:", "")] = v
	}
	mapProps = cleanMapProps
	// sort
	keys := make([]string, 0, len(mapProps))
	for _k := range mapProps {
		k := _k
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	// make
	props := make([]struct {
		Name  string
		Value PropValForSearch
	}, len(mapProps))
	for i, k := range keys {
		props[i].Name = k
		props[i].Value = mapProps[k]
	}
	return &PropsForSearch{
		props:       props,
		propsByName: cleanMapProps,
	}
}

func PropsForSearchFromNbt(nbt map[string]any) (*PropsForSearch, error) {
	mapProps := map[string]PropValForSearch{}
	for _k, _v := range nbt {
		k, v := _k, _v
		switch val := v.(type) {
		case uint8:
			if val > 1 {
				return nil, fmt.Errorf("as an uint8 in prop, value should be 0 or 1, get %v", val)
			}
			mapProps[k] = FuzzyPropValForSearchFromBool(val == 1)
		case bool:
			mapProps[k] = FuzzyPropValForSearchFromBool(val)
		case int32:
			mapProps[k] = FuzzyPropValForSearchFromInt32(val)
		case float64:
			mapProps[k] = FuzzyPropValForSearchFromInt32(int32(val))
		case string:
			mapProps[k] = FuzzyPropValForSearchFromString(strings.ReplaceAll(val, "\"", ""))
		default:
			return nil, fmt.Errorf("as a prop value from nbt, value can only be uint8/int32/string, get %v", v)
		}
	}
	return PropsForSearchFromMap(mapProps), nil
}

func PropsForSearchFromStr(stringProps string) (*PropsForSearch, error) {
	stringProps = strings.TrimSuffix(strings.TrimPrefix(stringProps, "{"), "}")
	stringProps = strings.TrimSuffix(strings.TrimPrefix(stringProps, "["), "]")
	frags := strings.Split(stringProps, ",")
	mapProps := map[string]PropValForSearch{}
	if len(frags) == 0 || len(frags) == 1 && frags[0] == "" {
		return nil, nil
	}
	for _, frag := range frags {
		frag = strings.TrimSpace(frag)
		if len(frag) == 0 {
			continue
		}
		var elems []string
		if strings.Contains(frag, "=") {
			elems = strings.Split(frag, "=")
		} else {
			elems = strings.Split(frag, ":")
		}
		if len(elems) != 2 {
			frag = strings.ReplaceAll(frag, "minecraft:", "")
			if strings.Contains(frag, "=") {
				elems = strings.Split(frag, "=")
			} else {
				elems = strings.Split(frag, ":")
			}
			if len(elems) == 0 {
				continue
			}
			if len(elems) != 2 {
				return nil, fmt.Errorf("prop has no key-value structure: %v", frag)
			}
		}
		key := strings.TrimSpace(elems[0])
		var val PropValForSearch
		value := strings.TrimSpace(elems[1])
		if value == "true" || value == "1b" {
			val = FuzzyPropValForSearchFromBool(true)
		} else if value == "false" || value == "0b" {
			val = FuzzyPropValForSearchFromBool(false)
		} else {
			if int32V, err := strconv.Atoi(value); err == nil {
				val = FuzzyPropValForSearchFromInt32(int32(int32V))
			} else {
				val = FuzzyPropValForSearchFromString(value)
			}
		}
		mapProps[key] = val
	}
	return PropsForSearchFromMap(mapProps), nil
}
