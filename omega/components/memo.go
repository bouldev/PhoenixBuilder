package components

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/omega/collaborate"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"strings"
	"time"
)

type Memo struct {
	*BasicComponent
	logger            defines.LineDst
	HintOnEmptyPlayer string   `json:"没有指定玩家时提示"`
	HintOnEmptyMsg    string   `json:"没有输入信息时提示"`
	Response          string   `json:"留言成功时提示"`
	FileName          string   `json:"留言记录文件"`
	LogFile           string   `json:"日志文件"`
	Triggers          []string `json:"触发词"`
	LoginDelay        int      `json:"登录时延迟发送"`
	Memos             map[string][]string
	Usage             string `json:"提示信息"`
	PlayerSearcher    collaborate.FUNC_GetPossibleName
}

func (me *Memo) send(playerName string) {
	if msgs, hasK := me.Memos[playerName]; hasK {
		if len(msgs) > 0 {
			if player := me.Frame.GetGameControl().GetPlayerKit(playerName); player != nil {
				player.Title("有新留言")
				player.SubTitle("查看聊天栏")
				for _, m := range msgs {
					player.Say(m)
					me.logger.Write("send to " + playerName + " " + m)
				}
				delete(me.Memos, playerName)
			}
		} else {
			delete(me.Memos, playerName)
		}
	}
}

func (me *Memo) save(chat *defines.GameChat) bool {
	dstPlayer := chat.Msg[0]
	msg := strings.Join(chat.Msg[1:], " ")
	me.logger.Write(fmt.Sprintf("[%v]->[%v]:%v ", chat.Type, chat.Name, msg))
	m := utils.FormateByRepalcment(me.Response, map[string]interface{}{
		"[src_player]": chat.Name,
		"[dst_player]": dstPlayer,
		"[msg]":        msg,
	})

	me.Frame.GetGameControl().SendCmd(m)
	if _, hasK := me.Memos[dstPlayer]; !hasK {
		me.Memos[dstPlayer] = make([]string, 0)
	}
	me.Memos[dstPlayer] = append(me.Memos[dstPlayer],
		fmt.Sprintf("你有一条来自 %v 的留言: %v", chat.Name, msg),
	)
	for _, p := range me.Frame.GetUQHolder().PlayersByEntityID {
		if p.Username == dstPlayer {
			me.send(dstPlayer)
		}
	}
	return true
}

func (me *Memo) askForMsg(chat *defines.GameChat) {
	dstPlayer := chat.Msg[0]
	if player := me.Frame.GetGameControl().GetPlayerKit(chat.Name); player != nil {
		if player.SetOnParamMsg(func(c *defines.GameChat) bool {
			c.Msg = utils.InsertHead[string](dstPlayer, c.Msg)
			me.save(c)
			return true
		}) == nil {
			me.Frame.GetGameControl().SayTo(chat.Name, me.HintOnEmptyMsg)
		}
	}
}

func (me *Memo) askForPlayer(chat *defines.GameChat) {
	if player := me.Frame.GetGameControl().GetPlayerKit(chat.Name); player != nil {
		if player.SetOnParamMsg(func(c *defines.GameChat) bool {
			me.record(c)
			return true
		}) == nil {
			me.Frame.GetGameControl().SayTo(chat.Name, me.HintOnEmptyPlayer)
		}
	}
}

func (me *Memo) isValidName(name string) {

}

func (me *Memo) record(chat *defines.GameChat) bool {
	if len(chat.Msg) >= 2 {
		return me.save(chat)
	}
	if len(chat.Msg) == 1 {
		me.askForMsg(chat)
		return true
	}
	if len(chat.Msg) == 0 {
		me.askForPlayer(chat)
		return true
	}
	return false
}

func (me *Memo) Init(cfg *defines.ComponentConfig) {
	m, _ := json.Marshal(cfg.Configs)
	if err := json.Unmarshal(m, me); err != nil {
		panic(err)
	}
}

func (me *Memo) Inject(frame defines.MainFrame) {
	me.Frame = frame
	me.logger = &utils.MultipleLogger{Loggers: []defines.LineDst{
		me.Frame.GetLogger(me.LogFile),
		me.Frame.GetBackendDisplay(),
	}}
	me.Frame.GetGameListener().AppendLoginInfoCallback(func(entry protocol.PlayerListEntry) {
		name := utils.ToPlainName(entry.Username)
		if _, hasK := me.Memos[name]; hasK {
			timer := time.NewTimer(time.Duration(me.LoginDelay) * time.Second)
			go func() {
				<-timer.C
				me.send(name)
			}()
		}
	})
	if me.Usage == "" {
		me.Usage = "给某个玩家留言，将在上线时转达留言"
	}
	me.Frame.GetGameListener().SetGameMenuEntry(&defines.GameMenuEntry{
		MenuEntry: defines.MenuEntry{
			Triggers:     me.Triggers,
			ArgumentHint: "[玩家] [消息]",
			Usage:        me.Usage,
			FinalTrigger: false,
		},
		OptionalOnTriggerFn: me.record,
	})
	me.Memos = map[string][]string{}
	err := frame.GetJsonData(me.FileName, &me.Memos)
	if err != nil {
		panic(err)
	}
	me.PlayerSearcher = (*frame.GetContext())[collaborate.INTERFACE_POSSIBLE_NAME].(collaborate.FUNC_GetPossibleName)
}

func (me *Memo) Stop() error {
	fmt.Printf("正在保存 %v\n", me.FileName)
	return me.Frame.WriteJsonData(me.FileName, me.Memos)
}
