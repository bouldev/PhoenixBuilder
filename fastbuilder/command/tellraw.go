package command

import (
	"fmt"
	"phoenixbuilder/fastbuilder/types"
	"phoenixbuilder/minecraft"
	"time"
	//"github.com/google/uuid"
	"encoding/json"
	"strings"
)

type TellrawItem struct {
	Text string `json:"text"`
}

type TellrawStruct struct {
	RawText []TellrawItem `json:"rawtext"`
}

func TellRawRequest(target types.Target, lines ...string) string {
	now := time.Now().Format("§6{15:04:05}§b")
	var items []TellrawItem
	for _, text := range lines {
		msg := fmt.Sprintf("%v %v", now, strings.Replace(text, "schematic", "sc***atic", -1))
		items=append(items,TellrawItem{Text:msg})
	}
	final := &TellrawStruct {
		RawText: items,
	}
	content, _ := json.Marshal(final)
	cmd := fmt.Sprintf("tellraw %v %s", target, content)
	return cmd
}

func Tellraw(conn *minecraft.Conn, lines ...string) error {
	//uuid1, _ := uuid.NewUUID()
	fmt.Printf("%s\n", lines[0])
	//return nil
	msg := strings.Replace(lines[0], "schematic", "sc***atic", -1)
	msg =  strings.Replace(msg, ".", "．", -1)
	// Netease set .bdx, .schematic, .mcacblock, etc as blocked words
	// So we should replace half-width points w/ full-width points to avoid being
	// blocked
	//return SendChat(fmt.Sprintf("§b%s",msg), conn)
	return SendSizukanaCommand(TellRawRequest(types.AllPlayers, lines...), conn)
}

func RawTellRawRequest(target types.Target, line string) string {
	var items []TellrawItem
	msg := strings.Replace(line, "schematic", "sc***atic", -1)
	items=append(items,TellrawItem{Text:msg})
	final := &TellrawStruct {
		RawText: items,
	}
	content, _ := json.Marshal(final)
	cmd := fmt.Sprintf("tellraw %v %s", target, content)
	return cmd
}

func WorldChatTellraw(conn *minecraft.Conn, sender string, content string) error {
	fmt.Printf("W <%s> %s\n", sender, content)
	str:=fmt.Sprintf("§eW §r<%s> %s",sender,content)
	return SendSizukanaCommand(RawTellRawRequest(types.AllPlayers, str), conn)
}