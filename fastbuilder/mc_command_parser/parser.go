package mc_command_parser

import (
	"strings"
)

// 以当前阅读进度为起始，
// 匹配一个前缀 expect 。
//
// 如果 isCommandHeader 为真，
// 则代表该前缀是一个 MC 命令，
// 否则应该是如 detect 的前缀。
//
// 如果前缀成功匹配，则返回真，
// 否则返回假。当成功匹配时，
// 底层阅读器的阅读进度将会更新，
// 否则将会保持不变
func (p *CommandParser) ExpectHeader(expect string, isCommandHeader bool) (is bool) {
	r := p.reader
	// 初始化
	r.JumpSpace()
	if isCommandHeader {
		switch r.Next(true) {
		case "/":
			r.JumpSpace()
		case "":
		default:
			r.SetPtr(r.Pointer() - 1)
		}
	}
	// 跳过 空格 及 斜杠(可选)
	older := r.Pointer()
	l := len(expect)
	if len(r.String()) >= r.Pointer()+l && strings.ToLower(r.Sentence(l)) == expect {
		return true
	}
	// 当前缀与 expect 匹配
	r.SetPtr(older)
	return false
	// 当前缀不与 expect 匹配
}

// 以当前阅读进度为起始，
// 解析并返回一个目标选择器
func (p *CommandParser) ParseSelector() (selector Selector) {
	r := p.reader
	// prepare
	switch r.Next(false) {
	case `@`:
		older := r.Pointer() - 1
		func() {
			for {
				switch r.Next(false) {
				case " ", "[", "~", "^", "\n", "+", "\t":
					r.SetPtr(r.Pointer() - 1)
					selector.Main = r.CutSentence(older)
					return
				}
			}
		}()
		r.JumpSpace()
		// ^ @...
		// e.g. `@p`
		switch r.Next(true) {
		case "[", "":
		default:
			r.SetPtr(r.Pointer() - 1)
			return
		}
		// ^ (Pre-Check) @...[
		// e.g. `@e   [`
		older = r.Pointer() - 1
		for {
			switch r.Next(false) {
			case `"`:
				r.ParseString()
			case `]`:
				tmp := r.CutSentence(older)
				selector.Sub = &tmp
				return
			}
		}
		// ^ @...[...]
		// e.g. `@s [name=abc,tag="\"abcdefg\\/higklmn\""]`
	case `"`:
		selector.Main = r.ParseString()
		return
		// ^ "..."
		// e.g. `"Happy\\2018/new"`
	default:
		older := r.Pointer() - 1
		for {
			switch op := r.Next(true); op {
			case " ", "~", "^", "\n", "+", "\t", "":
				if op != "" {
					r.SetPtr(r.Pointer() - 1)
				}
				if selector.Main = r.CutSentence(older); len(selector.Main) == 0 {
					panic("ParseSelector: EOF")
				}
				return
			}
		}
		// ^ ...
		// e.g. `你好世界`
	}
	// process and return
}

// 以当前阅读进度为起始，
// 解析一组坐标并返回其对应的字符串切片
func (p *CommandParser) ParsePosition() (pos [3]string) {
	r := p.reader
	// prepare
	for i := 0; i < 3; i++ {
		if i > 0 {
			r.JumpSpace()
		}
		// jump space
		switch op := r.Next(false); op {
		case "~", "^":
			pos[i] = op
		default:
			r.SetPtr(r.Pointer() - 1)
		}
		// get symbol
		switch op := r.Next(true); op {
		case "+", "-", "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
			r.SetPtr(r.Pointer() - 1)
			if len(pos[i]) == 0 && i != 1 {
				tmp, _ := r.ParseNumber(false)
				pos[i] = pos[i] + tmp
			} else {
				tmp, _ := r.ParseNumber(true)
				pos[i] = pos[i] + tmp
			}
			if pos[i] == "~0" || pos[i] == "^0" {
				pos[i] = pos[i][0:1]
			}
		default:
			if i != 2 {
				switch op {
				case "~", "^":
					r.SetPtr(r.Pointer() - 1)
				case " ", "\n", "\t":
				case "":
					if len(pos[i]) == 0 {
						panic("ParsePosition: EOF")
					}
				default:
					panic("ParsePosition: Invalid position")
				}
			} else if len(pos[i]) == 0 {
				panic("ParsePosition: Invalid position")
			} else if op != "" {
				r.SetPtr(r.Pointer() - 1)
			}
		}
		// get position
	}
	// scan the string
	return
	// return
}

// 以当前阅读进度为起始，
// 解析被测方块的各项预期参数
func (p *CommandParser) ParseDetectArgs() (detectArgs DetectArgs) {
	var isInt bool
	r := p.reader
	// prepare
	detectArgs.BlockPosition = p.ParsePosition()
	// block position
	r.JumpSpace()
	older := r.Pointer()
	func() {
		for {
			switch r.Next(true) {
			case " ", "\n", "\t":
				r.SetPtr(r.Pointer() - 1)
				return
			case "":
				return
			case "+":
				panic("ParseDetectArgs: Invalid block data value")
			}
		}
	}()
	detectArgs.BlockName = r.CutSentence(older)
	// block name
	r.JumpSpace()
	if detectArgs.BlockData, isInt = r.ParseNumber(true); !isInt {
		panic("ParseDetectArgs: Block data provided must be an integer")
	}
	// block data
	return
	// return
}
