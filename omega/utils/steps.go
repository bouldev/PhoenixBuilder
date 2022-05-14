package utils

import (
	"fmt"
	"phoenixbuilder/omega/collaborate"
	"phoenixbuilder/omega/defines"
)

func QueryForPlayerName(ctrl defines.GameControl, src string, dst string, searchFn collaborate.FUNC_GetPossibleName) (name string, cancel bool) {
	termCtx := make(chan bool)
	var hint []string
	var resolver func(params []string) (int, bool, error)
	candidateNames := []string{}
	for {
		if dst == "" {
			hint = []string{"请输入目标玩家名,或者目标玩家名的一部分(或输入: 取消 )"}
			resolver = nil
		} else {
			candidateNames = []string{}
			possibleNames := searchFn(dst, 3)
			if len(possibleNames) > 0 && possibleNames[0].Entry.CurrentName == dst {
				return dst, false
			}
			hint = []string{"请选择下方的序号，或者目标玩家名(或输入: 取消 )"}
			for i, name := range possibleNames {
				candidateNames = append(candidateNames, name.Entry.CurrentName)
				currentName, historyName := name.GenReadAbleStringPair()
				hint = append(hint, fmt.Sprintf("\n%v: %v %v", i+1, currentName, historyName))
			}
			if len(candidateNames) > 0 {
				_, resolver = GenStringListHintResolverWithIndex(candidateNames)
			} else {
				hint = []string{"没有匹配的玩家，请输入目标玩家名,或者目标玩家名的一部分(或输入: 取消 )"}
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
				}
				selection, cancel, err := resolver(chat.Msg)
				if err != nil {
					termCtx <- false
					dst = ""
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
