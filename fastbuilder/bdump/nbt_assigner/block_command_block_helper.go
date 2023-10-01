package NBTAssigner

import (
	"fmt"
	"phoenixbuilder/fastbuilder/mc_command_parser"
	"phoenixbuilder/fastbuilder/mcstructure"
	"strconv"
	"strings"
)

// 适用于 detect 修饰子命令中对方块数据值到方块状态的升级。
// 当返回的字符串为空指针时，
// 意味着未能找到对应的映射，
// 此时升级失败，否则认为升级成功。
// 特别地，如果传入的方块数据值为 -1 ，
// 则永远返回非空指针的空字符串
func upgradeBlock(name string, data int64) (states *string, err error) {
	tmp := ""
	states = &tmp
	// init values
	if data == -1 {
		return
	}
	// for special situation
	blockStates, err := get_block_states_from_legacy_block(
		strings.Replace(strings.ToLower(name), "minecraft:", "", 1),
		uint16(data),
	)
	if err != nil {
		return nil, nil
	}
	// get block_states(map)
	*states, err = mcstructure.MarshalBlockStates(blockStates)
	if err != nil {
		return nil, fmt.Errorf("upgradeBlock: Failed to marshal blockStates; blockStates = %#v, err = %v", blockStates, err)
	}
	// marshal block_states into string
	return
	// return
}

// 从 command 解析一个 execute 命令。
// 若返回 nil ，则当前不是一个 execute 命令
func parseExecuteCommand(command string) (e *mc_command_parser.ExecuteCommand, err error) {
	func() {
		defer func() {
			if errMessage := recover(); errMessage != nil {
				err = fmt.Errorf("parseExecuteCommand: %v", errMessage)
			}
		}()
		e = mc_command_parser.ParseExecuteCommand(command)
	}()
	return
}

// 将旧版本的 execute 命令升级为新格式。
// warningLogs 的状态用于指代是否需要提起警告，
// 若其中包含元素，这意味着在处理到部分 detect 字段时，
// 我们未能为其中的方块找到其对应的方块状态的映射。
// warningLogs 含有的元素即代表这些未能找到映射的方块
func UpgradeExecuteCommand(command string) (new string, warningLogs []string, err error) {
	var args *mc_command_parser.ExecuteCommand
	res := []string{}
	nextBlock := command
	wholeSelector := ""
	// init values
	for {
		current := ""
		// init value
		args, err = parseExecuteCommand(nextBlock)
		if err != nil {
			return command, nil, fmt.Errorf("UpgradeExecuteCommand: %v", err)
		}
		if args == nil {
			break
		}
		// parse execute command
		if args.Selector.Main[0:1] == "@" {
			wholeSelector = args.Selector.Main
			if args.Selector.Sub != nil {
				wholeSelector = wholeSelector + *args.Selector.Sub
			}
		} else {
			wholeSelector = fmt.Sprintf("%#v", args.Selector.Main)
		}
		current = current + fmt.Sprintf("as %s at @s ", wholeSelector)
		// upgrade selector
		switch args.Position {
		case [3]string{"~", "~", "~"}, [3]string{"^", "^", "^"}:
		default:
			current = current + fmt.Sprintf("positioned %s %s %s ", args.Position[0], args.Position[1], args.Position[2])
		}
		// upgrade poition
		if args.DetectArgs != nil {
			blockData, err := strconv.ParseInt(args.DetectArgs.BlockData, 10, 64)
			if err != nil {
				return command, nil, fmt.Errorf("UpgradeExecuteCommand: Failed to convert string into int; args.DetectArgs.BlockData = %#v, err = %v", args.DetectArgs.BlockData, err)
			}
			// convert block_data(string) into int
			blockStates, err := upgradeBlock(args.DetectArgs.BlockName, blockData)
			if blockStates == nil && err == nil {
				tmp := "[]"
				blockStates = &tmp
				warningLogs = append(warningLogs, fmt.Sprintf("%s(%d)", args.DetectArgs.BlockName, blockData))
			}
			if err != nil {
				return command, nil, fmt.Errorf("UpgradeExecuteCommand: %v", err)
			}
			// get block states from legacy block
			current = current + fmt.Sprintf(
				"if block %s %s %s %s ",
				args.DetectArgs.BlockPosition[0],
				args.DetectArgs.BlockPosition[1],
				args.DetectArgs.BlockPosition[2],
				args.DetectArgs.BlockName,
			)
			if len(*blockStates) > 0 {
				current = current + fmt.Sprintf("%s ", *blockStates)
			}
			// set new detect args
		}
		// upgrade detect args
		res = append(res, current)
		nextBlock = args.SubCommand
		// submit subresult
	}
	// scan provided command and upgrade each subcommand
	result := strings.Join(res, "")
	if len(result) == 0 {
		result = command
	} else {
		result = fmt.Sprintf("execute %srun %s", result, nextBlock)
	}
	return result, warningLogs, nil
	// splice and return
}
