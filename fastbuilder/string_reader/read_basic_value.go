package string_reader

import (
	"encoding/json"
	"strings"
)

// 以当前阅读进度为起始，
// 跳过空格、换行符和制表符，
// 直到抵达非空字符或 EOF 时止
func (s *StringReader) JumpSpace() {
	for {
		switch s.Next(true) {
		case " ", "\n", "\t":
		case "":
			return
		default:
			s.SetPtr(s.Pointer() - 1)
			return
		}
	}
}

// 以当前阅读进度为起始，
// 解析并返回一个布尔值
func (s *StringReader) ParseBool() (res bool) {
	if part1 := s.Sentence(4); len(part1) < 4 {
		panic("ParseBool: EOF")
	} else if strings.ToLower(part1) == "true" {
		res = true
	} else {
		part2 := s.Next(false)
		res = strings.ToLower(part1+part2) != "false"
		if res {
			panic("ParseBool: Invalid boolean")
		}
	}
	return
}

// 以当前阅读进度为起始，
// 解析并返回一个字符串。
// e.g. `233\\\""` -> `233\"`
func (s *StringReader) ParseString() (res string) {
	older := s.Pointer() - 1
	for {
		switch s.Next(false) {
		case `\`:
			s.SetPtr(s.Pointer() + 1)
		case `"`:
			tmp := s.CutSentence(older)
			json.Unmarshal([]byte(tmp), &res)
			return res
		}
	}
}

/*
以当前阅读进度为起始，
解析一个数字并返回其经过标准化后的字符串形式。
omission 为真时将对数字进行完全简化。

e.g.

	`02.300` -> `2.3`
	`+0` -> `0`
	`-0` -> `0`
	`+2+3` -> `2`

^ in general

	`02.000` -> `2`

^ omission = true

	`02.000` -> `2.0`

^ omission = false

The following example will panic.

	`2.` -> EOF
	`+` -> EOF
	`.2` -> Invalid number
	`+-` -> Invalid number
	`2..0` -> Invalid number
*/
func (s *StringReader) ParseNumber(omission bool) (res string, isInt bool) {
	isNegative := false
	isFirstOp := true
	isZero := true
	hasPoint := false
	// init values
	switch op := s.Next(false); op {
	case "+":
	case "-":
		isNegative = true
	default:
		s.SetPtr(s.Pointer() - 1)
	}
	// get symbol
	older := s.Pointer()
	func() {
		for {
			switch op := s.Next(true); op {
			case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
				if isFirstOp {
					isFirstOp = false
				}
				if op == "0" && isZero {
					older++
				}
				if op != "0" && isZero {
					isZero = false
				}
			case ".":
				if isFirstOp || hasPoint {
					panic("ParseNumber: Invalid number")
				}
				if isZero {
					older--
					isZero = false
				}
				hasPoint = true
			case "-":
				panic("ParseNumber: Invalid number")
			default:
				if op != "" {
					s.SetPtr(s.Pointer() - 1)
				}
				res = s.CutSentence(older)
				if len(res) == 0 && !isFirstOp {
					res = "0"
				}
				return
			}
		}
	}()
	// scan the string and format the integral part
	ptr := len(res)
	if ptr < 1 {
		panic("ParseNumber: EOF")
	}
	if res[ptr-1:ptr] == "." {
		panic("ParseNumber: EOF")
	}
	func() {
		if hasPoint {
			for {
				switch res[ptr-1 : ptr] {
				case "0":
					ptr--
				default:
					res = res[:ptr]
					return
				}
			}
		}
	}()
	if res[len(res)-1:] == "." {
		if omission {
			res = res[:len(res)-1]
		} else {
			res = res + "0"
		}
	}
	// format the fractional part
	if res != "0" && res != "0.0" && isNegative {
		res = "-" + res
	}
	// adjust positive and negative
	return res, !hasPoint
	// return
}
