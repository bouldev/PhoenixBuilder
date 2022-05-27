package components

import (
	"encoding/json"
	"fmt"
	"github.com/pterm/pterm"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/collaborate"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"strings"
	"time"
)

type packageRecord struct {
	Src           string `json:"寄出人"`
	Name          string `json:"物品名"`
	TimeStamp     string `json:"寄出时间"`
	StructureName string `json:"结构方块名"`
}

type ExpressInfo struct {
	Packages     map[string][]*packageRecord `json:"包裹"`
	CurrentIndex int                         `json:"当前序号"`
}

type Express struct {
	*BasicComponent
	Triggers          []string `json:"触发词"`
	LoginDelay        int      `json:"登录时延迟发送"`
	Usage             string   `json:"提示信息"`
	Response          string   `json:"寄出成功时提示"`
	HintOnPlayerReady string   `json:"提示玩家把要投递的东西丢到身边"`
	FileName          string   `json:"记录文件"`
	PackagePlatform   []int    `json:"打包平台"`
	SelectCmd         string   `json:"物品转移器"`
	Record            ExpressInfo
	PlayerSearcher    collaborate.FUNC_GetPossibleName
}

func (o *Express) formatPackage(idx int) string {
	return fmt.Sprintf("OMEp%v", idx)
}

func (o *Express) delivery(playerName string) {
	if pkgs, hasK := o.Record.Packages[playerName]; hasK {
		if len(pkgs) > 0 {
			if player := o.Frame.GetGameControl().GetPlayerKit(playerName); player != nil {
				player.Title("有新快递")
				player.SubTitle("物品将被放到周围")
				for _, p := range pkgs {
					player.Say(p.Name)
					player.Say(fmt.Sprintf("是 %v 寄给你的", p.Src))
					cmd := fmt.Sprintf("execute %v ~~~ structure load %v ~~~ 0_degrees none true false", playerName, p.StructureName)
					o.Frame.GetBackendDisplay().Write("将 " + p.Name + " 派送到 " + playerName + " " + p.StructureName)
					o.Frame.GetGameControl().SendCmd(cmd)
				}
				delete(o.Record.Packages, playerName)
			}
		} else {
			delete(o.Record.Packages, playerName)
		}
	}
}

func (o *Express) post(srcPlayer, dstPlayer, hint string) {
	fmt.Println(srcPlayer, dstPlayer, hint)
	cmd := utils.FormatByReplacingOccurrences(o.SelectCmd, map[string]interface{}{"[player]": srcPlayer})
	o.Frame.GetGameControl().SendCmdAndInvokeOnResponse(cmd, func(output *packet.CommandOutput) {
		if output.SuccessCount == 0 {
			o.Frame.GetGameControl().SayTo(srcPlayer, "物品转移失败")
		} else {
			ox, oy, oz := o.PackagePlatform[0], o.PackagePlatform[1], o.PackagePlatform[2]
			sx, sy, sz := ox-1, oy, oz-1
			ex, ey, ez := ox+1, oy+1, oz+1
			packageName := o.formatPackage(o.Record.CurrentIndex)
			cmd = fmt.Sprintf("structure save %v %v %v %v %v %v %v true disk false",
				packageName, sx, sy, sz, ex, ey, ez)
			o.Record.CurrentIndex++
			o.Frame.GetGameControl().SendCmdAndInvokeOnResponse(cmd, func(output *packet.CommandOutput) {
				if output.SuccessCount != 0 {
					o.Frame.GetGameControl().SayTo(srcPlayer, "打包成功！将在目标玩家上线时送到")
					o.Frame.GetGameControl().SendCmd(fmt.Sprintf("tp @e[r=3,x=%v,y=%v,z=%v] 0 -40 0", ox, oy, oz))
					o.Frame.GetBackendDisplay().Write(fmt.Sprintf("%v->%v: 寄出 %v (%v)", srcPlayer, dstPlayer, hint, packageName))
					if _, hask := o.Record.Packages[dstPlayer]; !hask {
						o.Record.Packages[dstPlayer] = make([]*packageRecord, 0)
					}
					o.Record.Packages[dstPlayer] = append(o.Record.Packages[dstPlayer], &packageRecord{
						Src:           srcPlayer,
						Name:          hint,
						TimeStamp:     utils.TimeToString(time.Now()),
						StructureName: packageName,
					})
					go func() {
						time.Sleep(3 * time.Second)
						o.delivery(dstPlayer)
					}()
				} else {
					o.Frame.GetGameControl().SendCmd(fmt.Sprintf("tp @e[r=3,x=%v,y=%v,z=%v] %v", ox, oy, oz, srcPlayer))
					o.Frame.GetGameControl().SayTo(srcPlayer, "打包失败，尝试退回物品")
					o.Frame.GetBackendDisplay().Write(fmt.Sprintf("%v->%v: 寄出失败 %v (%v)", srcPlayer, dstPlayer, hint, packageName))
				}

			})
		}
	})
}

func (o *Express) askForPackage(srcPlayer, dstPlayer string) {
	//dstPlayer := chat.Msg[0]
	if player := o.Frame.GetGameControl().GetPlayerKit(srcPlayer); player != nil {
		if player.SetOnParamMsg(func(c *defines.GameChat) bool {
			//c.Msg = utils.InsertHead[string](dstPlayer, c.Msg)
			o.post(srcPlayer, dstPlayer, strings.Join(c.Msg, " "))
			return true
		}) == nil {
			o.Frame.GetGameControl().SayTo(srcPlayer, o.HintOnPlayerReady)
		}
	}
}

func (o *Express) queryPlayer(chat *defines.GameChat) bool {
	go func() {
		if name, cancel := utils.QueryForPlayerName(
			o.Frame.GetGameControl(), chat.Name,
			"",
			(*o.Frame.GetContext())[collaborate.INTERFACE_POSSIBLE_NAME].(collaborate.FUNC_GetPossibleName)); !cancel {
			o.askForPackage(chat.Name, name)
		} else {
			o.Frame.GetGameControl().SayTo(chat.Name, "已取消")
		}
	}()
	return true
}

func (o *Express) Init(cfg *defines.ComponentConfig) {
	m, _ := json.Marshal(cfg.Configs)
	if err := json.Unmarshal(m, o); err != nil {
		panic(err)
	}
	if len(o.PackagePlatform) != 3 {
		panic(fmt.Errorf("打包平台 %v 坐标无效，应该为  [x,y,z] ", o.PackagePlatform))
	}
}

func (o *Express) Inject(frame defines.MainFrame) {
	o.Frame = frame
	o.Frame.GetGameListener().AppendLoginInfoCallback(func(entry protocol.PlayerListEntry) {
		name := utils.ToPlainName(entry.Username)
		if _, hasK := o.Record.Packages[name]; hasK {
			timer := time.NewTimer(time.Duration(o.LoginDelay) * time.Second)
			go func() {
				<-timer.C
				o.delivery(name)
			}()
		}
	})
	if o.Usage == "" {
		o.Usage = "给某个玩家寄送物资，将在上线时投递"
	}
	o.Frame.GetGameListener().SetGameMenuEntry(&defines.GameMenuEntry{
		MenuEntry: defines.MenuEntry{
			Triggers:     o.Triggers,
			ArgumentHint: "[玩家]",
			Usage:        o.Usage,
			FinalTrigger: false,
		},
		OptionalOnTriggerFn: o.queryPlayer,
	})
	o.Record = ExpressInfo{
		Packages:     map[string][]*packageRecord{},
		CurrentIndex: 1,
	}
	err := frame.GetJsonData(o.FileName, &o.Record)
	if err != nil {
		panic(err)
	}
	o.PlayerSearcher = (*frame.GetContext())[collaborate.INTERFACE_POSSIBLE_NAME].(collaborate.FUNC_GetPossibleName)
}

func (o *Express) Activate() {
	ox, oy, oz := o.PackagePlatform[0], o.PackagePlatform[1], o.PackagePlatform[2]
	for h := -1; h < 0; h++ {
		for wx := -1; wx < 2; wx++ {
			for wz := -1; wz < 2; wz++ {
				x, y, z := ox+wx, oy+h, oz+wz
				//fmt.Println(x, y, z)
				o.Frame.GetGameControl().
					SendCmdAndInvokeOnResponse(
						fmt.Sprintf("testforblock %v %v %v air", x, y, z),
						func(output *packet.CommandOutput) {
							//fmt.Println(output)
							if len(output.OutputMessages) > 0 && strings.Contains(output.OutputMessages[0].Message, "outOfWorld") {
								panic(pterm.Error.Sprintf("打包平台 %v 不在常加载区内！请修改打包平台位置或者设为常加载区", o.PackagePlatform))
							}
							if output.SuccessCount != 0 {
								o.Frame.GetGameControl().SendCmd(fmt.Sprintf("setblock %v %v %v sealantern", x, y, z))
							}
						},
					)
			}
		}
	}
}

func (o *Express) Stop() error {
	fmt.Printf("正在保存 %v\n", o.FileName)
	return o.Frame.WriteJsonData(o.FileName, o.Record)
}
