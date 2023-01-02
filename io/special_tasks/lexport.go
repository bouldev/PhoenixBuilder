//go:build !is_tweak
// +build !is_tweak

package special_tasks

import (
	"fmt"
	"phoenixbuilder/fastbuilder/bdump"
	"phoenixbuilder/fastbuilder/configuration"
	"phoenixbuilder/fastbuilder/environment"
	"phoenixbuilder/fastbuilder/parsing"
	"phoenixbuilder/fastbuilder/task"
	"phoenixbuilder/io/special_tasks/lexport_depends"
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
	// 取得起点坐标和终点坐标
	if beginPos.Y < -64 {
		beginPos.Y = -64
	}
	if endPos.Y > 320 {
		endPos.Y = 320
	}
	// 虽然应该不会有人尝试超高度，不过我觉得还是要处理一下这个细节（
	go func() {
		defer func() {
			err := recover()
			if err != nil {
				debug.PrintStack()
				env.CommandSender.Output(pterm.Error.Sprintf("go routine @ fastbuilder.task lexport crashed\n", err))
			}
		}()
		// 当出现惊慌的时候不应该全部崩掉，而是尝试恢复其
		u_d1, _ := uuid.NewUUID()
		env.CommandSender.SendWSCommand("gamemode c", u_d1)
		// 改创造
		allAreasSplitAns, allAreasFindUse, useForProgress := lexport_depends.SplitArea(beginPos.X, beginPos.Y, beginPos.Z, endPos.X, endPos.Y, endPos.Z, 64, 64, true)
		// 拆分目标导出区域为若干个小区域
		// 每个小区域最大 64*64
		allAreas := make([]lexport_depends.Mcstructure, 0)
		for key, value := range allAreasSplitAns {
			currentProgress := useForProgress[key]
			env.CommandSender.Output(pterm.Info.Sprintf("Now Fetching data from area(relative to the starting point) [%v, %v]", currentProgress.Posx, currentProgress.Posz))
			// 打印进度
			u_d2, _ := uuid.NewUUID()
			wchan := make(chan *packet.CommandOutput)
			(*env.CommandSender.GetUUIDMap()).Store(u_d2.String(), wchan)
			env.CommandSender.SendWSCommand(fmt.Sprintf("tp %d %d %d", value.BeginX, value.BeginY, value.BeginZ), u_d2)
			<-wchan
			close(wchan)
			// 传送玩家
			for {
				u_d3, _ := uuid.NewUUID()
				chann := make(chan *packet.CommandOutput)
				(*env.CommandSender.GetUUIDMap()).Store(u_d3.String(), chann)
				env.CommandSender.SendWSCommand(fmt.Sprintf("execute @a[name=\"%v\"] ~ ~ ~ testforblock ~ 2023 ~ air", env.Connection.(*minecraft.Conn).IdentityData().DisplayName), u_d3)
				resp := <-chann
				close(chann)
				if resp.SuccessCount > 0 {
					break
				}
			}
			// 确认目标区域是否已经加载
			ExportWaiter = make(chan map[string]interface{})
			env.Connection.(*minecraft.Conn).WritePacket(&packet.StructureTemplateDataRequest{
				StructureName: "PhoenixBuilder:LexportUsed",
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
				},
				RequestType: packet.StructureTemplateRequestExportFromSave,
			})
			exportData := <-ExportWaiter
			close(ExportWaiter)
			// 获取 mcstructure
			got, err := lexport_depends.GetMCStructureData(value, exportData)
			if err != nil {
				panic(err)
			} else {
				allAreas = append(allAreas, got)
			}
			// 添加数据
		}
		env.CommandSender.Output(pterm.Info.Sprint("Data received, processing......"))
		env.CommandSender.Output(pterm.Info.Sprint("Extracting blocks......"))
		// 打印进度
		ans, err := lexport_depends.ExportBaseOnChunkSize(allAreas, allAreasFindUse, lexport_depends.Area{
			BeginX: beginPos.X,
			BeginY: beginPos.Y,
			BeginZ: beginPos.Z,
			SizeX:  endPos.X - beginPos.X + 1,
			SizeY:  endPos.Y - beginPos.Y + 1,
			SizeZ:  endPos.Z - beginPos.Z + 1,
		})
		if err != nil {
			panic(err)
		}
		// 重排处理并写入方块数据
		outputResult := bdump.BDumpLegacy{
			Blocks: ans,
		}
		if strings.LastIndex(cfg.Path, ".bdx") != len(cfg.Path)-4 || len(cfg.Path) < 4 {
			cfg.Path += ".bdx"
		}
		// 确定输出位置并封装结果
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
		// 导出
	}()
	return nil
}
