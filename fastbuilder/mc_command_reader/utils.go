package mc_command_reader

import (
	"encoding/json"
)

// 以当前阅读进度为起始，
// 跳过空格、换行符和制表符，
// 直到抵达非空字符或 EOF 时止
func (c *CommandReader) JumpSpace() {
	for {
		switch c.Next() {
		case " ", "\n", "\t":
		default:
			c.SetPtr(c.Pointer() - 1)
			return
		}
	}
}

// 以当前阅读进度为起始，
// 解析并返回一个字符串。
// e.g. `233\\\""` -> `233\"`
func (c *CommandReader) ParseString() (res string) {
	older := c.Pointer() - 1
	for {
		k := c.Next()
		//fmt.Println(k)
		switch k {
		case `\`:
			c.SetPtr(c.Pointer() + 1)
		case `"`:
			tmp := c.SentenceThroughPtr(older, nil)
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

	`02.300 ` -> `2.3`
	`+0 ` -> `0`
	`-0 ` -> `0`
	`+2+3 ` -> `2`

^ in general

	`02.000 ` -> `2`

^ omission = true

	`02.000 ` -> `2.0`

^ omission = false

The following example will panic.

	`2.0` -> EOF
	`2. ` -> EOF
	`+ ` -> EOF
	`.2 ` -> Invalid number
	`+- ` -> Invalid number
	`2..0 ` -> Invalid number
*/
func (c *CommandParser) ParseNumber(omission bool) (res string, isInt bool) {
	isNegative := false
	isFirstOp := true
	isZero := true
	hasPoint := false
	// init values
	switch op := c.Next(); op {
	case "+":
	case "-":
		isNegative = true
	default:
		c.SetPtr(c.Pointer() - 1)
	}
	// get symbol
	older := c.Pointer()
	func() {
		for {
			switch op := c.Next(); op {
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
				c.SetPtr(c.Pointer() - 1)
				res = c.SentenceThroughPtr(older, nil)
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
