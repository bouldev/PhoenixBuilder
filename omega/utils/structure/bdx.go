package structure

import (
	"bytes"
	"fmt"
	"io"
	I18n "phoenixbuilder/fastbuilder/i18n"
	"phoenixbuilder/fastbuilder/types"
	"phoenixbuilder/fastbuilder/world_provider"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/mirror/chunk"
	"phoenixbuilder/mirror/define"
	"phoenixbuilder/fastbuilder/bdump/command"

	"github.com/andybalholm/brotli"
)

func ReadBrString(src io.Reader) (string, error) {
	str := ""
	c := make([]byte, 1)
	for {
		_, err := src.Read(c)
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

func handleBDXCMD(br io.Reader, infoSender func(string)) (author string, blockChan chan *IOBlockForDecoder, cancelFn func(), err error) {
	legacyRunTimeIDRemapper := NewDoubleValueLegacyBlockToRuntimeIDMapper()
	blockChan = make(chan *IOBlockForDecoder, 10240)
	{
		tempbuf := make([]byte, 4)
		_, err = io.ReadAtLeast(br, tempbuf, 4)
		if err != nil {
			err = fmt.Errorf(I18n.T(I18n.InvalidFileError))
			return
		}
		if string(tempbuf) != "BDX\x00" {
			err = fmt.Errorf(I18n.T(I18n.BDump_NotBDX_Invinnerheader))
			return
		}
	}
	author = ""
	// This field is outdated and should be ignored
	_, _err := ReadBrString(br)
	if _err != nil {
		err = fmt.Errorf("cannot read author info, %v", err)
		return
	}
	brushPosition := define.CubePos{0, 0, 0}
	paletteIDCounter := uint16(0)
	runtimeIdPoolUsing := NewRuntimeIDConvertor()
	stop := false
	go func() {
		defer close(blockChan)
		for {
			if stop {
				break
			}
			_cmd, err := command.ReadCommand(br)
			if(err!=nil) {
				err=fmt.Errorf("%s: %v", I18n.T(I18n.BDump_FailedToGetConstructCmd), err)
				return
			}
			_, isTerminate:=_cmd.(*command.Terminate)
			if isTerminate {
				break
			}
			switch cmd:=_cmd.(type) {
			case *command.CreateConstantString:
				legacyRunTimeIDRemapper.AddBlockNamePalette(paletteIDCounter, cmd.ConstantString)
				paletteIDCounter++
				continue
			case *command.AddInt16ZValue0:
				brushPosition[2] += int(cmd.Value)
			case *command.PlaceBlock:
				blockId := cmd.BlockConstantStringID
				blockData := cmd.BlockData
				blockChan <- &IOBlockForDecoder{
					Pos:       brushPosition,
					BlockName: legacyRunTimeIDRemapper.palatteIDToBlockNameMapping[blockId],
					BlockData: blockData,
				}
			case *command.AddZValue0:
				brushPosition[2]++
			case *command.NoOperation:
				// Command: NOP
			case *command.AddInt32ZValue0:
				brushPosition[2] += int(cmd.Value)
			case *command.PlaceBlockWithBlockStates:
				blockId := cmd.BlockConstantStringID
				block_states_string := cmd.BlockStatesString
				if err != nil {
					infoSender("Failed to get argument for cmd[pos5], file may be corrupted")
					return
				}
				blockChan <- &IOBlockForDecoder{
					Pos:         brushPosition,
					BlockStates: block_states_string,
					BlockName:   legacyRunTimeIDRemapper.palatteIDToBlockNameMapping[blockId],
				}
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
				// Omega originally deal w/ it in a wrong way
				// I have no idea on how to correct it.
				commandBlockData := cmd.CommandBlockData
				cbmode := commandBlockData.Mode
				command := commandBlockData.Command
				cusname := commandBlockData.CustomName
				blockName := "command_block"
				switch cbmode {
				case packet.CommandBlockImpulse:
				case packet.CommandBlockRepeating:
					blockName = "repeating_command_block"
				case packet.CommandBlockChain:
					blockName = "chain_command_block"
				}
				commandBlockNbt := map[string]interface{}{
					"id":                 "CommandBlock",
					"Command":            command,
					"CustomName":         cusname,
					"ExecuteOnFirstTick": commandBlockData.ExecuteOnFirstTick,
					"TickDelay":          commandBlockData.TickDelay,
					"auto":               !commandBlockData.NeedsRedstone,
					"TrackOutput":        commandBlockData.TrackOutput,
					"conditionalMode":    commandBlockData.Conditional,
				}
				blockChan <- &IOBlockForDecoder{
					Pos:  brushPosition,
					BlockName: blockName,
					NBT:  commandBlockNbt,
				}
			case *command.PlaceCommandBlockWithCommandBlockData:
				commandBlockData := cmd.CommandBlockData
				cbmode := commandBlockData.Mode
				command := commandBlockData.Command
				cusname := commandBlockData.CustomName
				blockName := "command_block"
				switch cbmode {
				case packet.CommandBlockImpulse:
				case packet.CommandBlockRepeating:
					blockName = "repeating_command_block"
				case packet.CommandBlockChain:
					blockName = "chain_command_block"
				}
				commandBlockNbt := map[string]interface{}{
					"id":                 "CommandBlock",
					"Command":            command,
					"CustomName":         cusname,
					"ExecuteOnFirstTick": commandBlockData.ExecuteOnFirstTick,
					"TickDelay":          commandBlockData.TickDelay,
					"auto":               !commandBlockData.NeedsRedstone,
					"TrackOutput":        commandBlockData.TrackOutput,
					"conditionalMode":    commandBlockData.Conditional,
				}
				blockChan <- &IOBlockForDecoder{
					Pos:  brushPosition,
					BlockName: blockName,
					NBT:  commandBlockNbt,
				}
			case *command.AddInt8XValue:
				brushPosition[0] += int(cmd.Value)
			case *command.AddInt8YValue:
				brushPosition[1] += int(cmd.Value)
			case *command.AddInt8ZValue:
				brushPosition[2] += int(cmd.Value)
			case *command.UseRuntimeIDPool:
				var pool []*types.ConstBlock
				if cmd.PoolID == 117 {
					pool = world_provider.RuntimeIdArray_117
				} else if cmd.PoolID == 118 {
					pool = world_provider.RuntimeIdArray_2_1_10
				} else {
					infoSender("This file is using an unknown runtime id pool, we're unable to resolve it.")
					return
				}
				runtimeIdPoolUsing = NewRuntimeIDConvertor()
				runtimeIdPoolUsing.ConvertFN = func(u uint32) uint32 {
					blk := pool[u].Take()
					if rtid, found := chunk.LegacyBlockToRuntimeID(*blk.Name, blk.Data); found {
						return rtid
					} else {
						return chunk.AirRID
					}
				}
			case *command.PlaceRuntimeBlock:
				blockChan <- &IOBlockForDecoder{
					Pos:  brushPosition,
					RTID: runtimeIdPoolUsing.Convert(uint32(cmd.BlockRuntimeID)),
				}
			case *command.PlaceRuntimeBlockWithUint32RuntimeID:
				blockChan <- &IOBlockForDecoder{
					Pos:  brushPosition,
					RTID: runtimeIdPoolUsing.Convert(cmd.BlockRuntimeID),
				}
			case *command.PlaceRuntimeBlockWithCommandBlockData:
				commandBlockData:=cmd.CommandBlockData
				commandBlockNbt := map[string]interface{}{
					"id":                 "CommandBlock",
					"Command":            commandBlockData.Command,
					"CustomName":         commandBlockData.CustomName,
					"ExecuteOnFirstTick": commandBlockData.ExecuteOnFirstTick,
					"TickDelay":          commandBlockData.TickDelay,
					"auto":               !commandBlockData.NeedsRedstone,
					"TrackOutput":        commandBlockData.TrackOutput,
					"LastOutput":         "",
					"conditionalMode":    commandBlockData.Conditional,
				}
				blockChan <- &IOBlockForDecoder{
					Pos:  brushPosition,
					RTID: runtimeIdPoolUsing.Convert(uint32(cmd.BlockRuntimeID)),
					NBT:  commandBlockNbt,
				}
			case *command.PlaceRuntimeBlockWithCommandBlockDataAndUint32RuntimeID:
				commandBlockData:=cmd.CommandBlockData
				commandBlockNbt := map[string]interface{}{
					"id":                 "CommandBlock",
					"Command":            commandBlockData.Command,
					"CustomName":         commandBlockData.CustomName,
					"ExecuteOnFirstTick": commandBlockData.ExecuteOnFirstTick,
					"TickDelay":          commandBlockData.TickDelay,
					"auto":               !commandBlockData.NeedsRedstone,
					"TrackOutput":        commandBlockData.TrackOutput,
					"LastOutput":         "",
					"conditionalMode":    commandBlockData.Conditional,
				}
				blockChan <- &IOBlockForDecoder{
					Pos:  brushPosition,
					RTID: runtimeIdPoolUsing.Convert(cmd.BlockRuntimeID),
					NBT:  commandBlockNbt,
				}
			case *command.PlaceRuntimeBlockWithChestData:
				// This part does not work at first
			case *command.PlaceBlockWithChestData:
				// Does not work at first
			case *command.PlaceRuntimeBlockWithChestDataAndUint32RuntimeID:
			case *command.AssignNBTData:
			default:
				// fmt.Println("ERROR!")
				infoSender(fmt.Sprintf("BUG: unimplemented method found : %#v", _cmd))
				return
			}
		}
	}()
	return author, blockChan, func() { stop = true }, nil
}

func checkAndGetBDXHeader(src io.Reader) (err error) {
	header3bytes := make([]byte, 3)
	_, err = src.Read(header3bytes)
	if err != nil {
		return fmt.Errorf("file too short, %v", err)

	}
	if string(header3bytes) != "BD@" {
		return fmt.Errorf(I18n.T(I18n.BDump_NotBDX_Invheader))
	}
	return nil
}

func DecodeBDX(data []byte, infoSender func(string)) (blockFeeder chan *IOBlockForDecoder, cancelFn func(), suggestMinCacheChunks int, totalBlocks int, err error) {
	defer func() {
		r := recover()
		if r != nil {
			err = fmt.Errorf("unknown error %v", r)
		}
	}()
	file := bytes.NewBuffer(data)
	err = checkAndGetBDXHeader(file)
	if err != nil {
		err = ErrImportFormatNotSupport
		return
	}

	bro := brotli.NewReader(file)
	br := &bytes.Buffer{}
	filelen, _ := br.ReadFrom(bro)
	if filelen == 0 {
		err = fmt.Errorf("cannnot use brotli for decompress")
		return
	}
	infoSender("正在检查 BDX 文件，需要消耗大量时间")
	author, blockChan, _, err := handleBDXCMD(br, func(s string) {
		err = fmt.Errorf(s)
	})
	if err != nil {
		return
	}
	blockCounter := 0
	b0 := <-blockChan
	minX := b0.Pos.X()
	maxX := minX
	for b := range blockChan {
		x := b.Pos.X()
		if x < minX {
			minX = x
		} else if x > maxX {
			maxX = x
		}
		blockCounter++
	}
	infoSender(fmt.Sprintf("文件检查完毕, 方块数 %v", blockCounter))
	if err == nil {
		infoSender("author: " + author)
		file = bytes.NewBuffer(data)
		checkAndGetBDXHeader(file)
		bro := brotli.NewReader(file)
		br := &bytes.Buffer{}
		_, _ = br.ReadFrom(bro)
		_, blockChan, stopFn, _ := handleBDXCMD(br, func(s string) {})
		return blockChan, stopFn, (maxX-minX)/16 + 2, blockCounter, nil
	}
	return
}
