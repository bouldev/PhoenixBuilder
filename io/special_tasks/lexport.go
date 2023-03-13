//go:build !is_tweak
// +build !is_tweak

package special_tasks

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/fastbuilder/bdump"
	"phoenixbuilder/fastbuilder/configuration"
	"phoenixbuilder/fastbuilder/environment"
	"phoenixbuilder/fastbuilder/mcstructure"
	"phoenixbuilder/fastbuilder/parsing"
	"phoenixbuilder/fastbuilder/task"
	"phoenixbuilder/minecraft"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"runtime/debug"
	"strings"

	"github.com/google/uuid"
	"github.com/pterm/pterm"
)

func CreateLegacyExportTask(commandLine string, env *environment.PBEnvironment) *task.Task {
	cfg, err := parsing.Parse(commandLine, configuration.GlobalFullConfig(env).Main())
	if err != nil {
		env.CommandSender.Output(pterm.Error.Sprintf("Failed to parse command: %v", err))
		return nil
	}
	// 解析控制台输入
	beginPos := cfg.Position
	endPos := cfg.End
	if beginPos.X > endPos.X {
		save := beginPos.X
		beginPos.X = endPos.X
		endPos.X = save
	}
	if beginPos.Y > endPos.Y {
		save := beginPos.Y
		beginPos.Y = endPos.Y
		endPos.Y = save
	}
	if beginPos.Z > endPos.Z {
		save := beginPos.Z
		beginPos.Z = endPos.Z
		endPos.Z = save
	}

	if beginPos.Y < -64 {
		beginPos.Y = -64
	}
	if endPos.Y > 320 {
		endPos.Y = 320
	}
	go func() {
		defer func() {
			err := recover()
			if err != nil {
				debug.PrintStack()
				env.CommandSender.Output(pterm.Error.Sprintf("go routine @ fastbuilder.task lexport crashed"))
				env.CommandSender.Output(pterm.Error.Sprintf("%v", err))
			}
		}()

		u_d0, _ := uuid.NewUUID()
		env.CommandSender.SendWSCommand("gamemode c", u_d0)

		u_d1, _ := uuid.NewUUID()
		chann := make(chan *packet.CommandOutput)
		(*env.CommandSender.GetUUIDMap()).Store(u_d1.String(), chann)
		env.CommandSender.SendWSCommand("querytarget @s", u_d1)
		resp := <-chann
		close(chann)
		var dimension float64 = 0
		var got interface{}
		var testAreaIsLoaded string = "testforblocks ~-31 -64 ~-31 ~31 319 ~31 ~-31 -64 ~-31"
		json.Unmarshal([]byte(resp.OutputMessages[0].Parameters[0]), &got)
		dimension = got.([]interface{})[0].(map[string]interface{})["dimension"].(float64)
		if dimension == 1 {
			testAreaIsLoaded = "testforblocks ~-31 0 ~-31 ~31 127 ~31 ~-31 0 ~-31"
		}
		if dimension == 2 {
			testAreaIsLoaded = "testforblocks ~-31 0 ~-31 ~31 255 ~31 ~-31 0 ~-31"
		}
		// 这个前置准备用于后面判断被导出区域是否加载
		// 如果尝试请求一个没有被完全加载的区域，那么返回的结构将是只包括空气的结构，但不会报错
		// 如果被请求的区域部分没有加载，那么可能地，没有加载的部分就是空气了
		// 所以为了永远地规避区域加载问题，我这里使用了 testforblocks 方法用于到时候给每个待导出区域检查是否已完全加载
		// 这里应该说一下，如果尝试将 testforblocks 命令用于一个没有被加载的区域，返回的
		// OutputMessages[0].Message 字段是 "commands.generic.outOfWorld"
		// 你可能会说，为什么不用 testforblock 命令给单个方块作检测，这是因为
		// 目标待导出区域最大是 64*64 ，而只对单方块检测并不能保证整个待导出区域都已经加载了
		splittedAreas, reversedMap, indicativeMap := mcstructure.SplitArea(mcstructure.BlockPos{int32(beginPos.X), int32(beginPos.Y), int32(beginPos.Z)}, mcstructure.BlockPos{int32(endPos.X), int32(endPos.Y), int32(endPos.Z)}, 64, 64, true)
		// 拆分目标导出区域为若干个小区域
		// 每个小区域最大 64*64
		allAreas := make([]mcstructure.Mcstructure, 0)
		for key, value := range splittedAreas {
			currentProgress := indicativeMap[key]
			env.CommandSender.Output(pterm.Info.Sprintf("Fetching data from area [%d, %d]", currentProgress[0], currentProgress[1]))
			u_d2, _ := uuid.NewUUID()
			wchan := make(chan *packet.CommandOutput)
			(*env.CommandSender.GetUUIDMap()).Store(u_d2.String(), wchan)
			env.CommandSender.SendWSCommand(fmt.Sprintf("tp %d %d %d", value.BeginX+value.SizeX/2, value.BeginY+value.SizeY/2, value.BeginZ+value.SizeZ/2), u_d2)
			<-wchan
			close(wchan)

			for {
				u_d3, _ := uuid.NewUUID()
				chann := make(chan *packet.CommandOutput)
				(*env.CommandSender.GetUUIDMap()).Store(u_d3.String(), chann)
				env.CommandSender.SendWSCommand(testAreaIsLoaded, u_d3)
				resp := <-chann
				close(chann)
				//fmt.Printf("%#v\n",resp)
				if resp.OutputMessages[0].Message != "commands.generic.outOfWorld" {
					break
				}
			}
			// 等待当前被访问的区块加载完成
			ExportWaiter = make(chan map[string]interface{})
			env.Connection.(*minecraft.Conn).WritePacket(&packet.StructureTemplateDataRequest{
				StructureName: "mystructure:aaaaa",
				Position:      protocol.BlockPos{int32(value.BeginX), int32(value.BeginY), int32(value.BeginZ)},
				Settings: protocol.StructureSettings{
					PaletteName:               "default",
					IgnoreEntities:            true,
					IgnoreBlocks:              false,
					Size:                      protocol.BlockPos{int32(value.SizeX), int32(value.SizeY), int32(value.SizeZ)},
					Offset:                    protocol.BlockPos{0, 0, 0},
					LastEditingPlayerUniqueID: env.Connection.(*minecraft.Conn).GameData().EntityUniqueID,
					Rotation:                  0,
					Mirror:                    0,
					Integrity:                 100,
					Seed:                      0,
					AllowNonTickingChunks:     false,
				},
				RequestType: packet.StructureTemplateRequestExportFromSave,
			})
			exportData := <-ExportWaiter
			close(ExportWaiter)
			// 获取 mcstructure
			got, err := mcstructure.GetMCStructureData(value, exportData)
			if err != nil {
				panic(err)
			} else {
				allAreas = append(allAreas, got)
			}
		}
		env.CommandSender.Output(pterm.Info.Sprint("Data received, processing......"))
		env.CommandSender.Output(pterm.Info.Sprint("Extracting blocks......"))

		processedData, err := mcstructure.DumpBlocks(allAreas, reversedMap, mcstructure.Area{
			BeginX: int32(beginPos.X),
			BeginY: int32(beginPos.Y),
			BeginZ: int32(beginPos.Z),
			SizeX:  int32(endPos.X - beginPos.X + 1),
			SizeY:  int32(endPos.Y - beginPos.Y + 1),
			SizeZ:  int32(endPos.Z - beginPos.Z + 1),
		})
		if err != nil {
			panic(err)
		}

		outputResult := bdump.BDump{
			Blocks: processedData,
		}
		if strings.LastIndex(cfg.Path, ".bdx") != len(cfg.Path)-4 || len(cfg.Path) < 4 {
			cfg.Path += ".bdx"
		}

		env.CommandSender.Output(pterm.Info.Sprint("Writing output file......"))
		err, signerr := outputResult.WriteToFile(cfg.Path, env.LocalCert, env.LocalKey)
		if err != nil {
			env.CommandSender.Output(pterm.Error.Sprintf("Failed to export: %v", err))
			return
		} else if signerr != nil {
			env.CommandSender.Output(pterm.Info.Sprintf("Note: The file is unsigned since the following error was trapped: %v", signerr))
		} else {
			env.CommandSender.Output(pterm.Success.Sprint("File signed successfully"))
		}
		env.CommandSender.Output(pterm.Success.Sprintf("Successfully exported your structure to %v", cfg.Path))
	}()
	return nil
}
