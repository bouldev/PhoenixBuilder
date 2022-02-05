package nbtconstructor

import (
	"phoenixbuilder/fastbuilder/function"
	"phoenixbuilder/fastbuilder/command"
	"phoenixbuilder/minecraft"
	"encoding/base64"
	"encoding/json"
	"encoding/hex"
	"strings"
	"reflect"
	"fmt"
	"io/ioutil"
)

const (
	TypeNull uint8 = iota
	TypeVoid
	TypeUInt8
	TypeUInt16
	TypeUInt32
	TypeUInt64
	TypeChar
	TypeShort
	TypeInt
	TypeLong
	TypeBase64Bytes
	TypeHexBytes
	TypeInt32Array
	TypeInt64Array
	TypeFloat
	TypeDouble
)

func parseToRegularType(item interface{}, typeDeclaration string, isRoot bool) (interface{}, error) {
	if(item==nil) {
		return nil, fmt.Errorf("Null value is not allowed in NBT. (%#v)", item)
	}
	itemType:=reflect.TypeOf(item).Kind()
	if(itemType==reflect.Invalid) {
		return nil, fmt.Errorf("Invalid type for item: (%#v)",item)
	}
	if(itemType==reflect.Slice) {
		content:=item.([]interface{})
		out:=make([]interface{},len(content))
		for i, c := range content {
			t,e:=parseToRegularType(c, typeDeclaration, false)
			if e!=nil {
				return nil,e
			}
			out[i]=t
		}
		return out, nil
	}
	if(itemType==reflect.Map) {
		mapitem:=item.(map[string]interface{})
		r,err:=parseJsonNBT(mapitem, false)
		return r,err
	}
	if(itemType==reflect.String&&len(typeDeclaration)==0) {
		typeDeclaration="char *"
	}
	isPtr:=false
	typeDeclaration=strings.Replace(typeDeclaration,"\t"," ",-1)
	typeDeclaration=strings.Replace(typeDeclaration,"\n"," ",-1)
	typeDeclaration=strings.Replace(typeDeclaration,"\r"," ",-1)
	if len(typeDeclaration) != 0 {
		for {
			if(typeDeclaration[len(typeDeclaration)-1]==' ') {
				typeDeclaration=typeDeclaration[:len(typeDeclaration)-1]
				continue
			}
			break
		}
		for {
			if(typeDeclaration[0]==' ') {
				typeDeclaration=typeDeclaration[1:]
				continue
			}
			break
		}
		if(typeDeclaration[len(typeDeclaration)-1]=='*') {
			isPtr=true
			typeDeclaration=typeDeclaration[:len(typeDeclaration)-1]
		}
		for {
			if(typeDeclaration[len(typeDeclaration)-1]==' ') {
				typeDeclaration=typeDeclaration[:len(typeDeclaration)-1]
				continue
			}
			break
		}
		for {
			if(typeDeclaration[0]==' ') {
				typeDeclaration=typeDeclaration[1:]
				continue
			}
			break
		}
	}
	for {
		r:=strings.Replace(typeDeclaration, "  "," ",-1)
		if r != typeDeclaration {
			typeDeclaration=r
			continue
		}
		break
	}
	if(typeDeclaration==" ") {
		typeDeclaration=""
	}
	if itemType==reflect.Bool {
		boolVal:=item.(bool)
		if len(typeDeclaration)==0 {
			if boolVal {
				return byte(1), nil
			}
			return byte(0), nil
		}
		if boolVal {
			item=float64(1)
		}else{
			item=float64(0)
		}
		itemType=reflect.Float64
	}
	switch typeDeclaration {
	case "uint8_t":
		fallthrough
	case "unsigned char":
		if isPtr {
			return nil, fmt.Errorf("unsigned char *: Type not assigned")
		}
		if itemType!=reflect.Float64 {
			return nil, fmt.Errorf("unsigned char: Expected number, but got (%#v).",item)
		}
		numVal, succ:=item.(float64)
		if !succ {
			return nil, fmt.Errorf("unsigned char: Expected a number value.")
		}
		if numVal>255{
			return nil, fmt.Errorf("unsigned char: Expected value 0~255, but got %d.",int(numVal))
		}
		return byte(uint8(numVal)), nil
	case "int8_t":
	case "char":
		if isPtr {
			if itemType!=reflect.String {
				return nil, fmt.Errorf("char *:Expected string, but got (%#v).",item)
			}
			return item, nil
		}
		if itemType!=reflect.Float64 {
			return nil, fmt.Errorf("char: Expected number, but got (%#v).",item)
		}
		numVal, succ:=item.(float64)
		if !succ {
			return nil, fmt.Errorf("char: Expected a number value.")
		}
		if numVal>127 || numVal< -128 {
			return nil, fmt.Errorf("char: Expected value -128 ~ 127, but got %d.",int(numVal))
		}
		return byte(int8(numVal)), nil
	case "uint16_t":
		fallthrough
	case "unsigned short":
		if !isRoot {
			return nil, fmt.Errorf("unsigned short is only allowed in the root component.")
		}
		if isPtr {
			return nil, fmt.Errorf("unsigned short * is not allowed.")
		}
		numVal, succ:=item.(float64)
		if !succ {
			return nil, fmt.Errorf("unsigned short: Expected a number value.")
		}
		if numVal>65535 {
			return nil, fmt.Errorf("unsigned short: Expected value 0 ~ 65535, but got %d.",int(numVal))
		}
		return uint16(numVal), nil
	case "int16_t":
		fallthrough
	case "short":
		if isPtr {
			return nil, fmt.Errorf("short * is not allowed.")
		}
		numVal, succ:=item.(float64)
		if !succ {
			return nil, fmt.Errorf("short: Expected a number value.")
		}
		if numVal>32767 || numVal < -32768 {
			return nil, fmt.Errorf("short: Expected value -32768 ~ 32767, but got %d.",int(numVal))
		}
		return int16(numVal), nil
	case "int32_t":
		fallthrough
	case "int":
		if isPtr {
			gotArr, succ:=item.([]interface{})
			if !succ {
				return nil, fmt.Errorf("int *: Not an int array.(%#v)",item)
			}
			intArr:=make([]int32, len(gotArr))
			for d, ci := range gotArr {
				ii, succ:=ci.(float64)
				if !succ {
					return nil, fmt.Errorf("int * must contains numbers only.(%#v)",item)
				}
				if ii>2147483647||ii< -2147483648 {
					return nil, fmt.Errorf("int */int: Expected -2147483648 ~ 2147483647, but got %d.",int(ii))
				}
				intArr[d]=int32(ii)
			}
			return intArr[:], nil
		}
		ii, succ:=item.(float64)
		if !succ {
			return nil, fmt.Errorf("int: Not a number value.(%#v)",item)
		}
		if ii>2147483647||ii< -2147483648 {
			return nil, fmt.Errorf("int: Expected -2147483648 ~ 2147483647, but got %d.",int(ii))
		}
		return int32(ii), nil
	case "int64_t":
		fallthrough
	case "long":
		if isPtr {
			gotArr, succ:=item.([]interface{})
			if !succ {
				return nil, fmt.Errorf("long *: Not a long array.(%#v)",item)
			}
			intArr:=make([]int64, len(gotArr))
			for d, ci := range gotArr {
				ii, succ:=ci.(float64)
				if !succ {
					return nil, fmt.Errorf("long * must contains numbers only.(%#v)",item)
				}
				if ii>9223372036854775807||ii< -9223372036854775808 {
					return nil, fmt.Errorf("long */long: Expected -9223372036854775808 ~ 9223372036854775807, but got %d.",int(ii))
				}
				intArr[d]=int64(ii)
			}
			return intArr[:], nil
		}
		ii, succ:=item.(float64)
		if !succ {
			return nil, fmt.Errorf("long: Not a number value.(%#v)",item)
		}
		if ii>9223372036854775807||ii< -9223372036854775808 {
			return nil, fmt.Errorf("long: Expected -9223372036854775808 ~ 9223372036854775807, but got %d.",int(ii))
		}
		return int64(ii), nil
	case "float":
		if isPtr {
			return nil, fmt.Errorf("float * is not allowed.")
		}
		nv, succ:=item.(float64)
		if !succ {
			return nil, fmt.Errorf("float: Expected a numberic value.")
		}
		return float32(nv), nil
	case "double":
		nv, succ:=item.(float64)
		if !succ {
			return nil, fmt.Errorf("double: Expected a numberic value.")
		}
		return nv, nil
	case "__hex uint8_t":
		fallthrough
	case "unsigned __hex char":
		fallthrough
	case "__hex unsigned char":
		if !isPtr {
			return nil, fmt.Errorf("unsigned char isn't allowed.")
		}
		if itemType!=reflect.String {
			return nil, fmt.Errorf("__hex unsigned char * requires string value contains hex.")
		}
		str:=item.(string)
		dst:=make([]byte, hex.DecodedLen(len(str)))
		_, err:=hex.Decode(dst, []byte(str))
		if err != nil {
			return nil, fmt.Errorf("__hex unsigned char *: failed to decode hex string: %v.(%#v)",err,item)
		}
		return dst[:], nil
	case "__b64 uint8_t":
		fallthrough
	case "__base64 uint8_t":
		fallthrough
	case "unsigned __b64 char":
		fallthrough
	case "unsigned __base64 char":
		fallthrough
	case "__b64 unsigned char":
		fallthrough
	case "__base64 unsigned char":
		if !isPtr {
			return nil, fmt.Errorf("unsigned char isn't allowed.")
		}
		if itemType!=reflect.String {
			return nil, fmt.Errorf("__hex unsigned char * requires string value contains hex.")
		}
		str:=item.(string)
		dst, err:=base64.StdEncoding.DecodeString(str)
		if err != nil {
			return nil, fmt.Errorf("__base64 unsigned char *: failed to decode base64 string: %v.(%#v)",err,item)
		}
		return dst[:], nil
	default:
		return nil, fmt.Errorf("Unknown type declaration: %v.",typeDeclaration)
	}
	return nil, nil
}

func parseJsonNBT(st map[string]interface{}, isRoot bool) (map[string]interface{}, error) {
	out:=make(map[string]interface{})
	for key, content := range st {
		ka:=strings.Split(key, ":")
		realKey:=key
		declaration:=""
		if len(ka)!=1 {
			declaration=ka[len(ka)-1]
			realKey=strings.Join(ka[:len(ka)-1], ":")
		}
		a,b:=parseToRegularType(content, declaration, isRoot)
		if b!=nil {
			return nil, b
		}
		out[realKey]=a
	}
	return out,nil
}

func evalNBT(content string,conn *minecraft.Conn) {
	parsedstruct:=make(map[string]interface{})
	err:=json.Unmarshal([]byte(content),&parsedstruct)
	if(err!=nil) {
		command.Tellraw(conn, fmt.Sprintf("Invalid json context, %v",err))
		return
	}
	m,err:=parseJsonNBT(parsedstruct, true)
	if err != nil {
		command.Tellraw(conn, fmt.Sprintf("%v",err))
		return
	}
	itemName,found:=m["name"]
	if !found {
		command.Tellraw(conn, "Item name not assigned")
		return
	}
	itemNameStr, converted:=itemName.(string)
	if !converted {
		command.Tellraw(conn, "Type of item name should be string.")
		return
	}
	item,hasItem:=ItemMap[itemNameStr]
	if !hasItem {
		command.Tellraw(conn, "The item with the assigned name doesn't exist.")
		return
	}
	data,hasData:=m["data"]
	if !hasData {
		data=uint16(0)
	}
	dataVal, converted:=data.(uint16)
	if !converted {
		command.Tellraw(conn, "The item's data should be an uint16_t value.")
		return
	}
	/*if dataVal>item.MaxDamage {
		command.Tellraw(conn, fmt.Sprintf("Invalid data value %d, the max damage value for the item %s is %d.",dataVal,itemNameStr,item.MaxDamage))
		return
	}*/
	nbtcontent,hasNBT:=m["nbt"]
	if !hasNBT {
		nbtcontent=map[string]interface{} {}
	}
	nbtmap,converted:=nbtcontent.(map[string]interface{})
	if !converted {
		command.Tellraw(conn, "The \"nbt\" should be an object.")
		return
	}
	go func() {
		StartSessionWithCustomNBT(conn, int32(item.NetworkID), uint32(dataVal), nbtmap)
	} ()
	return
}

func InitNBTConstructor() {
	function.RegisterFunction(&function.Function {
		Name: "constructItem_simple",
		OwnedKeywords: []string{"simpleconstruct"},
		FunctionType: function.FunctionTypeSimple,
		SFArgumentTypes: []byte { function.SimpleFunctionArgumentMessage },
		SFMinSliceLen: 1,
		FunctionContent: func(conn *minecraft.Conn, args []interface{}) {
			toEval:=args[0].(string)
			evalNBT(toEval, conn)
		},
	})
	function.RegisterFunction(&function.Function {
		Name: "constructItem",
		OwnedKeywords: []string{"construct"},
		FunctionType: function.FunctionTypeSimple,
		SFArgumentTypes: []byte { function.SimpleFunctionArgumentMessage },
		SFMinSliceLen: 1,
		FunctionContent: func(conn *minecraft.Conn, args []interface{}) {
			path:=args[0].(string)
			content, err:=ioutil.ReadFile(path)
			if err != nil {
				command.Tellraw(conn, fmt.Sprintf("Error: %v",err))
				return
			}
			evalNBT(string(content), conn)
		},
	})
}