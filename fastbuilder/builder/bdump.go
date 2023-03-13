package builder

import (
	"fmt"
	"io"
	"os"
	"phoenixbuilder/fastbuilder/bdump"
	"phoenixbuilder/fastbuilder/bdump/command"

	//bridge_path "phoenixbuilder/fastbuilder/builder/path"
	I18n "phoenixbuilder/fastbuilder/i18n"
	"phoenixbuilder/fastbuilder/types"
	"phoenixbuilder/fastbuilder/world_provider"

	"github.com/andybalholm/brotli"
)

func readZeroTerminatedString(br io.Reader) (string, error) {
	str := ""
	c := make([]byte, 1)
	for {
		_, err := br.Read(c)
		if err != nil {
			return "", err
		}
		if c[0] == 0 {
			break
		}
		str += string(c)
	}
	return str, nil
}

func BDump(config *types.MainConfig, blc chan *types.Module) error {
	file, err := os.Open(config.Path)
	if err != nil {
		return I18n.ProcessSystemFileError(err)
	}
	defer file.Close()
	{
		header3bytes := make([]byte, 3)
		_, err := io.ReadAtLeast(file, header3bytes, 3)
		if err != nil {
			return fmt.Errorf(I18n.T(I18n.BDump_EarlyEOFRightWhenOpening))
		}
		if string(header3bytes) != "BD@" {
			return fmt.Errorf(I18n.T(I18n.BDump_NotBDX_Invheader))
		}
	}
	br := brotli.NewReader(file)
	signed, corrupted, signer_username, err := bdump.VerifyStreamBDX(br)
	if !signed {
		if config.Strict {
			return fmt.Errorf("%s.", I18n.T(I18n.BDump_FileNotSigned))
		} else {
			types.ForwardedBrokSender <- fmt.Sprintf("%s!", I18n.T(I18n.BDump_FileNotSigned))
		}
	}
	if corrupted {
		return fmt.Errorf(I18n.T(I18n.FileCorruptedError))
	}
	// The original one is consumed
	file.Seek(3, os.SEEK_SET)
	br = brotli.NewReader(file)
	if err != nil {
		e := fmt.Errorf(I18n.T(I18n.BDump_VerificationFailedFor), err)
		if config.Strict {
			return e
		} else {
			types.ForwardedBrokSender <- fmt.Sprintf("%s(%s): %v", I18n.T(I18n.ERRORStr), I18n.T(I18n.IgnoredStr), e)
		}
	} else {
		types.ForwardedBrokSender <- fmt.Sprintf(I18n.T(I18n.BDump_FileSigned), signer_username)
	}
	{
		tempbuf := make([]byte, 4)
		_, err := io.ReadAtLeast(br, tempbuf, 4)
		if err != nil {
			return fmt.Errorf(I18n.T(I18n.InvalidFileError))
		}
		if string(tempbuf) != "BDX\x00" {
			return fmt.Errorf(I18n.T(I18n.BDump_NotBDX_Invinnerheader))
		}
	}
	readZeroTerminatedString(br) // Ignores author field
	brushPosition := []int{0, 0, 0}
	var blocksStrPool []string
	var runtimeIdPoolUsing []*types.ConstBlock
	for {
		_cmd, err := command.ReadCommand(br)
		if err != nil {
			return fmt.Errorf("%s: %v", I18n.T(I18n.BDump_FailedToGetConstructCmd), err)
		}
		_, isTerminate := _cmd.(*command.Terminate)
		if isTerminate {
			break
		}
		switch cmd := _cmd.(type) {
		case *command.CreateConstantString:
			blocksStrPool = append(blocksStrPool, cmd.ConstantString)
		case *command.AddInt16ZValue0:
			brushPosition[2] += int(cmd.Value)
		case *command.PlaceBlock:
			if int(cmd.BlockConstantStringID) >= len(blocksStrPool) {
				return fmt.Errorf("Error: BlockID exceeded BlockPool")
			}
			blockName := &blocksStrPool[int(cmd.BlockConstantStringID)]
			blc <- &types.Module{
				Block: &types.Block{
					Name: blockName,
					Data: cmd.BlockData,
				},
				Point: types.Position{
					X: brushPosition[0] + config.Position.X,
					Y: brushPosition[1] + config.Position.Y,
					Z: brushPosition[2] + config.Position.Z,
				},
			}
		case *command.AddZValue0:
			brushPosition[2]++
		case *command.NoOperation:
			// Command: NOP, DO NOTHING
			break
		case *command.AddInt32ZValue0:
			brushPosition[2] += int(cmd.Value)
		case *command.AddXValue:
			brushPosition[0]++
		case *command.SubtractXValue:
			brushPosition[0]--
		case *command.AddYValue:
			brushPosition[1]++
		case *command.SubtractYValue:
			brushPosition[1]--
		case *command.AddZValue:
			brushPosition[2]++
		case *command.SubtractZValue:
			brushPosition[2]--
		case *command.AddInt16XValue:
			brushPosition[0] += int(cmd.Value)
		case *command.AddInt32XValue:
			brushPosition[0] += int(cmd.Value)
		case *command.AddInt16YValue:
			brushPosition[1] += int(cmd.Value)
		case *command.AddInt32YValue:
			brushPosition[1] += int(cmd.Value)
		case *command.AddInt16ZValue:
			brushPosition[2] += int(cmd.Value)
		case *command.AddInt32ZValue:
			brushPosition[2] += int(cmd.Value)
		case *command.SetCommandBlockData:
			blc <- &types.Module{
				CommandBlockData: cmd.CommandBlockData,
				Point: types.Position{
					X: brushPosition[0] + config.Position.X,
					Y: brushPosition[1] + config.Position.Y,
					Z: brushPosition[2] + config.Position.Z,
				},
			}
		case *command.PlaceBlockWithCommandBlockData:
			if int(cmd.BlockConstantStringID) >= len(blocksStrPool) {
				return fmt.Errorf("Error: BlockConstantStringID exceeded BlockPool length")
			}
			blockName := &blocksStrPool[int(cmd.BlockConstantStringID)]
			cmdl := &types.Module{
				Block: &types.Block{
					Name: blockName,
					Data: cmd.BlockData,
				},
				Point: types.Position{
					X: brushPosition[0] + config.Position.X,
					Y: brushPosition[1] + config.Position.Y,
					Z: brushPosition[2] + config.Position.Z,
				},
				CommandBlockData: cmd.CommandBlockData,
			}
			blc <- cmdl
		case *command.PlaceCommandBlockWithCommandBlockData:
			blockName := "command_block"
			cmdl := &types.Module{
				Block: &types.Block{
					Name: &blockName,
					Data: cmd.BlockData,
				},
				Point: types.Position{
					X: brushPosition[0] + config.Position.X,
					Y: brushPosition[1] + config.Position.Y,
					Z: brushPosition[2] + config.Position.Z,
				},
				CommandBlockData: cmd.CommandBlockData,
			}
			blc <- cmdl
		case *command.AddInt8XValue:
			brushPosition[0] += int(cmd.Value)
		case *command.AddInt8YValue:
			brushPosition[1] += int(cmd.Value)
		case *command.AddInt8ZValue:
			brushPosition[2] += int(cmd.Value)
		case *command.UseRuntimeIDPool:
			if cmd.PoolID == 117 {
				runtimeIdPoolUsing = world_provider.RuntimeIdArray_117
			} else if cmd.PoolID == 118 {
				runtimeIdPoolUsing = world_provider.RuntimeIdArray_2_1_10
			} else {
				return fmt.Errorf("This file is using an unknown runtime id pool, we're unable to resolve it.")
			}
		case *command.PlaceRuntimeBlock:
			if int(cmd.BlockRuntimeID) >= len(runtimeIdPoolUsing) {
				return fmt.Errorf("Fatal: Block with runtime ID %d not found", cmd.BlockRuntimeID)
			}
			blc <- &types.Module{
				Block: runtimeIdPoolUsing[int(cmd.BlockRuntimeID)].Take(),
				Point: types.Position{
					X: brushPosition[0] + config.Position.X,
					Y: brushPosition[1] + config.Position.Y,
					Z: brushPosition[2] + config.Position.Z,
				},
			}
		case *command.PlaceRuntimeBlockWithUint32RuntimeID:
			if int(cmd.BlockRuntimeID) >= len(runtimeIdPoolUsing) {
				return fmt.Errorf("Fatal: Block with runtime ID %d not found", cmd.BlockRuntimeID)
			}
			blc <- &types.Module{
				Block: runtimeIdPoolUsing[cmd.BlockRuntimeID].Take(),
				Point: types.Position{
					X: brushPosition[0] + config.Position.X,
					Y: brushPosition[1] + config.Position.Y,
					Z: brushPosition[2] + config.Position.Z,
				},
			}
		case *command.PlaceRuntimeBlockWithCommandBlockData:
			if int(cmd.BlockRuntimeID) >= len(runtimeIdPoolUsing) {
				return fmt.Errorf("Fatal: Block with runtime ID %d not found", cmd.BlockRuntimeID)
			}
			cmdl := &types.Module{
				Block: runtimeIdPoolUsing[int(cmd.BlockRuntimeID)].Take(),
				Point: types.Position{
					X: brushPosition[0] + config.Position.X,
					Y: brushPosition[1] + config.Position.Y,
					Z: brushPosition[2] + config.Position.Z,
				},
				CommandBlockData: cmd.CommandBlockData,
			}
			blc <- cmdl
		case *command.PlaceRuntimeBlockWithCommandBlockDataAndUint32RuntimeID:
			if int(cmd.BlockRuntimeID) >= len(runtimeIdPoolUsing) {
				return fmt.Errorf("Fatal: Block with runtime ID %d not found", cmd.BlockRuntimeID)
			}
			cmdl := &types.Module{
				Block: runtimeIdPoolUsing[cmd.BlockRuntimeID].Take(),
				Point: types.Position{
					X: brushPosition[0] + config.Position.X,
					Y: brushPosition[1] + config.Position.Y,
					Z: brushPosition[2] + config.Position.Z,
				},
				CommandBlockData: cmd.CommandBlockData,
			}
			blc <- cmdl
		case *command.PlaceRuntimeBlockWithChestData:
			if int(cmd.BlockRuntimeID) >= len(runtimeIdPoolUsing) {
				return fmt.Errorf("Fatal: Block with runtime ID %d not found", cmd.BlockRuntimeID)
			}
			pos := types.Position{
				X: brushPosition[0] + config.Position.X,
				Y: brushPosition[1] + config.Position.Y,
				Z: brushPosition[2] + config.Position.Z,
			}
			blc <- &types.Module{
				Block: runtimeIdPoolUsing[int(cmd.BlockRuntimeID)].Take(),
				Point: pos,
			}
			for _, slot := range cmd.ChestSlots {
				slotcopy := types.ChestSlot(slot)
				blc <- &types.Module{
					ChestSlot: &slotcopy,
					Point:     pos,
				}
			}
		case *command.PlaceBlockWithChestData:
			if int(cmd.BlockConstantStringID) >= len(blocksStrPool) {
				return fmt.Errorf("Error: BlockConstantStringID exceeded BlockPool length")
			}
			blockName := &blocksStrPool[int(cmd.BlockConstantStringID)]
			pos := types.Position{
				X: brushPosition[0] + config.Position.X,
				Y: brushPosition[1] + config.Position.Y,
				Z: brushPosition[2] + config.Position.Z,
			}
			blc <- &types.Module{
				Block: &types.Block{
					Name: blockName,
					Data: cmd.BlockData,
				},
				Point: pos,
			}
			for _, slot := range cmd.ChestSlots {
				slotcopy := types.ChestSlot(slot)
				blc <- &types.Module{
					ChestSlot: &slotcopy,
					Point:     pos,
				}
			}
		case *command.PlaceRuntimeBlockWithChestDataAndUint32RuntimeID:
			if int(cmd.BlockRuntimeID) >= len(runtimeIdPoolUsing) {
				return fmt.Errorf("Fatal: Block with runtime ID %d not found", cmd.BlockRuntimeID)
			}
			pos := types.Position{
				X: brushPosition[0] + config.Position.X,
				Y: brushPosition[1] + config.Position.Y,
				Z: brushPosition[2] + config.Position.Z,
			}
			blc <- &types.Module{
				Block: runtimeIdPoolUsing[cmd.BlockRuntimeID].Take(),
				Point: pos,
			}
			for _, slot := range cmd.ChestSlots {
				slotcopy := types.ChestSlot(slot)
				blc <- &types.Module{
					ChestSlot: &slotcopy,
					Point:     pos,
				}
			}
		case *command.AssignDebugData:
			// Not going to do anything with those data
		case *command.PlaceBlockWithBlockStatesDeprecated:
			if int(cmd.BlockConstantStringID) >= len(blocksStrPool) {
				return fmt.Errorf("Error: BlockID exceeded BlockPool")
			}
			blockName := &blocksStrPool[int(cmd.BlockConstantStringID)]
			blc <- &types.Module{
				Block: &types.Block{
					Name:        blockName,
					BlockStates: cmd.BlockStatesString,
				},
				Point: types.Position{
					X: brushPosition[0] + config.Position.X,
					Y: brushPosition[1] + config.Position.Y,
					Z: brushPosition[2] + config.Position.Z,
				},
			}
		case *command.PlaceBlockWithBlockStates:
			if int(cmd.BlockConstantStringID) >= len(blocksStrPool) {
				return fmt.Errorf("Error: BlockID exceeded StringPool")
			}
			if int(cmd.BlockStatesConstantStringID) >= len(blocksStrPool) {
				return fmt.Errorf("Error: BlockStatesID exceeded StringPool")
			}
			blockName := &blocksStrPool[int(cmd.BlockConstantStringID)]
			blockStates := blocksStrPool[int(cmd.BlockStatesConstantStringID)]
			blc <- &types.Module{
				Block: &types.Block{
					Name:        blockName,
					BlockStates: blockStates,
				},
				Point: types.Position{
					X: brushPosition[0] + config.Position.X,
					Y: brushPosition[1] + config.Position.Y,
					Z: brushPosition[2] + config.Position.Z,
				},
			}
		case *command.PlaceBlockWithNBTData:
			if int(cmd.BlockConstantStringID) >= len(blocksStrPool) {
				return fmt.Errorf("Error: BlockID exceeded StringPool")
			}
			if int(cmd.BlockStatesConstantStringID) >= len(blocksStrPool) {
				return fmt.Errorf("Error: BlockStatesID exceeded StringPool")
			}
			blockName := &blocksStrPool[int(cmd.BlockConstantStringID)]
			blockStates := blocksStrPool[int(cmd.BlockStatesConstantStringID)]
			blc <- &types.Module{
				Block: &types.Block{
					Name:        blockName,
					BlockStates: blockStates,
				},
				NBTMap: cmd.BlockNBT,
				Point: types.Position{
					X: brushPosition[0] + config.Position.X,
					Y: brushPosition[1] + config.Position.Y,
					Z: brushPosition[2] + config.Position.Z,
				},
			}
		default:
			fmt.Printf("WARNING: BDump/Import: Unknown method found: %#v\n\n", _cmd)
			fmt.Printf("WARNING: BDump/Import: THIS IS A BUG\n")
		}
	}
	return nil
}
