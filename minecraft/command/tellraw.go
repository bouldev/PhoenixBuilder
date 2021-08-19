package command

import (
	"fmt"
	"phoenixbuilder/minecraft/mctype"
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

func TellRawRequest(target mctype.Target, lines ...string) string {
	now := time.Now().Format("§6[15:04:05]§b")
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
	return SendSizukanaCommand(TellRawRequest(mctype.AllPlayers, lines...), conn)
}