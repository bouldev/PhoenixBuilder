package mc_command_reader

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

// 描述一个 Execute 命令
type ExecuteCommand struct {
	Selector   Selector    // 指定的命令执行者
	Position   [3]string   // 指定的命令执行位置
	DetectArgs *DetectArgs // 被测方块的各项预期参数
	SubCommand string      // 子命令
}

// 从 command 解析一个 execute 命令。
// 若返回 nil ，则当前不是一个 execute 命令
func ParseExecuteCommand(command string) (e *ExecuteCommand) {
	command = command + "X"
	// avoid EOF error
	p := NewCommandReader(&command).Parser()
	if p.ExpectHeader("execute", true) {
		e = &ExecuteCommand{}
	} else {
		return
	}
	// check header
	p.JumpSpace()
	e.Selector = p.ParseSelector()
	// get selector
	p.JumpSpace()
	e.Position = p.ParsePosition()
	// get block position
	p.JumpSpace()
	if p.ExpectHeader("detect", false) {
		p.JumpSpace()
		tmp := p.ParseDetectArgs()
		e.DetectArgs = &tmp
	}
	// get detect args
	p.JumpSpace()
	e.SubCommand = command[p.ptr : len(command)-1]
	// get sub command
	return
	// return
}
