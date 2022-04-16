package utils

import (
	"fmt"
	"strings"
)

func GetStringContents(s string) []string {
	_s := strings.Split(RemoveFormate(s), " ")
	for i, c := range _s {
		_s[i] = strings.TrimSpace(c)
	}
	ss := make([]string, 0, len(_s))
	for _, c := range _s {
		if c != "" {
			ss = append(ss, c)
		}
	}
	return ss
}

func RemoveFormate(in string) string {
	ss := make([]byte, 0, len(in))
	flag := 0
	for i := 0; i < len(in); i++ {
		if flag != 0 {
			flag--
			continue
		}
		if in[i] == 194 {
			flag = 2
			continue
		}
		ss = append(ss, in[i])
	}
	return string(ss)
}

func CanTrigger(ss []string, triggers []string, allowNoSpace bool, removeColor bool) (bool, []string) {
	if len(ss) == 0 {
		//for _, t := range triggers {
		//	if t == "" {
		//		return true, ss
		//	}
		//}
		return false, ss
	}
	s := ss[0]
	if removeColor {
		for {
			if strings.HasPrefix(s, "§") {
				s = s[1:]
				if len(s) > 0 {
					s = s[1:]
				}
			} else {
				break
			}
		}
	}
	flag := false
	for _, tw := range triggers {
		if strings.HasPrefix(s, tw) {
			if s == tw {
				s = ""
				flag = true
				break
			} else if allowNoSpace {
				s = s[len(tw):]
				flag = true
				break
			}
		}
	}
	if !flag {
		return false, ss
	}
	as := make([]string, len(ss))
	for i, _s := range ss {
		as[i] = _s
	}
	if s == "" {
		return true, as[1:]
	} else {
		as[0] = s
		return true, as
	}
}

func FormateByRepalcment(tmp string, replacements map[string]interface{}) string {
	s := tmp
	for k, v := range replacements {
		s = strings.ReplaceAll(s, k, fmt.Sprintf("%v", v))
	}
	return s
}

func ToPlainName(name string) string {
	if strings.Contains(name, ">") {
		name = strings.ReplaceAll(name, ">", " ")
		name = strings.ReplaceAll(name, "<", " ")
	}
	if name != "" {
		names := GetStringContents(name)
		name = names[len(names)-1]
	}
	return name
}
