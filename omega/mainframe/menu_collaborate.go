package mainframe

import (
	"fmt"
	"phoenixbuilder/omega/collaborate"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"regexp"
	"strconv"
	"strings"
)

// GenStringListHintResolverSettings
type GSLHRSettings struct {
	FrameText  string `json:"提示样式"`
	JoinText   string `json:"选项间隔符"`
	CancelText string `json:"取消选择触发词"`
	EmptyText  string `json:"选项为空时的提示"`
	ErrorText  string `json:"不在选项范围内的提示"`
}

// GenStringListHintResolverWithIndexSettings
type GSLHRWISettings struct {
	FrameText  string `json:"提示样式"`
	ItemText   string `json:"选项样式"`
	JoinText   string `json:"选项间隔符"`
	CancelText string `json:"取消选择触发词"`
	EmptyText  string `json:"选项为空时的提示"`
	ErrorText  string `json:"不在选项范围内的提示"`
}

// GenIntRangeResolverSettings
type GIRRSettings struct {
	FrameText      string `json:"提示样式"`
	CancelText     string `json:"取消选择触发词"`
	EmptyText      string `json:"选项为空时的提示"`
	OutOfRangeText string `json:"不在选项范围内的提示"`
	ErrorText      string `json:"输入无效时的提示"`
}

// GenYesNoResolverSettings
type GYNSettings struct {
	FrameText string `json:"提示样式"`
	EmptyText string `json:"选项为空时的提示"`
	ErrorText string `json:"输入无效时的提示"`
}

// QueryForPlayerNameSettings
type QFPNSettings struct {
	ItemText          string `json:"选项样式"`
	CancelText        string `json:"取消选择触发词"`
	RequireInputText  string `json:"要求输入玩家名时提示"`
	RequireSelectText string `json:"要求选择玩家名时提示"`
	SearchFailText    string `json:"找不到玩家时提示"`
}

// From PhoenixBuilder/omega/utils/params.go
func (m *Menu) GenStringListHintResolver(available []string) (string, func(params []string) (selection int, cancel bool, err error)) {
	options := available
	if m.GSLHRSettings.CancelText != "" {
		options = append(options, m.GSLHRSettings.CancelText)
	}
	hint := utils.FormatByReplacingOccurrences(m.GSLHRSettings.FrameText, map[string]interface{}{
		"[options]": strings.Join(options, m.GSLHRSettings.JoinText),
	})
	resolver := func(params []string) (int, bool, error) {
		if len(params) == 0 {
			return 0, false, fmt.Errorf(m.GSLHRSettings.EmptyText)
		}
		p := params[0]
		if p == m.GSLHRSettings.CancelText {
			return 0, true, nil
		}
		for i, _p := range available {
			if _p == p {
				return i, false, nil
			}
		}
		return 0, false, fmt.Errorf(m.GSLHRSettings.ErrorText)
	}
	return hint, resolver
}

// From PhoenixBuilder/omega/utils/params.go
func (m *Menu) GenStringListHintResolverWithIndex(_available []string) (string, func(params []string) (selection int, cancel bool, err error)) {
	if len(_available) == 0 {
		return m.GSLHRWISettings.EmptyText, func(params []string) (int, bool, error) {
			return 0, false, fmt.Errorf(m.GSLHRWISettings.EmptyText)
		}
	}
	raw_available := make([]string, len(_available)+1)
	available := make([]string, len(_available)+1)
	if m.GSLHRWISettings.CancelText != "" {
		_available = append(_available, m.GSLHRWISettings.CancelText)
	}
	itemText := m.GSLHRWISettings.ItemText
	for i, m := range _available {
		available[i] = utils.FormatByReplacingOccurrences(itemText, map[string]interface{}{
			"[i]":      i + 1,
			"[option]": m,
		})
		raw_available[i] = m
	}
	intHint, intResolver := m.GenIntRangeResolver(1, len(available))
	hint := utils.FormatByReplacingOccurrences(m.GSLHRWISettings.FrameText, map[string]interface{}{
		"[options]": strings.Join(available, m.GSLHRWISettings.JoinText),
		"[intHint]": intHint,
	})
	resolver := func(params []string) (int, bool, error) {
		if len(params) == 0 {
			return 0, false, fmt.Errorf(m.GSLHRWISettings.EmptyText)
		}
		p := params[0]
		if p == m.GSLHRWISettings.CancelText {
			return 0, true, nil
		}
		for i, _p := range raw_available {
			if _p == p {
				return i, false, nil
			}
		}
		resolver, cancel, err := intResolver(params)
		if resolver == len(available) || cancel {
			return 0, true, nil
		}
		if err != nil {
			return 0, false, fmt.Errorf(m.GSLHRWISettings.ErrorText)
		}
		return resolver - 1, false, nil
	}
	return hint, resolver
}

// From PhoenixBuilder/omega/utils/params.go
func (m *Menu) GenIntRangeResolver(min int, max int) (string, func(params []string) (selection int, cancel bool, err error)) {
	hint := utils.FormatByReplacingOccurrences(m.GIRRSettings.FrameText, map[string]interface{}{
		"[min]": min,
		"[max]": max,
	})
	resolver := func(params []string) (int, bool, error) {
		if len(params) == 0 {
			return 0, false, fmt.Errorf(m.GIRRSettings.EmptyText)
		}
		val := regexp.MustCompile("^[-]?[0-9]+").FindAllString(params[0], 1)
		if params[0] == m.GIRRSettings.CancelText {
			return 0, true, nil
		}
		if len(val) == 1 {
			v, _ := strconv.Atoi(val[0])
			if v >= min && v <= max {
				return v, false, nil
			}
			return 0, false, fmt.Errorf(m.GIRRSettings.OutOfRangeText)
		} else {
			return 0, false, fmt.Errorf(m.GIRRSettings.ErrorText)
		}
	}
	return hint, resolver
}

// From PhoenixBuilder/omega/utils/params.go
func (m *Menu) GenYesNoResolver() (string, func(params []string) (bool, error)) {
	resolver := func(params []string) (bool, error) {
		if len(params) == 0 {
			return false, fmt.Errorf(m.GYNSettings.EmptyText)
		}
		p := params[0]
		if strings.HasPrefix(p, "是") || strings.HasPrefix(p, "Y") || strings.HasPrefix(p, "y") {
			return true, nil
		} else if strings.HasPrefix(p, "否") || strings.HasPrefix(p, "N") || strings.HasPrefix(p, "n") {
			return false, nil
		}
		return false, fmt.Errorf(m.GYNSettings.ErrorText)
	}
	return m.GYNSettings.FrameText, resolver
}

// From PhoenixBuilder/omega/utils/steps.go
func (m *Menu) QueryForPlayerName(src string, dst string, searchFn collaborate.FUNCTYPE_GET_POSSIBLE_NAME) (name string, cancel bool) {
	termCtx := make(chan bool)
	var hint []string
	var resolver func(params []string) (int, bool, error)
	resolver = nil
	candidateNames := []string{}
	for {
		if dst == "" {
			hint = []string{utils.FormatByReplacingOccurrences(m.QFPNSettings.RequireInputText, map[string]interface{}{
				"[cancel]": m.QFPNSettings.CancelText,
			})}
			resolver = nil
		} else {
			candidateNames = []string{}
			possibleNames := searchFn(dst, 3)
			//fmt.Println(possibleNames)
			if len(possibleNames) > 0 && possibleNames[0].Entry.CurrentName == dst {
				return dst, false
			}
			hint = []string{utils.FormatByReplacingOccurrences(m.QFPNSettings.RequireSelectText, map[string]interface{}{
				"[cancel]": m.QFPNSettings.CancelText,
			})}
			for i, name := range possibleNames {
				candidateNames = append(candidateNames, name.Entry.CurrentName)
				currentName, historyName := name.GenReadAbleStringPair()
				hint = append(hint, utils.FormatByReplacingOccurrences(m.QFPNSettings.ItemText, map[string]interface{}{
					"[i]":           i + 1,
					"[currentName]": currentName,
					"[historyName]": historyName,
				}))
			}
			if len(candidateNames) > 0 {
				hint = append(hint, utils.FormatByReplacingOccurrences(m.QFPNSettings.ItemText, map[string]interface{}{
					"[i]":           len(possibleNames) + 1,
					"[currentName]": m.QFPNSettings.CancelText,
					"[historyName]": "",
				}))
				_, resolver = m.GenStringListHintResolverWithIndex(candidateNames)
			} else {
				hint = []string{utils.FormatByReplacingOccurrences(m.QFPNSettings.SearchFailText, map[string]interface{}{
					"[cancel]": m.QFPNSettings.CancelText,
				})}
				resolver = nil
			}
		}
		if m.mainFrame.GetGameControl().SetOnParamMsg(src, func(chat *defines.GameChat) (catch bool) {
			if len(chat.Msg) == 0 || chat.Msg[0] == m.QFPNSettings.CancelText {
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
				m.mainFrame.GetGameControl().SayTo(src, h)
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

// 将上面的函数通过框架共享给其他组件使用
func (m *Menu) SetCollaborateFunc() {
	var func1 collaborate.GEN_STRING_LIST_HINT_RESOLVER = func(available []string) (string, func(params []string) (selection int, cancel bool, err error)) {
		return m.GenStringListHintResolver(available)
	}
	var func2 collaborate.GEN_STRING_LIST_HINT_RESOLVER_WITH_INDEX = func(_available []string) (string, func(params []string) (selection int, cancel bool, err error)) {
		return m.GenStringListHintResolverWithIndex(_available)
	}
	var func3 collaborate.GEN_INT_RANGE_RESOLVER = func(min, max int) (string, func(params []string) (selection int, cancel bool, err error)) {
		return m.GenIntRangeResolver(min, max)
	}
	var func4 collaborate.GEN_YES_NO_RESOLVER = func() (string, func(params []string) (bool, error)) {
		return m.GenYesNoResolver()
	}
	var func5 collaborate.QUERY_FOR_PLAYER_NAME = func(src, dst string, searchFn collaborate.FUNCTYPE_GET_POSSIBLE_NAME) (name string, cancel bool) {
		return m.QueryForPlayerName(src, dst, searchFn)
	}
	m.mainFrame.SetContext(collaborate.INTERFACE_GEN_STRING_LIST_HINT_RESOLVER, func1)
	m.mainFrame.SetContext(collaborate.INTERFACE_GEN_STRING_LIST_HINT_RESOLVER_WITH_INDEX, func2)
	m.mainFrame.SetContext(collaborate.INTERFACE_GEN_INT_RANGE_RESOLVER, func3)
	m.mainFrame.SetContext(collaborate.INTERFACE_GEN_YES_NO_RESOLVER, func4)
	m.mainFrame.SetContext(collaborate.INTERFACE_QUERY_FOR_PLAYER_NAME, func5)
}
