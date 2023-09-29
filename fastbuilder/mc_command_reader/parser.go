package mc_command_reader

import "strings"

// 描述一个单个的命令解析器，
// 其底层由命令阅读器构成
type CommandParser struct {
	*CommandReader
}

// 返回当前阅读器对应的命令解析器
func (c *CommandReader) Parser() *CommandParser {
	return &CommandParser{c}
}

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
	p.JumpSpace()
	if isCommandHeader && p.Next() != "/" {
		p.SetPtr(p.Pointer() - 1)
	}
	// 跳过 空格 及 斜杠(可选)
	older := p.Pointer()
	l := len(expect)
	if len(p.String()) >= p.Pointer()+l && strings.ToLower(p.Sentence(l)) == expect {
		return true
	}
	// 当前缀与 expect 匹配
	p.SetPtr(older)
	return false
	// 当前缀不与 expect 匹配
}

// 以当前阅读进度为起始，
// 解析并返回一个目标选择器
func (p *CommandParser) ParseSelector() (selector Selector) {
	switch p.Next() {
	case `@`:
		older := p.Pointer() - 1
		func() {
			for {
				switch p.Next() {
				case " ", "[", "~", "^", "\n", "+", "\t":
					p.SetPtr(p.Pointer() - 1)
					selector.Main = p.SentenceThroughPtr(older, nil)
					return
				}
			}
		}()
		p.JumpSpace()
		// ^ @s
		// e.g. `@p`
		if p.Next() != `[` {
			p.SetPtr(p.Pointer() - 1)
			return
		}
		// ^ (Pre-Check) @s...[
		// e.g. `@e   [`
		older = p.Pointer() - 1
		for {
			switch p.Next() {
			case `"`:
				p.ParseString()
			case `]`:
				tmp := p.SentenceThroughPtr(older, nil)
				selector.Sub = &tmp
				return
			}
		}
		// ^ @s...[...]
		// e.g. `@s [name=abc,tag="\"abcdefg\\/higklmn\""]`
	case `"`:
		selector.Main = p.ParseString()
		return
		// ^ "..."
		// e.g. `"Happy\\2018/new"`
	default:
		older := p.Pointer() - 1
		for {
			switch p.Next() {
			case " ", "\n", "+", "\t":
				p.SetPtr(p.Pointer() - 1)
				selector.Main = p.SentenceThroughPtr(older, nil)
				return
			}
		}
		// ^ ...
		// e.g. `你好世界`
	}
}

// 以当前阅读进度为起始，
// 解析一组坐标并返回其对应的字符串切片
func (p *CommandParser) ParsePosition() (pos [3]string) {
	for i := 0; i < 3; i++ {
		if i > 0 {
			p.JumpSpace()
		}
		// jump space
		switch op := p.Next(); op {
		case "~", "^":
			pos[i] = op
		default:
			p.SetPtr(p.Pointer() - 1)
		}
		// get symbol
		switch op := p.Next(); op {
		case "+", "-", "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
			p.SetPtr(p.Pointer() - 1)
			if len(pos[i]) == 0 && i != 1 {
				tmp, _ := p.ParseNumber(false)
				pos[i] = pos[i] + tmp
			} else {
				tmp, _ := p.ParseNumber(true)
				pos[i] = pos[i] + tmp
			}
			if pos[i] == "~0" || pos[i] == "^0" {
				pos[i] = pos[i][0:1]
			}
		default:
			if i != 2 {
				switch op {
				case "~", "^":
					p.SetPtr(p.Pointer() - 1)
				case " ", "\n", "\t":
				default:
					panic("ParsePosition: Invalid position")
				}
			} else if len(pos[i]) == 0 {
				panic("ParsePosition: Invalid position")
			} else {
				p.SetPtr(p.Pointer() - 1)
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
	detectArgs.BlockPosition = p.ParsePosition()

	p.JumpSpace()
	older := p.Pointer()
	func() {
		for {
			switch p.Next() {
			case " ", "\n", "\t":
				p.SetPtr(p.Pointer() - 1)
				return
			case "+":
				panic("ParseDetectArgs: Invalid block data value")
			}
		}
	}()
	detectArgs.BlockName = p.SentenceThroughPtr(older, nil)

	p.JumpSpace()
	if detectArgs.BlockData, isInt = p.ParseNumber(true); !isInt {
		panic("CommandParser: Block data provided must be an integer")
	}

	return
}
