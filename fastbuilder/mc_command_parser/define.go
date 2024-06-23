package mc_command_parser

import "phoenixbuilder/fastbuilder/string_reader"

// ------------------------- CommandParser ------------------------

// 描述一个单个的命令解析器，
// 其底层由命令阅读器构成
type CommandParser struct {
	reader *string_reader.StringReader
}

// 返回以 command 为底层的命令解析器
func NewCommandParser(command *string) *CommandParser {
	return &CommandParser{
		string_reader.NewStringReader(command),
	}
}

// ------------------------- Parameter ------------------------

// "color":"orange"
// or
// "color"="orange" [current]
const BlockStatesDefaultSeparator string = "="

// 描述一个目标选择器及其参数
type Selector struct {
	Main string
	Sub  *string
}

// 描述被测方块的各项预期参数
type DetectArgs struct {
	BlockPosition [3]string // 被测方块位置
	BlockName     string    // 被测方块名
	BlockData     string    // 被测方块的数据值
}

// ------------------------- Command ------------------------

// 描述一个 Execute 命令
type ExecuteCommand struct {
	Selector   Selector    // 指定的命令执行者
	Position   [3]string   // 指定的命令执行位置
	DetectArgs *DetectArgs // 被测方块的各项预期参数
	SubCommand string      // 子命令
}
