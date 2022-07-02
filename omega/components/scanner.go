package components

import (
	"encoding/json"
	"fmt"
	"os"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/mirror"
	"phoenixbuilder/mirror/define"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"strings"
	"time"

	"github.com/pterm/pterm"
)

type Scanner struct {
	*defines.BasicComponent
	isScanning   bool
	resultWriter *simpleFileLineDstWrapper
	Trigger      string   `json:"触发词"`
	FilterHas    []string `json:"如果包含以下关键词则忽略"`
	FilterHasnt  []string `json:"如果不包含以下关键词则忽略"`
}

func (o *Scanner) Init(cfg *defines.ComponentConfig) {
	m, _ := json.Marshal(cfg.Configs)
	err := json.Unmarshal(m, o)
	if err != nil {
		panic(err)
	}
}

func (o *Scanner) Filter(l string) string {
	if o.FilterHas != nil && len(o.FilterHas) != 0 {
		for _, h := range o.FilterHas {
			if strings.Contains(l, h) {
				return ""
			}
		}
	}
	if o.FilterHasnt != nil && len(o.FilterHasnt) != 0 {
		for _, h := range o.FilterHasnt {
			if strings.Contains(l, h) {
				return l
			}
		}
		return ""
	}
	return l
}

func (o *Scanner) onLevelChunk(cd *mirror.ChunkData) {
	if o.isScanning {
		for _, nbt := range cd.BlockNbts {
			if x, y, z, success := define.GetPosFromNBT(nbt); success {
				marshal, _ := json.Marshal(nbt)
				if o.resultWriter != nil {
					l := fmt.Sprintf("block @ %v %v %v: %v\n", x, y, z, string(marshal))
					if l = o.Filter(l); l == "" {
						continue
					}
					fmt.Print(l)
					o.resultWriter.Write(l)
				}
			}
		}
	}
}

func (o *Scanner) onBlockActorData(pk *packet.BlockActorData) {
	if o.isScanning {
		nbt := pk.NBTData
		if x, y, z, success := define.GetPosFromNBT(nbt); success {
			marshal, _ := json.Marshal(nbt)
			if o.resultWriter != nil {
				l := fmt.Sprintf("block @ %v %v %v: %v\n", x, y, z, string(marshal))
				if l = o.Filter(l); l == "" {
					return
				}
				fmt.Print(l)
				o.resultWriter.Write(l)
			}
		}
	}
}

func (o *Scanner) onActor(pk *packet.AddActor) {
	if o.isScanning {
		if o.resultWriter != nil {
			marshal, _ := json.Marshal(pk.EntityMetadata)
			l := fmt.Sprintf("entity %v @ %v %v %v: %v\n",
				strings.TrimLeft(pk.EntityType, "minecraft:"),
				pk.Position.X(), pk.Position.Y(), pk.Position.Z(),
				string(marshal),
			)
			if l = o.Filter(l); l == "" {
				return
			}
			fmt.Print(l)
			o.resultWriter.Write(l)
		}
	}
}

type simpleFileLineDstWrapper struct {
	fp   *os.File
	name string
}

func (s *simpleFileLineDstWrapper) Write(data string) {
	s.fp.Write([]byte(data))
}

func (o *Scanner) handleScan(cmds []string) bool {
	// fmt.Println(cmds)
	if len(cmds) != 0 && cmds[0] == "done" {
		pterm.Info.Println("停止扫描,结果已经导出到 " + o.resultWriter.name)
		pterm.Warning.Printfln("注意：机器人可能忽略扫描前已经看到的生物和nbt方块")
		o.resultWriter.fp.Close()
		o.isScanning = false
		o.resultWriter = nil
		return true
	}
	var x, z int
	if utils.SimplePrase(&cmds, []string{"x"}, &x, true) &&
		utils.SimplePrase(&cmds, []string{"z"}, &z, true) {
		if o.isScanning {
			pterm.Error.Println("上一个扫描尚未关闭，请输入 " + o.Trigger + " done 关闭上一个扫描")
			return true
		}
		o.Frame.GetBotTaskScheduler().CommitNormalTask(&defines.BasicBotTaskPauseAble{
			BasicBotTask: defines.BasicBotTask{
				Name: fmt.Sprintf("Scan"),
				ActivateFn: func() {
					pterm.Info.Println("正在准备扫描")
					o.Frame.GetGameControl().SendCmdAndInvokeOnResponseWithFeedback(
						fmt.Sprintf("tp @s %v %v %v", x+1000, 255, z+1000), func(output *packet.CommandOutput) {
							o.Frame.GetGameControl().SendCmdAndInvokeOnResponseWithFeedback(
								fmt.Sprintf("tp @s %v %v %v", x, 255, z), func(output *packet.CommandOutput) {
									data_mark := time.Now().Format("2006-01-02-15:04:05")
									fileName := "扫描日志" + data_mark + ".txt"
									fileName = o.Frame.GetRelativeFileName(fileName)
									fp, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0777)
									o.resultWriter = &simpleFileLineDstWrapper{fp: fp, name: fileName}
									if err != nil {
										o.Frame.GetBackendDisplay().Write(pterm.Error.Sprintfln("记录文件打开失败，扫描终止,%v", err))
									}
									o.isScanning = true
									pterm.Info.Printfln("扫描开始，结果将保存到 %v 中，输入 "+o.Trigger+" done 停止扫描", fileName)
									pterm.Warning.Printfln("注意：机器人可能忽略扫描前已经看到的生物和nbt方块")

								},
							)
						},
					)
					time.Sleep(time.Second * 10)
					if o.isScanning {
						pterm.Info.Println("停止扫描,结果已经导出到 " + o.resultWriter.name)
						pterm.Warning.Printfln("注意：机器人可能忽略扫描前已经看到的生物和nbt方块")
						o.resultWriter.fp.Close()
						o.isScanning = false
						o.resultWriter = nil
					}
				},
			},
		})
		return true
	}
	pterm.Error.Printfln("错误的指令格式，应该为\n" + o.Trigger + " x z\n或者\n" + o.Trigger + " done")

	return true
}

func (o *Scanner) Inject(frame defines.MainFrame) {
	o.Frame = frame
	o.Frame.GetGameListener().SetOnLevelChunkCallBack(o.onLevelChunk)
	o.Frame.GetGameListener().SetOnTypedPacketCallBack(packet.IDAddActor, func(p packet.Packet) {
		o.onActor(p.(*packet.AddActor))
	})
	o.Frame.GetGameListener().SetOnTypedPacketCallBack(packet.IDBlockActorData, func(p packet.Packet) {
		o.onBlockActorData(p.(*packet.BlockActorData))
	})
	o.Frame.SetBackendMenuEntry(&defines.BackendMenuEntry{
		MenuEntry: defines.MenuEntry{
			Triggers:     []string{o.Trigger},
			ArgumentHint: "坐标，以 x z 形式/或 done",
			FinalTrigger: false,
			Usage:        "扫描指定区域的所有nbt方块和实体",
		},
		OptionalOnTriggerFn: o.handleScan,
	})
}
