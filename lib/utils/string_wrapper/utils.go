package string_wrapper

import (
	"fmt"
	"strconv"
	"strings"
)

func TranslateVersionStringToNumber(version string) (uint64, error) {
	version = strings.TrimSpace(version)
	frags := strings.Split(version, ".")
	maxFragLen := (64 - len(frags) + 1) / len(frags)
	versionNumber := uint64(0)
	for i := 0; i < len(frags); i++ {
		if i != 0 {
			versionNumber <<= maxFragLen
		}
		if fversion, err := strconv.Atoi(frags[i]); err != nil {
			return 0, err
		} else {
			versionNumber += uint64(fversion)
		}
	}
	return versionNumber, nil
}

func TranslateStringToIntPos(str string) (pos [3]int, err error) {
	s, e := 0, len(str)
	if str == "" {
		return [3]int{}, fmt.Errorf("empty str")
	}
	for s < e {
		if (str[s] >= '0' && str[s] <= '9') || (str[s] == '-') {
			break
		}
		s++
	}
	for e > s {
		if (str[e-1] >= '0' && str[e-1] <= '9') || (str[e-1] == '-') {
			break
		}
		e--
	}
	str = str[s:e]
	if str == "" {
		return [3]int{}, fmt.Errorf("empty str")
	}
	str = strings.ReplaceAll(str, ",", " ")
	str = strings.ReplaceAll(str, "，", " ")
	ss := strings.Split(str, " ")
	i := 0
	for _, s := range ss {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		atoi, err := strconv.Atoi(s)
		if err != nil {
			return [3]int{}, err
		}
		pos[i] = atoi
		i++
	}
	if i != 3 {
		return pos, fmt.Errorf("insufficient pos length")
	}
	return pos, nil
}
