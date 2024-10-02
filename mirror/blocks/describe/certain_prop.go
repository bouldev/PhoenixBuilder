package describe

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"phoenixbuilder/fastbuilder/alter/snbt"
)

type PropValType uint8

// mc 使用的属性值有三种
// 1. uint8 但是只使用 true/false (0,1) 两种值, bedrock state str: true/false, snbt 0b/1b
const PropValTypeUint8 = PropValType(1)

// 2. int32 bedrock state str, snbt = 0,1,2,3....
const PropValTypeInt32 = PropValType(2)

// 3. string bedrock state str bedrock state str, snbt = "some value"
const PropValTypeString = PropValType(4)

type PropVal interface {
	// 对于 PropVal, 其 type 只为 PropValUint8/PropValInt32/PropValString 中的一种
	Type() PropValType
	// 返回结果为 (Type() & val)!=0
	HasType(val PropValType) bool
	// 若不为 PropValUint8 则 panic
	Uint8Val() uint8
	// 若不为 PropValInt32 则 panic
	Int32Val() int32
	// 若不为 PropValString 则 panic
	StringVal() string
	// uint8 -> true/false; int32 -> 1,2,3,...; string -> "string"
	BedrockString() string
	// SNBT -> 1b/0b; int32 -> 1,2,3,...; string -> "string"
	SNBTString() string

	Raw() any

	Equal(other PropVal) bool
}

type PropValUint8 bool

func (v PropValUint8) Raw() any {
	if v {
		return uint8(1)
	}
	return uint8(0)
}

func (v PropValUint8) Type() PropValType {
	return PropValTypeUint8
}

func (v PropValUint8) HasType(val PropValType) bool {
	return (v.Type() & val) != 0
}

func (v PropValUint8) Uint8Val() uint8 {
	if v {
		return uint8(1)
	}
	return uint8(0)
}

func (v PropValUint8) Int32Val() int32 {
	panic("not int32")
}

func (v PropValUint8) StringVal() string {
	panic("not string")
}

func (v PropValUint8) BedrockString() string {
	if v {
		return "true"
	} else {
		return "false"
	}
}

func (v PropValUint8) SNBTString() string {
	if v {
		return "1b"
	} else {
		return "0b"
	}
}

func (v PropValUint8) Equal(val2 PropVal) bool {
	if v.HasType(val2.Type()) {
		return v.Uint8Val() == val2.Uint8Val()
	}
	return false
}

var PropVal0 = PropValUint8(false)
var PropVal1 = PropValUint8(true)

func propValFromBool(b bool) PropVal {
	if b {
		return PropVal1
	}
	return PropVal0
}

type PropValInt32 int32

func (v PropValInt32) Raw() any {
	return int32(v)
}

func (v PropValInt32) Type() PropValType {
	return PropValTypeInt32
}

func (v PropValInt32) HasType(val PropValType) bool {
	return (v.Type() & val) != 0
}

func (v PropValInt32) Uint8Val() uint8 {
	panic("not uint8")
}

func (v PropValInt32) Int32Val() int32 {
	return int32(v)
}

func (v PropValInt32) StringVal() string {
	panic("not string")
}

func (v PropValInt32) BedrockString() string {
	return strconv.Itoa(int(v))
}

func (v PropValInt32) SNBTString() string {
	return strconv.Itoa(int(v))
}

func (v PropValInt32) Equal(val2 PropVal) bool {
	if v.HasType(val2.Type()) {
		return v.Int32Val() == val2.Int32Val()
	}
	return false
}

func PropValFromInt32(val int32) PropVal {
	return PropValInt32(val)
}

type PropValString string

func (v PropValString) Raw() any {
	return string(v)
}

func (v PropValString) Type() PropValType {
	return PropValTypeString
}

func (v PropValString) HasType(val PropValType) bool {
	return (v.Type() & val) != 0
}

func (v PropValString) Uint8Val() uint8 {
	panic("not uint8")
}

func (v PropValString) Int32Val() int32 {
	panic("not int32")
}

func (v PropValString) StringVal() string {
	return string(v)
}

func (v PropValString) BedrockString() string {
	return "\"" + string(v) + "\""
}

func (v PropValString) SNBTString() string {
	return "\"" + string(v) + "\""
}

func (v PropValString) Equal(val2 PropVal) bool {
	if v.HasType(val2.Type()) {
		return v.StringVal() == val2.StringVal()
	}
	return false
}

func PropValFromString(val string) PropVal {
	val = strings.TrimPrefix(strings.TrimPrefix(val, "\""), "\"")
	return PropValString(val)
}

type Props []struct {
	Name  string
	Value PropVal
}

func (ps Props) ToNBT() map[string]any {
	nbt := map[string]any{}
	for _, p := range ps {
		nbt[p.Name] = p.Value.Raw()
	}
	return nbt
}

func (ps Props) BedrockString(bracket bool) string {
	if len(ps) == 0 {
		if bracket {
			return "[]"
		}
		return ""
	}
	props := make([]string, 0, len(ps))
	for _, p := range ps {
		props = append(props, "\""+p.Name+"\""+"="+p.Value.BedrockString())
	}
	stateStr := strings.Join(props, ",")
	if !bracket {
		return stateStr
	}
	return "[" + stateStr + "]"
}

func (ps Props) SNBTString() string {
	if len(ps) == 0 {
		return "{}"
	}
	props := make([]string, 0, len(ps))
	for _, p := range ps {
		props = append(props, p.Name+":"+p.Value.SNBTString())
	}
	stateStr := strings.Join(props, ",")
	return "{" + stateStr + "}"
}

func PropsFromMap(mapProps map[string]PropVal) Props {
	// clean
	cleanMapProps := map[string]PropVal{}
	for k, v := range mapProps {
		cleanMapProps[strings.ReplaceAll(k, "\"", "")] = v
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
		Value PropVal
	}, len(mapProps))
	for i, k := range keys {
		props[i].Name = k
		props[i].Value = mapProps[k]
	}
	return props
}

func PropsFromNbt(nbt map[string]any) Props {
	mapProps := map[string]PropVal{}
	for _k, _v := range nbt {
		k, v := _k, _v
		switch val := v.(type) {
		case uint8:
			if val > 1 {
				panic(fmt.Errorf("as an uint8 in prop, value should be 0 or 1, get %v", val))
			}
			mapProps[k] = propValFromBool(val == 1)
		case int32:
			mapProps[k] = PropValFromInt32(val)
		case string:
			mapProps[k] = PropValFromString(strings.ReplaceAll(val, "\"", ""))
		default:
			panic(fmt.Errorf("as a prop value from nbt, value can only be uint8/int32/string, get %v", v))
		}
	}
	return PropsFromMap(mapProps)
}

func PropsFromSNBT(snbtProps string) Props {
	nbt, err := snbt.SNBToNbt(snbtProps)
	if err != nil {
		panic(nbt)
	}
	nbtMap, ok := nbt.(map[string]any)
	if !ok {
		panic(nbt)
	}
	return PropsFromNbt(nbtMap)
}
