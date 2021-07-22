package command

import (
	"fmt"
	"phoenixbuilder/minecraft/mctype"
	"time"
)

func TellRawRequest(target mctype.Target, lines ...string) string {
	now := time.Now().Format("§6[15:04:05]§b")
	cmd := fmt.Sprintf(`tellraw %v {"rawtext":[`, target)
	for i, text := range lines {
		msg := fmt.Sprintf("%v %v", now, text)
		cmd += `{"text":"` + msg + `"}`
		if i != len(lines)-1 {
			cmd += `,`
		}
	}
	return cmd + `]}`
}