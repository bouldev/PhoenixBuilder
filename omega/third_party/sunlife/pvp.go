package sunlife

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	"sort"
	"strconv"
	"time"

	"github.com/pterm/pterm"
)

type Pvp struct {
	*defines.BasicComponent
	Usage     string            `json:"描述"`
	Triggers  []string          `json:"触发词"`
	Menu      map[string]string `json:"菜单"`
	Score     map[string]string `json:"所需计分板"`
	TitleWord map[string]string `json:"提示词"`
	//单挑
	Solo Solo `json:"单挑"`
	//多人对战
	Multiplayer map[string]*Multiplayer `json:"多人对战"`
	//混战
	ScrimmagePvp map[string]*ScrimmagePvp `json:"混战"`
	//快捷命令
	FastCmds map[string]string `json:"快捷命令"`
	//map[房间名字]房间数据
	Data map[string]*DataOfPvp
	//
	BackPos  []int  `json:"非法进入归位坐标"`
	PassWord string `json:"激活密码"`
	//玩家生存时间分数
	PlayerTime map[string]int
	//据点 已开放的据点与其坐标
	StrongHoldData map[string][]int
}
type DataOfPvp struct {
	//房间类型一共有 Solo Scr Mul分别对应单挑 混战 多人对战
	RoomType string
	//房间主人名字
	Master string
	//成员名字
	Member map[string]string
	//是否开启
	IsOpen bool
	//总体时间
	Time int
	//总奖池
	Pool int
	//进入价格
	EnterPrice int
	//地图名字
	ThePvpName string
	//队伍
	TheTeam map[string][]string
}

// 混战
type ScrimmagePvp struct {
	StarPrice  int `json:"开启场地金额"`
	EnterPrice int `json:"入场金额"`
	//利息
	Interest    float64  `json:"利息"`
	EndCmds     []string `json:"结束时执行指令"`
	RandomTpPos [][]int  `json:"场地传送点"`
	//初始生存时间
	InitTime int `json:"初始生存时间(秒)"`
	//游戏总时间
	AllTime int `json:"游戏总时间(分钟)"`
	//据点
	StrongHold StrongHold `json:"据点"`
	//范围
	StartPos []int `json:"场地起点"`
	RangePos []int `json:"场地范围"`
}

// 据点
type StrongHold struct {
	Pos     [][]int `json:"据点坐标"`
	AddTime int     `json:"据点每次加时间"`
	//据点刷新周期
	CycleTime int `json:"据点刷新周期"`
	//据点刷新指令
	CycleCmds []string `json:"据点刷新执行指令"`
	//接近据点时执行
	ApproachCmds []string `json:"当玩家接近据点时执行指令"`
	//玩家占领执行
	OccupiedCmds []string `json:"玩家占领据点执行指令"`
	//最大人数
	MaxNum int `json:"场地人数限制"`
	//观众点位
	WatchPos []int `json:"观众点位"`
}

// 单挑
type Solo struct {
	OpenTime   int `json:"房间存在时间(分钟)"`
	StarPrice  int `json:"开启金额"`
	EnterPrice int `json:"入场金额"`
	//利息
	Interest    float64  `json:"利息"`
	RandomTpPos [][]int  `json:"传送点随机坐标"`
	WinCmds     []string `json:"胜利执行指令"`
	StartPos    []int    `json:"场地起点"`
	RangePos    []int    `json:"场地范围"`
}

// 多人对战
type Multiplayer struct {
	RandomTpPos [][]int  `json:"随机传送坐标"`
	StarPrice   int      `json:"开启费用"`
	EnterPrice  int      `json:"入场费用"`
	WatchPos    []int    `json:"观众点位"`
	ExistTime   int      `json:"游戏存在时间(分钟)"`
	StartPos    []int    `json:"场地起点"`
	RangePos    []int    `json:"场地范围"`
	WinCmds     []string `json:"胜利执行指令"`
}

func (o *Pvp) Init(cfg *defines.ComponentConfig) {
	marshal, _ := json.Marshal(cfg.Configs)
	if err := json.Unmarshal(marshal, o); err != nil {
		panic(err)
	}
	o.Data = make(map[string]*DataOfPvp)
	o.PlayerTime = make(map[string]int)
	o.StrongHoldData = make(map[string][]int)
	if cfg.Version == "0.0.1" {
		delete(cfg.Configs, "激活密码")
		cfg.Description = "免费的pvp组件"
		cfg.Version = "0.0.2"
		cfg.Upgrade()
	}
}

const (
	//单挑类型名字
	SOLOTYPE = "Solo"
	//多人对战类型名字
	MULTYPE = "Mul"
	//混战类型名字
	SCITYPE = "Sci"
)

func (o *Pvp) Inject(frame defines.MainFrame) {
	o.Frame = frame
	o.BasicComponent.Inject(frame)
	o.Frame.GetGameListener().SetGameMenuEntry(&defines.GameMenuEntry{
		MenuEntry: defines.MenuEntry{
			Triggers:     o.Triggers,
			ArgumentHint: "",
			FinalTrigger: false,
			Usage:        o.Usage,
		},
		OptionalOnTriggerFn: o.onMenu,
	})
	o.Frame.GetGameListener().SetGameChatInterceptor(o.onChat)
}
func (b *Pvp) onChat(chat *defines.GameChat) bool {
	isIn, isMaster, roomName := b.CheckInRoom(chat.Name)
	if !isIn && len(chat.Msg) > 0 && (chat.Msg[0] == b.FastCmds["开启游戏"] || chat.Msg[0] == b.FastCmds["退出房间"]) {
		Sayto(b.Frame, chat.Name, "你没有加入任何房间")
		return false
	}
	if len(chat.Msg) > 0 {
		if chat.Msg[0] == b.FastCmds["开启游戏"] {
			//检查是否为房主
			if !isMaster {
				Sayto(b.Frame, chat.Name, "你不是房主")
				return true
			}
			//检查房间是否开启状态
			if b.Data[roomName].IsOpen {
				Sayto(b.Frame, chat.Name, "房间已经开始状态不能重复开启")
				return true
			}
			b.StarGame(roomName)
		} else if chat.Msg[0] == b.FastCmds["退出房间"] {
			//房主再提醒一下
			if isMaster {
				Sayto(b.Frame, chat.Name, "确定你要退出？退出则房间自动解散\n 确定请输入 0")
				b.Frame.GetGameControl().SetOnParamMsg(chat.Name, func(chat *defines.GameChat) (catch bool) {
					if len(chat.Msg) > 0 {
						if chat.Msg[0] == "0" {
							b.QuitRoom(chat.Name, roomName)
							Sayto(b.Frame, chat.Name, "退出成功")
						}
					}
					return true
				})
			} else {
				b.QuitRoom(chat.Name, roomName)
				Sayto(b.Frame, chat.Name, "退出成功")
			}

		}
	}

	return false
}

// 命令发送者
func (b *Pvp) CmdSender(str string) {
	b.Frame.GetGameControl().SendCmdAndInvokeOnResponse(str, func(output *packet.CommandOutput) {
		if output.SuccessCount > 0 {

		} else {
			pterm.Info.Printfln("指令错误 错误信息为:%v\n错误指令为:%v", output.OutputMessages, str)
		}
	})

}

// 退出房间
func (b *Pvp) QuitRoom(name string, roomName string) {
	isIn, isMaster, _ := b.CheckInRoom(name)
	if isMaster {
		for k, _ := range b.Data[roomName].Member {
			Sayto(b.Frame, k, "房间已被删除")
		}
		delete(b.Data, roomName)
	} else if isIn {
		b.Data[roomName].Pool -= b.Data[roomName].EnterPrice
		b.CmdSender(fmt.Sprintf("scoreboard players add @a[name=\"%v\"] %v %v", name, b.Score["购买所需计分板"], b.Data[roomName].EnterPrice))
		delete(b.Data[roomName].Member, name)
	}
}

// 随机传送
func (b *Pvp) GameTp(roomName string, thePosList [][]int) {
	rand.Seed(time.Now().Unix())
	thepos := thePosList[rand.Intn(len(thePosList))]
	b.CmdSender(fmt.Sprintf("tp @a[name=\"%v\"] %v %v %v", b.Data[roomName].Master, thepos[0], thepos[1], thepos[2]))
	num := 1
	for k, _ := range b.Data[roomName].Member {
		rand.Seed(time.Now().Unix() + int64(num))
		thepos = thePosList[rand.Intn(len(thePosList))]
		b.CmdSender(fmt.Sprintf("tp @a[name=\"%v\"] %v %v %v", k, thepos[0], thepos[1], thepos[2]))
		num++
	}
}

// 开启游戏
func (b *Pvp) StarGame(roomName string) bool {
	TheRoomType := b.Data[roomName].RoomType
	//检查人数
	relist := map[string]interface{}{
		"房间名": roomName,
		"类型":  b.TranslationType(TheRoomType),
	}
	if !(len(b.Data[roomName].Member)+1 >= 2) {
		Sayto(b.Frame, b.Data[roomName].Master, "人数不够 请至少满足2人")
		return false
	}
	//混战开启
	if TheRoomType == SCITYPE {
		if b.CheckRoom([]string{SOLOTYPE, MULTYPE}) {
			Sayto(b.Frame, b.Data[roomName].Master, "有房间处于开启状态\n不能开启")
			return false
		}
		//传送人员
		b.GameTp(roomName, b.ScrimmagePvp[b.Data[roomName].ThePvpName].RandomTpPos)
		//房间开启

		b.Frame.GetGameControl().SayTo("@a", FormateMsg(b.Frame, relist, b.TitleWord["游戏开启时提示词"]))
		//初始化生存时间分数
		b.PlayerTime = make(map[string]int)
		//初始化
		b.PlayerTime[b.Data[roomName].Master] = b.ScrimmagePvp[b.Data[roomName].ThePvpName].InitTime
		for k, _ := range b.Data[roomName].Member {
			b.PlayerTime[k] = b.ScrimmagePvp[b.Data[roomName].ThePvpName].InitTime
		}
		b.Data[roomName].IsOpen = true
	} else if TheRoomType == SOLOTYPE {
		//如果有模式开启了 则不开启
		if b.CheckRoom([]string{SCITYPE}) {
			Sayto(b.Frame, b.Data[roomName].Master, "有房间处于开启状态\n不能开启")
			return false
		}
		//传送人员
		b.GameTp(roomName, b.Solo.RandomTpPos)
		//房间开启
		b.Frame.GetGameControl().SayTo("@a", FormateMsg(b.Frame, relist, b.TitleWord["游戏开启时提示词"]))
		b.Data[roomName].IsOpen = true
	} else if TheRoomType == MULTYPE {
		if b.CheckRoom([]string{SCITYPE}) {
			Sayto(b.Frame, b.Data[roomName].Master, "有房间处于开启状态\n不能开启")
			return false
		}
		theTeamName := "一号队伍"
		b.Data[roomName].TheTeam = make(map[string][]string)
		b.Data[roomName].TheTeam[theTeamName] = append(b.Data[roomName].TheTeam[theTeamName], b.Data[roomName].Master)
		theTeamName = "二号队伍"
		//分配队伍
		for k, _ := range b.Data[roomName].Member {
			b.Data[roomName].TheTeam[theTeamName] = append(b.Data[roomName].TheTeam[theTeamName], k)
			if theTeamName == "一号队伍" {
				theTeamName = "二号队伍"
			} else {
				theTeamName = "一号队伍"
			}
		}
		//房间处于打开状态
		b.Data[roomName].IsOpen = true
		rand.Seed(time.Now().Unix())
		//传送玩家
		num := 0
		for _, v := range b.Data[roomName].TheTeam {
			//随机出一个
			thePvpName := b.Data[roomName].ThePvpName
			rand.Seed(time.Now().Unix() + int64(num))
			pterm.Info.Println("thePvpName", thePvpName)
			thePos := b.Multiplayer[thePvpName].RandomTpPos[rand.Intn(len(b.Multiplayer[thePvpName].RandomTpPos))]
			for _, member := range v {
				pterm.Info.Println("传送成功")
				b.CmdSender(fmt.Sprintf("tp @a[name=\"%v\"] %v %v %v", member, thePos[0], thePos[1], thePos[2]))
			}
			num++
		}
		b.Frame.GetGameControl().SayTo("@a", FormateMsg(b.Frame, relist, b.TitleWord["游戏开启时提示词"]))
	}
	return true
}
func (b *Pvp) onMenu(chat *defines.GameChat) (stop bool) {
	Sayto(b.Frame, chat.Name, b.Menu["主菜单"])
	b.Frame.GetGameControl().SetOnParamMsg(chat.Name, func(chat *defines.GameChat) (catch bool) {
		if len(chat.Msg) > 0 {
			switch chat.Msg[0] {
			case "0":
				b.SoloMenu(chat.Name)
			case "1":
				b.MultiplayerMenu(chat.Name)
			case "2":
				b.ScrimmageMenu(chat.Name)
				//Sayto(b.Frame, chat.Name, "未开放")
			case "3":
				b.GetBackDataOfRoom(chat.Name)
			case "4":
				b.ChoseRoom(chat.Name)
			}
		}

		return true
	})
	return true
}

// 返回对应的类型名字
func (b *Pvp) TranslationType(typename string) string {
	list := map[string]string{
		SOLOTYPE: "单挑",
		SCITYPE:  "混战",
		MULTYPE:  "多人对战",
	}
	return list[typename]
}

// 返回房间详细信息
func (b *Pvp) GetBackDataOfRoom(name string) {
	isIn, _, roomName := b.CheckInRoom(name)
	if isIn {
		memberListStr := ""
		for k, _ := range b.Data[roomName].Member {

			memberListStr += k + "\\"
		}
		relist := map[string]interface{}{
			"房主":   b.Data[roomName].Master,
			"成员列表": memberListStr,
			"游戏类型": b.TranslationType(b.Data[roomName].RoomType),
			"奖池金额": b.Data[roomName].Pool,
		}
		Sayto(b.Frame, name, FormateMsg(b.Frame, relist, b.Menu["房间详细信息"]))
	} else {
		Sayto(b.Frame, name, "你没有加入任何房间")
	}
}

// 加入房间
func (b *Pvp) ChoseRoom(name string) {
	num := 0
	msg := ""
	list := make(map[string]string)
	for k, v := range b.Data {
		relist := map[string]interface{}{
			"i":   num,
			"房间名": k,
			"类型":  b.TranslationType(v.RoomType),
		}
		msg += FormateMsg(b.Frame, relist, b.Menu["加入房间模板"]) + "\n"
		list[strconv.Itoa(num)] = k
		num++
	}
	Sayto(b.Frame, name, msg)
	b.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
		if len(chat.Msg) > 0 {
			if roomName, ok := list[chat.Msg[0]]; ok {
				memberListStr := ""
				for k, _ := range b.Data[roomName].Member {
					memberListStr += k + "\\"
				}
				relist := map[string]interface{}{
					"房主":   b.Data[roomName].Master,
					"成员列表": memberListStr,
					"游戏类型": b.TranslationType(b.Data[roomName].RoomType),
					"奖池金额": b.Data[roomName].Pool,
				}

				Sayto(b.Frame, name, FormateMsg(b.Frame, relist, b.Menu["房间详细信息"]))
				Sayto(b.Frame, name, b.TitleWord["加入房间提示词"])
				b.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
					if len(chat.Msg) > 0 {
						if chat.Msg[0] == "0" {
							//进入加入房间流程
							if b.Data[roomName].RoomType == SOLOTYPE && (len(b.Data[roomName].Member)+1) >= 2 {
								Sayto(b.Frame, name, "人数已满")
							} else {
								b.EnterRoom(name, roomName)
							}

						}
					}

					return true
				})

			} else {
				Sayto(b.Frame, name, "输入有效数字")
			}
		}
		return true
	})
}

// 进入房间
func (b *Pvp) EnterRoom(name string, roomName string) bool {
	isIn, isMaster, _ := b.CheckInRoom(name)
	if isIn && isMaster {
		Sayto(b.Frame, name, "你已经处于房间中")
		return false
	}
	cmd := b.ScoreboardPlayersRemove(name, b.Data[roomName].EnterPrice, b.Score["购买所需计分板"])
	b.Frame.GetGameControl().SendCmdAndInvokeOnResponse(cmd, func(output *packet.CommandOutput) {
		if output.SuccessCount > 0 {
			b.Data[roomName].Pool += b.Data[roomName].EnterPrice
			b.Data[roomName].Member[name] = "这是一个成员"
			Sayto(b.Frame, name, b.TitleWord["加入房间成功提示词"])
			relist := map[string]interface{}{
				"玩家名": name,
				"人数":  1 + len(b.Data[roomName].Member),
			}
			Sayto(b.Frame, b.Data[roomName].Master, FormateMsg(b.Frame, relist, b.TitleWord["提示房主有人进入"]))

		} else {
			b.WarnErrCmd(output.OutputMessages, cmd)
			b.WarnBalance(name, b.Data[roomName].EnterPrice)
		}
	})
	return true
}

// 混战
func (b *Pvp) ScrimmageMenu(name string) {
	if b.CheckRoom([]string{MULTYPE, SCITYPE, SOLOTYPE}) {
		Sayto(b.Frame, name, "已经有房间存在")
	} else {
		if isin, _, _ := b.CheckInRoom(name); isin {
			Sayto(b.Frame, name, "你已经在某个房间之中了")
		} else {
			list := make(map[string]string)
			num := 0
			msg := ""
			for k, _ := range b.ScrimmagePvp {
				relist := map[string]interface{}{
					"i":    num,
					"场地名字": k,
				}
				msg += FormateMsg(b.Frame, relist, b.Menu["选择场地模板"]) + "\n"
				list[strconv.Itoa(num)] = k
				num++
			}
			Sayto(b.Frame, name, msg)
			b.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
				if len(chat.Msg) > 0 {
					if roomName, ok := list[chat.Msg[0]]; ok {
						cmd := b.ScoreboardPlayersRemove(name, b.ScrimmagePvp[roomName].StarPrice, b.Score["购买所需计分板"])
						b.Frame.GetGameControl().SendCmdAndInvokeOnResponse(cmd, func(output *packet.CommandOutput) {
							if output.SuccessCount > 0 {
								Sayto(b.Frame, name, b.TitleWord["提示输入房间名字"])
								b.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
									if len(chat.Msg) > 0 {
										RoomName := chat.Msg[0]
										b.Data[RoomName] = &DataOfPvp{
											RoomType:   SCITYPE,
											Master:     name,
											Member:     make(map[string]string),
											IsOpen:     false,
											Time:       b.ScrimmagePvp[roomName].AllTime,
											Pool:       b.ScrimmagePvp[roomName].StarPrice,
											EnterPrice: b.ScrimmagePvp[roomName].EnterPrice,
											ThePvpName: roomName,
										}
										Sayto(b.Frame, name, b.TitleWord["创建房间提示词"])
									}
									return true
								})
							} else {
								b.WarnErrCmd(output.OutputMessages, cmd)
								b.WarnBalance(name, b.ScrimmagePvp[roomName].StarPrice)
							}
						})
					} else {
						Sayto(b.Frame, name, "输入有效数字")
					}
				}
				return true
			})
		}

	}
}

// 多人对战
func (b *Pvp) MultiplayerMenu(name string) {
	if b.CheckRoom([]string{MULTYPE, SCITYPE}) {
		Sayto(b.Frame, name, "已经有房间存在")
	} else {
		isin, _, _ := b.CheckInRoom(name)
		if isin {
			Sayto(b.Frame, name, "你已经存在在某个房间中了")
		} else {
			list := make(map[string]string)
			num := 0
			msg := ""
			for k, v := range b.Multiplayer {
				relist := map[string]interface{}{
					"i":    num,
					"场地名字": k,
					"价格":   v.StarPrice,
				}
				msg += FormateMsg(b.Frame, relist, b.Menu["选择场地模板"]) + "\n"
				list[strconv.Itoa(num)] = k
				num++
			}
			Sayto(b.Frame, name, msg)
			b.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
				if len(chat.Msg) > 0 {
					if roomName, ok := list[chat.Msg[0]]; ok {
						cmd := b.ScoreboardPlayersRemove(name, b.Multiplayer[roomName].StarPrice, b.Score["购买所需计分板"])
						b.Frame.GetGameControl().SendCmdAndInvokeOnResponse(cmd, func(output *packet.CommandOutput) {
							if output.SuccessCount > 0 {
								Sayto(b.Frame, name, b.TitleWord["提示输入房间名字"])
								b.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
									if len(chat.Msg) > 0 {
										RoomName := chat.Msg[0]
										b.Data[RoomName] = &DataOfPvp{
											RoomType:   MULTYPE,
											Master:     name,
											Member:     make(map[string]string),
											IsOpen:     false,
											Time:       b.Multiplayer[roomName].ExistTime,
											Pool:       b.Multiplayer[roomName].StarPrice,
											EnterPrice: b.Multiplayer[roomName].EnterPrice,
											ThePvpName: roomName,
										}
										Sayto(b.Frame, name, b.TitleWord["创建房间提示词"])
									}
									return true
								})
							} else {
								b.WarnErrCmd(output.OutputMessages, cmd)
								b.WarnBalance(name, b.Multiplayer[roomName].StarPrice)
							}
						})
					} else {
						Sayto(b.Frame, name, "输入有效数字")
					}
				}
				return true
			})
		}

	}

}

// 单挑菜单
func (b *Pvp) SoloMenu(name string) {
	if b.CheckRoom([]string{SCITYPE, SOLOTYPE}) {
		Sayto(b.Frame, name, "已经有房间存在")
	} else {
		isin, _, _ := b.CheckInRoom(name)
		if isin {
			Sayto(b.Frame, name, "你已经存在在某个房间中了")
		} else {
			cmd := b.ScoreboardPlayersRemove(name, b.Solo.StarPrice, b.Score["购买所需计分板"])
			b.Frame.GetGameControl().SendCmdAndInvokeOnResponse(cmd, func(output *packet.CommandOutput) {
				if output.SuccessCount > 0 {
					Sayto(b.Frame, name, b.TitleWord["提示输入房间名字"])
					b.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
						if len(chat.Msg) > 0 {
							RoomName := chat.Msg[0]

							b.Data[RoomName] = &DataOfPvp{
								RoomType:   SOLOTYPE,
								Master:     name,
								Member:     make(map[string]string),
								IsOpen:     false,
								Time:       b.Solo.OpenTime,
								Pool:       b.Solo.StarPrice,
								EnterPrice: b.Solo.EnterPrice,
							}
							pterm.Info.Println("存入结果:b.Data数据为:", b.Data)
							Sayto(b.Frame, name, b.TitleWord["创建房间提示词"])
						}
						return true
					})
				} else {
					b.WarnErrCmd(output.OutputMessages, cmd)
					b.WarnBalance(name, b.Solo.StarPrice)
				}
			})
		}

	}

}

func (o *Pvp) Activate() {
	go o.CheckGameIsOver()
}

// 返回在已开启game中 在领地范围内的人	map[房间名]在内的人
func (b *Pvp) GetInGamePlayer(playerPos map[string][]int) map[string][]string {
	list := make(map[string][]string)
	//pterm.Info.Println("playerPos:", playerPos)
	//num := 0
	//检查非法在内人员
	func() bool {
		for player, pos := range playerPos {
			if b.CheckIsIn(pos, b.Solo.StartPos, b.Solo.RangePos) {
				if isin, _, _ := b.CheckInRoom(player); !isin {
					b.CmdSender(fmt.Sprintf("tp @a[name=\"%v\"] %v %v %v", player, b.BackPos[0], b.BackPos[1], b.BackPos[2]))
					return true
				}
			}
			for pvpName, data := range b.Multiplayer {
				if b.CheckIsIn(pos, data.StartPos, data.RangePos) {
					if isin, _, roomName := b.CheckInRoom(player); !isin {
						if roomData, ok := b.Data[roomName]; ok && roomData.ThePvpName != pvpName {
							b.CmdSender(fmt.Sprintf("tp @a[name=\"%v\"] %v %v %v", player, b.BackPos[0], b.BackPos[1], b.BackPos[2]))
						} else if !ok {
							b.CmdSender(fmt.Sprintf("tp @a[name=\"%v\"] %v %v %v", player, b.BackPos[0], b.BackPos[1], b.BackPos[2]))
						}
					}
				}
			}
			for pvpName, data := range b.ScrimmagePvp {
				if b.CheckIsIn(pos, data.StartPos, data.RangePos) {
					if isin, _, roomName := b.CheckInRoom(player); !isin {
						if roomData, ok := b.Data[roomName]; ok && roomData.ThePvpName != pvpName {
							b.CmdSender(fmt.Sprintf("tp @a[name=\"%v\"] %v %v %v", player, b.BackPos[0], b.BackPos[1], b.BackPos[2]))
						} else if !ok {
							b.CmdSender(fmt.Sprintf("tp @a[name=\"%v\"] %v %v %v", player, b.BackPos[0], b.BackPos[1], b.BackPos[2]))
						}
					}
				}
			}
		}
		return false
	}()
	//返回人员
	for k, v := range b.Data {
		//检查在内的人
		if v.RoomType == SOLOTYPE {
			//先判断房主
			if b.CheckIsIn(playerPos[v.Master], b.Solo.StartPos, b.Solo.RangePos) {
				//pterm.Info.Println("在坐标内", v.Master)
				list[k] = append(list[k], v.Master)
			}
			//判断成员
			for memberName, _ := range v.Member {
				if b.CheckIsIn(playerPos[memberName], b.Solo.StartPos, b.Solo.RangePos) {
					//pterm.Info.Println("在坐标内", memberName)
					list[k] = append(list[k], memberName)
				}
			}
		} else if v.RoomType == MULTYPE {
			//先判断房主
			if b.CheckIsIn(playerPos[v.Master], b.Multiplayer[v.ThePvpName].StartPos, b.Multiplayer[v.ThePvpName].RangePos) {
				//pterm.Info.Println("在坐标内")
				list[k] = append(list[k], v.Master)
			}
			//判断成员
			for memberName, _ := range v.Member {
				if b.CheckIsIn(playerPos[memberName], b.Multiplayer[v.ThePvpName].StartPos, b.Multiplayer[v.ThePvpName].RangePos) {
					//pterm.Info.Println("在坐标内")
					list[k] = append(list[k], memberName)
				}
			}
		} else if v.RoomType == SCITYPE {
			if b.CheckIsIn(playerPos[v.Master], b.ScrimmagePvp[v.ThePvpName].StartPos, b.ScrimmagePvp[v.ThePvpName].RangePos) {
				//pterm.Info.Println("在坐标内")
				list[k] = append(list[k], v.Master)
			}
			//判断成员
			for memberName, _ := range v.Member {
				if b.CheckIsIn(playerPos[memberName], b.ScrimmagePvp[v.ThePvpName].StartPos, b.ScrimmagePvp[v.ThePvpName].RangePos) {

					list[k] = append(list[k], memberName)
				}
			}
		}
	}
	//pterm.Info.Println("list", list)
	return list
}

// 检查是否在范围内
func (b *Pvp) CheckIsIn(pos []int, starPos []int, Range []int) bool {
	//pterm.Info.Printfln("pos:%v,starpos:%v,range:%v", pos, starPos, Range)
	if pos[0] >= starPos[0] && pos[0] <= (starPos[0]+Range[0]) && pos[1] >= starPos[1] && pos[1] <= (starPos[1]+Range[1]) && pos[2] >= starPos[2] && pos[2] <= (starPos[2]+Range[2]) {
		return true
	}
	return false
}

// 判断结束器 为了方便设置的k与v
func (b *Pvp) judger(roomPlayerList map[string][]string, v DataOfPvp, k string) {
	if v.RoomType == SOLOTYPE {
		//判断为游戏结束

		if len(roomPlayerList[k]) == 1 {
			prizeNum := v.Pool - int(float64(v.Pool)*b.Solo.Interest)
			winer := roomPlayerList[k][0]
			relist := map[string]interface{}{
				"player": winer,
				"奖金":     prizeNum,
			}
			pterm.Info.Println("单挑游戏结束 通知:", FormateMsg(b.Frame, relist, b.TitleWord["单挑胜利提示词"]))
			Sayto(b.Frame, winer, FormateMsg(b.Frame, relist, b.TitleWord["单挑胜利提示词"]))
			for _, v := range b.Solo.WinCmds {
				relist = map[string]interface{}{
					"获胜者": winer,
					"计分板": b.Score["购买所需计分板"],
					"总金额": prizeNum,
				}

				cmd := FormateMsg(b.Frame, relist, v)
				pterm.Info.Println(cmd)
				b.CmdSender(cmd)
			}
			//b.CmdSender(fmt.Sprintf("scoreboard players add @a[name=\"%v\"] %v %v", winer, b.Score["购买所需计分板"], prizeNum))
			//删除房间
			delete(b.Data, k)
		} else if len(roomPlayerList[k]) < 1 && v.IsOpen {
			//如果人数小于 1 则自动删除房间
			pterm.Info.Println("删除房间")
			delete(b.Data, k)

		}

	} else if v.RoomType == SCITYPE {
		//混战
		//每秒减少
		for k, v := range b.PlayerTime {
			if v == 20 {
				Sayto(b.Frame, k, fmt.Sprintf("剩余时间:%v", v))
				b.PlayerTime[k]--
			} else if v <= 0 {
				isIn, isMaster, roomName := b.CheckInRoom(k)
				if isIn && !isMaster {
					delete(b.Data[roomName].Member, k)
				} else if isMaster {
					b.Data[roomName].Master = ""
				}
				relist := map[string]interface{}{
					"player": k,
				}
				Sayto(b.Frame, k, FormateMsg(b.Frame, relist, b.TitleWord["因为生存时间而退出"]))
				delete(b.PlayerTime, k)
			}

		}
		if _, ok := roomPlayerList[k]; ok && len(roomPlayerList) > 0 && len(roomPlayerList[k]) == 1 {
			prizeNum := int(float64(b.Data[k].Pool) - (float64(b.Data[k].Pool) * b.ScrimmagePvp[v.ThePvpName].Interest))
			pterm.Info.Println("奖金金额为:", prizeNum)
			relist := map[string]interface{}{
				"胜利者":  roomPlayerList[k][0],
				"奖金金额": prizeNum,
				"计分板":  b.Score["购买所需计分板"],
			}
			b.Frame.GetGameControl().SayTo("@a", FormateMsg(b.Frame, relist, b.TitleWord["多人混战胜利"]))
			for _, v := range b.ScrimmagePvp[v.ThePvpName].EndCmds {
				b.CmdSender(FormateMsg(b.Frame, relist, v))
			}
			pterm.Info.Println("删除DataOfPvp")
			delete(b.Data, k)
		}

	} else if v.RoomType == MULTYPE {
		//多人团队战
		if b.checkOneTeam(k, roomPlayerList, v.TheTeam) && v.IsOpen {
			prizeNum := v.Pool - int(float64(v.Pool)*b.Solo.Interest)
			relist := map[string]interface{}{
				"奖金": prizeNum,
			}
			pterm.Info.Println("多人游戏结束 通知:", FormateMsg(b.Frame, relist, b.TitleWord["多人胜利提示词"]))
			b.Frame.GetGameControl().SayTo("@a", FormateMsg(b.Frame, relist, b.TitleWord["多人胜利提示词"]))
			//Sayto(b.Frame, winer, FormateMsg(b.Frame, relist, b.TitleWord["多人胜利提示词"]))
			for _, winer := range roomPlayerList[k] {
				rePrize := int(prizeNum / len(roomPlayerList[k]))
				pterm.Info.Println("每人均分的奖金", rePrize)

				for _, v := range b.Multiplayer[v.ThePvpName].WinCmds {
					relist = map[string]interface{}{
						"获胜者": winer,
						"计分板": b.Score["购买所需计分板"],
						"总金额": prizeNum,
					}
					cmd := FormateMsg(b.Frame, relist, v)
					pterm.Info.Println(cmd)
					b.CmdSender(cmd)
				}
				//b.CmdSender(fmt.Sprintf("scoreboard players add @a[name=\"%v\"] %v %v", winer, b.Score["购买所需计分板"], prizeNum))
				//删除房间

			}
			delete(b.Data, k)
		} else if len(roomPlayerList[k]) == 0 && v.IsOpen {
			pterm.Info.Println("删除房间")
			delete(b.Data, k)
		}
	}

}

// 检查时间是否超时
func (b *Pvp) CheckTimeOver() {

	for {
		time.Sleep(time.Second * 60)
		for k, v := range b.Data {
			if v.IsOpen {
				//时间自减去

				v.Time--
				replaceList := map[string]interface{}{
					"房间名字": k,
					"剩余时间": v.Time,
				}
				if v.Time <= 0 && v.RoomType != SCITYPE {
					b.Frame.GetGameControl().SayTo("@a", FormateMsg(b.Frame, replaceList, "[房间名字] 超时 已遣送所有玩家回归"))
					delete(b.Data, k)
				} else if v.Time <= 0 && v.RoomType == SCITYPE {
					go func() {
						//排序
						posList := <-GetPos(b.Frame, "@a")
						roomPlayerList := b.GetInGamePlayer(posList)
						if _, ok := roomPlayerList[k]; ok && len(roomPlayerList[k]) >= 2 {
							theList := roomPlayerList[k]
							//更新一下playerTime里面的名单
							DicList := make(map[string]string)
							for _, _v := range theList {
								DicList[_v] = " "
							}
							//检查

							for _k, _ := range b.PlayerTime {
								//不在在线列表中则提出
								if _, isok := DicList[_k]; !isok {
									delete(b.PlayerTime, _k)
									if isin, isMaster, roomName := b.CheckInRoom(_k); isin && !isMaster {
										delete(b.Data[roomName].Member, _k)
									} else if isin && isMaster {
										b.Data[roomName].Master = ""
									}
								}
							}
							theTimeList := []int{}
							for _, time := range b.PlayerTime {
								theTimeList = append(theTimeList, time)
							}
							//降序排列
							sort.Sort(sort.Reverse(sort.IntSlice(theTimeList)))
							if len(theTimeList) >= 1 {
								winerNum := theTimeList[0]
								//根据数字找到对象
								for winer, num := range b.PlayerTime {
									if num == winerNum {
										prizeNum := int(float64(b.Data[k].Pool) * b.ScrimmagePvp[v.ThePvpName].Interest)
										pterm.Info.Println("奖金金额为:", prizeNum)
										relist := map[string]interface{}{
											"胜利者":  winer,
											"奖金金额": prizeNum,
											"计分板":  b.Score["购买所需计分板"],
										}
										b.Frame.GetGameControl().SayTo("@a", FormateMsg(b.Frame, relist, b.TitleWord["多人混战胜利"]))
										for _, v := range b.ScrimmagePvp[v.ThePvpName].EndCmds {
											b.CmdSender(FormateMsg(b.Frame, relist, v))
										}
										pterm.Info.Println("删除DataOfPvp")
										delete(b.Data, k)
										break
									}
								}
							} else {
								delete(b.Data, k)
							}

						}

					}()

				} else {

					b.Frame.GetGameControl().SayTo("@a", FormateMsg(b.Frame, replaceList, b.TitleWord["每分钟提示"]))
				}
			}

		}
	}

}

// 判断游戏是否结束
func (b *Pvp) judgeGameIsOver() {
	go func() {
		//每秒一次循环
		for {
			time.Sleep(time.Second * 1)
			//检查一次退出游戏的人

			go func() {
				posList := <-GetPos(b.Frame, "@a")
				roomPlayerList := b.GetInGamePlayer(posList)
				//检查在线人员

				for k, v := range b.Data {
					if v.IsOpen {
						roomPlayerDic := make(map[string]string)
						for _, member := range roomPlayerList[k] {
							roomPlayerDic[member] = "这是一个字典"
						}
						func() {
							//如果房主离开则删除房间
							if _, ok := roomPlayerDic[v.Master]; !ok {

								b.Frame.GetGameControl().SayTo("@a", fmt.Sprintf("房间因为房主离开 已删除"))
								defer delete(b.Data, k)
							}
							for member, _ := range v.Member {
								if _, ok := roomPlayerDic[member]; !ok {
									delete(b.Data[k].Member, member)
								}
							}
						}()
					}

				}
				//pterm.Info.Println("roomPlayerList:", roomPlayerList)
				for k, v := range b.Data {
					if v.IsOpen {
						b.judger(roomPlayerList, *v, k)
					}
					//已有据点检测玩家是否靠近
					if v.RoomType == SCITYPE && v.IsOpen {
						for player, pos := range posList {
							b.CheckIsApproach(player, pos, *v)
						}

					}

				}

			}()

		}
		//提取所有的房间
	}()
}
func (b *Pvp) CheckIsApproach(name string, pos []int, v DataOfPvp) {
	for StrongHoldName, spos := range b.StrongHoldData {
		if pos[0] == spos[0] && pos[2] == spos[2] {
			//代表站在上面
			b.PlayerTime[name] += b.ScrimmagePvp[v.ThePvpName].StrongHold.AddTime
			relist := map[string]interface{}{
				"player": name,
				"添加生存时间": b.ScrimmagePvp[v.ThePvpName].StrongHold.AddTime,
			}
			for _, cmd := range b.ScrimmagePvp[v.ThePvpName].StrongHold.OccupiedCmds {
				b.CmdSender(FormateMsg(b.Frame, relist, cmd))
			}
			pterm.Info.Println("删除据点数据成功")
			delete(b.StrongHoldData, StrongHoldName)
		} else if pos[0] > (spos[0]-5) && pos[0] < (spos[0]+5) && pos[1] > (spos[1]-5) && pos[1] < (spos[1]+5) && pos[2] > (spos[2]-5) && pos[2] < (spos[2]+5) {
			relist := map[string]interface{}{
				"据点坐标": fmt.Sprintf("%v %v %v", spos[0], spos[1], spos[2]),
			}
			for _, cmd := range b.ScrimmagePvp[v.ThePvpName].StrongHold.ApproachCmds {
				b.CmdSender(FormateMsg(b.Frame, relist, cmd))
			}
		}
	}

}

// 检查在现有的游戏是否结束
func (b *Pvp) CheckGameIsOver() {
	//判断游戏是否时间超时
	go b.CheckTimeOver()
	//检查游戏是否达到条件结束
	go b.judgeGameIsOver()
	go b.StrongHoldStar()
}

// 开启据点
func (b *Pvp) StrongHoldStar() {
	for {
		if ok, roomName := b.CheckIsSci(); ok {
			thePvpName := b.Data[roomName].ThePvpName
			posList := ""
			b.StrongHoldData = make(map[string][]int)
			for k, v := range b.ScrimmagePvp[thePvpName].StrongHold.Pos {
				posList += fmt.Sprintf("[%v %v %v]/", v[0], v[1], v[2])
				b.StrongHoldData[strconv.Itoa(k)] = v
			}
			reList := map[string]interface{}{
				"据点坐标列表": posList,
			}
			for _, cmd := range b.ScrimmagePvp[thePvpName].StrongHold.CycleCmds {
				b.CmdSender(FormateMsg(b.Frame, reList, cmd))
			}
			pterm.Info.Println(b.ScrimmagePvp[thePvpName].StrongHold.CycleTime)
			time.Sleep(time.Second * time.Duration(b.ScrimmagePvp[thePvpName].StrongHold.CycleTime))
		} else {
			time.Sleep(time.Second * 1)
		}

	}

}
func (b *Pvp) CheckIsSci() (isopen bool, roomName string) {
	for k, v := range b.Data {
		if v.RoomType == SCITYPE && v.IsOpen {
			return true, k
		}
	}
	return false, ""

}

// 检查是否只有一个队伍了
func (b *Pvp) checkOneTeam(roomName string, roomPlayerList map[string][]string, theTeam map[string][]string) bool {
	list := make(map[string]string)
	for _, v := range roomPlayerList[roomName] {
		list[v] = " "
	}
	//teamlist := make(map[string]string)

	//遍历两个队伍
	for _, v := range theTeam {
		teamMemberNum := 0
		for _, memberName := range v {
			if _, ok := list[memberName]; ok {
				teamMemberNum++
			}
		}
		if teamMemberNum == 0 {
			return true
		}
	}
	return false

}

// 检查当前是否有哪些类型(符合一点即可)
func (b *Pvp) CheckRoom(list []string) bool {
	TheRoomType := b.GetRoomType()
	//检查传来列表
	fmt.Println("theroomtype", TheRoomType)
	for _, v := range list {
		//只要一个符合就返回true
		if _, ok := TheRoomType[v]; ok {
			return true
		}
	}
	return false
}

// 检查当前房间 并返回类型map
func (b *Pvp) GetRoomType() map[string]string {
	list := make(map[string]string)
	pterm.Info.Println("b.Data:", b.Data)
	if len(b.Data) > 0 {
		for _, v := range b.Data {
			pterm.Info.Println("v.roomtype:", v.RoomType)
			list[v.RoomType] = "这是一个房间类型"
		}
	}
	return list

}

// 警告错误指令
func (b *Pvp) WarnErrCmd(str, cmd interface{}) {
	pterm.Error.Println("错误指令 错误信息如下:%v\n  错误指令:%v\n如果是余额不足可忽略", str, cmd)
}

// 获取扣分指令
func (b *Pvp) ScoreboardPlayersRemove(name, num, score interface{}) string {
	relist := map[string]interface{}{
		"player": name,
		"score":  score,
		"price":  num,
	}
	cmd := "scoreboard players remove @a[name=\"[player]\",scores={[score]=[price]..}] [score] [price]"
	return FormateMsg(b.Frame, relist, cmd)
}

// 警告余额不足
func (b *Pvp) WarnBalance(name string, num interface{}) {
	relist := map[string]interface{}{
		"金额": num,
	}
	Sayto(b.Frame, name, FormateMsg(b.Frame, relist, b.TitleWord["余额不足提示词"]))
}

// 检查是否在房间内 且返回是否为房主 房间名字
func (b *Pvp) CheckInRoom(name string) (isIn bool, isMaster bool, roomName string) {
	for k, v := range b.Data {
		if v.Master == name {
			return true, true, k
		}
		if _, ok := v.Member[name]; ok {
			return true, false, k
		}
	}
	return false, false, ""
}
