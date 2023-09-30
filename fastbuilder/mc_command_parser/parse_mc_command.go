package mc_command_parser

// 从 command 解析一个 execute 命令。
// 若返回 nil ，则当前不是一个 execute 命令
func ParseExecuteCommand(command string) (e *ExecuteCommand) {
	p := NewCommandParser(&command)
	r := p.reader
	// prepare
	if p.ExpectHeader("execute", true) {
		e = &ExecuteCommand{}
	} else {
		return
	}
	// check header
	r.JumpSpace()
	e.Selector = p.ParseSelector()
	// parse selector
	r.JumpSpace()
	e.Position = p.ParsePosition()
	// parse block position
	r.JumpSpace()
	if p.ExpectHeader("detect", false) {
		r.JumpSpace()
		tmp := p.ParseDetectArgs()
		e.DetectArgs = &tmp
	}
	// parse detect args
	r.JumpSpace()
	e.SubCommand = command[r.Pointer():]
	// get sub command
	return
	// return
}
