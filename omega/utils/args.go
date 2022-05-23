package utils

import (
	"reflect"
	"strconv"
	"strings"
)

func Compact(cmd string) []string {
	cmds := strings.Split(cmd, " ")
	compactCmds := make([]string, 0)
	for _, cmd := range cmds {
		frag := strings.TrimSpace(cmd)
		if len(frag) > 0 {
			compactCmds = append(compactCmds, frag)
		}
	}
	return compactCmds
}

func SimplePrase(cmds *[]string, argNames []string, data interface{}, firstVal bool) bool {
	v := reflect.ValueOf(data)
	if v.Kind() != reflect.Ptr {
		return false
	}
	// get the value that the pointer v points to.
	v = v.Elem()
	if (firstVal && len(*cmds) < 1) || (!firstVal && len(*cmds) < 2) {
		return false
	}
	ret := make([]string, 0)
	valStr := ""
	i := 0
	c := ""
	flag := false
	for i, c = range *cmds {
		for _, a := range argNames {
			if a == c && i < len(*cmds)-1 {
				flag = true
				break
			}
		}
		if flag {
			break
		}
	}
	if flag {
		valStr = (*cmds)[i+1]
		for _i, s := range *cmds {
			if _i < i || _i > i+1 {
				ret = append(ret, s)
			}
		}
	} else if firstVal {
		valStr = (*cmds)[0]
		ret = (*cmds)[1:]
	} else {
		return false
	}
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		realV, err := strconv.Atoi(valStr)
		if err != nil {
			return false
		}
		if v.OverflowInt(int64(realV)) {
			return false
		}
		v.SetInt(int64(realV))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		realV, err := strconv.Atoi(valStr)
		if err != nil || realV < 0 {
			return false
		}
		if v.OverflowUint(uint64(realV)) {
			return false
		}
		v.SetUint(uint64(realV))
	case reflect.Float32, reflect.Float64:
		realV, err := strconv.ParseFloat(valStr, 64)
		if err != nil {
			return false
		}
		if v.OverflowFloat(realV) {
			return false
		}
		v.SetFloat(realV)
	case reflect.String:
		v.SetString(valStr)
	default:
		return false
	}
	*cmds = ret
	return true
}
