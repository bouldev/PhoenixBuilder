package collaborate

import (
	"fmt"
	"phoenixbuilder/omega/defines"
	"regexp"
	"strconv"
	"strings"
)

const (
	INTERFACE_GEN_STRING_LIST_HINT_RESOLVER            = "INTERFACE_GEN_STRING_LIST_HINT_RESOLVER"
	INTERFACE_GEN_STRING_LIST_HINT_RESOLVER_WITH_INDEX = "INTERFACE_GEN_STRING_LIST_HINT_RESOLVER_WITH_INDEX"
	INTERFACE_GEN_INT_RANGE_RESOLVER                   = "INTERFACE_GEN_INT_RANGE_RESOLVER"
	INTERFACE_GEN_YES_NO_RESOLVER                      = "INTERFACE_GEN_YES_NO_RESOLVER"
	INTERFACE_QUERY_FOR_PLAYER_NAME                    = "INTERFACE_QUERY_FOR_PLAYER_NAME"
)

type GEN_STRING_LIST_HINT_RESOLVER func(available []string) (string, func(params []string) (selection int, cancel bool, err error))

type GEN_STRING_LIST_HINT_RESOLVER_WITH_INDEX func(_available []string) (string, func(params []string) (selection int, cancel bool, err error))

type GEN_INT_RANGE_RESOLVER func(min int, max int) (string, func(params []string) (selection int, cancel bool, err error))

type GEN_YES_NO_RESOLVER func() (string, func(params []string) (bool, error))

type QUERY_FOR_PLAYER_NAME func(src string, dst string, searchFn FUNCTYPE_GET_POSSIBLE_NAME) (name string, cancel bool)

// 为确保始终可用，会由主框架将以下函数设置为默认Context
func GenStringListHintResolver(available []string) (string, func(params []string) (selection int, cancel bool, err error)) {
	hint := "[ " + strings.Join(available, ", ") + ", 取消] 之一"
	resolver := func(params []string) (int, bool, error) {
		if len(params) == 0 {
			return 0, false, fmt.Errorf("为空")
		}
		p := params[0]
		if p == "取消" {
			return 0, true, nil
		}
		for i, _p := range available {
			if _p == p {
				return i, false, nil
			}
		}
		return 0, false, fmt.Errorf("不在选项范围内")
	}
	return hint, resolver
}

func GenStringListHintResolverWithIndex(_available []string) (string, func(params []string) (selection int, cancel bool, err error)) {
	raw_available := make([]string, len(_available))
	available := make([]string, len(_available))
	if len(_available) == 0 {
		return "[没有可选项]", func(params []string) (int, bool, error) {
			return 0, false, fmt.Errorf("[没有可选项]")
		}
	}
	for i, m := range _available {
		available[i] = fmt.Sprintf("%v:%v", i+1, m)
		raw_available[i] = m
	}
	intHint, intResolver := GenIntRangeResolver(1, len(available))

	hint := "[ " + strings.Join(available, ", ") + ", 取消] 之一,或者" + intHint
	resolver := func(params []string) (int, bool, error) {
		if len(params) == 0 {
			return 0, false, fmt.Errorf("为空")
		}
		p := params[0]
		if p == "取消" {
			return 0, true, nil
		}
		for i, _p := range raw_available {
			if _p == p {
				return i, false, nil
			}
		}

		resolver, cancel, err := intResolver(params)
		if cancel {
			return 0, true, nil
		}
		if err != nil {
			return 0, false, fmt.Errorf("不在可选范围内")
		}
		return resolver - 1, false, nil
	}
	return hint, resolver
}

func GenIntRangeResolver(min int, max int) (string, func(params []string) (selection int, cancel bool, err error)) {
	re := regexp.MustCompile("^[-]?[0-9]+")
	hint := fmt.Sprintf("%v~%v之间的整数", min, max)
	resolver := func(params []string) (int, bool, error) {
		if len(params) == 0 {
			return 0, false, fmt.Errorf("为空")
		}
		val := re.FindAllString(params[0], 1)
		if params[0] == "取消" {
			return 0, true, nil
		}
		if len(val) == 1 {
			v, _ := strconv.Atoi(val[0])
			if v >= min && v <= max {
				return v, false, nil
			}
			return 0, false, fmt.Errorf("不在范围中")
		} else {
			return 0, false, fmt.Errorf("不是一个有效的数")
		}
	}
	return hint, resolver
}

func GenYesNoResolver() (string, func(params []string) (bool, error)) {
	hint := "输入 是/y 同意; 否/n 拒绝"
	resolver := func(params []string) (bool, error) {
		if len(params) == 0 {
			return false, fmt.Errorf("为空")
		}
		p := params[0]
		if strings.HasPrefix(p, "是") || strings.HasPrefix(p, "Y") || strings.HasPrefix(p, "y") {
			return true, nil
		} else if strings.HasPrefix(p, "否") || strings.HasPrefix(p, "N") || strings.HasPrefix(p, "n") {
			return false, nil
		}
		return false, fmt.Errorf("不是有效的回答")
	}
	return hint, resolver
}

func QueryForPlayerName(ctrl defines.GameControl, src string, dst string, searchFn FUNCTYPE_GET_POSSIBLE_NAME) (name string, cancel bool) {
	termCtx := make(chan bool)
	var hint []string
	var resolver func(params []string) (int, bool, error)
	resolver = nil
	candidateNames := []string{}
	for {
		if dst == "" {
			hint = []string{"请输入目标玩家名,或者目标玩家名的一部分(或输入: 取消 )"}
			resolver = nil
		} else {
			candidateNames = []string{}
			possibleNames := searchFn(dst, 3)
			//fmt.Println(possibleNames)
			if len(possibleNames) > 0 && possibleNames[0].Entry.CurrentName == dst {
				return dst, false
			}
			hint = []string{"请选择下方的序号，或者目标玩家名(或输入: 取消 )"}
			for i, name := range possibleNames {
				candidateNames = append(candidateNames, name.Entry.CurrentName)
				currentName, historyName := name.GenReadAbleStringPair()
				hint = append(hint, fmt.Sprintf("%v: %v %v", i+1, currentName, historyName))
			}
			if len(candidateNames) > 0 {
				_, resolver = GenStringListHintResolverWithIndex(candidateNames)
			} else {
				hint = []string{"没有搜索到匹配的玩家，请输入目标玩家名,或者目标玩家名的一部分(或输入: 取消 )"}
				resolver = nil
			}
		}
		if ctrl.SetOnParamMsg(src, func(chat *defines.GameChat) (catch bool) {
			if len(chat.Msg) == 0 || chat.Msg[0] == "取消" {
				termCtx <- true
				return
			} else {
				if resolver == nil {
					dst = chat.Msg[0]
					termCtx <- false
					return
				}
				selection, cancel, err := resolver(chat.Msg)
				if err != nil {
					termCtx <- false
					dst = chat.Msg[0]
					return
				}
				if cancel {
					termCtx <- true
					return
				}
				name = candidateNames[selection]
				termCtx <- false
			}
			return true
		}) == nil {
			for _, h := range hint {
				ctrl.SayTo(src, h)
			}
		} else {
			return "", true
		}
		c := <-termCtx
		if name != "" {
			return name, false
		}
		if c {
			return "", true
		}
	}
}
