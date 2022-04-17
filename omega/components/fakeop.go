package components

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/pterm/pterm"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"strings"
)

type FakeOp struct {
	*BasicComponent
	AuthFile string `json:"授权文件"`
	Auth     map[string]map[string][]string
}

func (o *FakeOp) hasPermission(name string, cmdT string) []string {
	if p, hasK := o.Auth[name]; hasK {
		if a, hasK := p[cmdT]; hasK {
			if a == nil || len(a) == 0 {
				return nil
			}
			return a
		}
	}
	if p, hasK := o.Auth["*"]; hasK {
		if a, hasK := p[cmdT]; hasK {
			if a == nil || len(a) == 0 {
				return nil
			}
			return a
		}
	}
	return nil
}

func (o *FakeOp) Init(cfg *defines.ComponentConfig) {
	m, _ := json.Marshal(cfg.Configs)
	err := json.Unmarshal(m, o)
	if err != nil {
		panic(err)
	}
}

func (o *FakeOp) onChat(chat *defines.GameChat) bool {
	if len(chat.Msg) == 0 {
		return false
	}
	cmd := chat.Msg[0]
	tmps := o.hasPermission(chat.Name, cmd)
	if tmps == nil {
		return false
	}
	args := strings.Join(chat.Msg[1:], " ")
	for _, tmp := range tmps {
		c := utils.FormateByRepalcment(tmp, map[string]interface{}{
			"[player]": chat.Name,
			"[args]":   args,
		})
		o.Frame.GetBackendDisplay().Write(fmt.Sprintf("%v@%v: %v", chat.Name, cmd, c))
		o.Frame.GetGameControl().SendCmd(c)
	}
	return true
}

//go:embed default_fakeop.json
var defaultFakeOP []byte

func (o *FakeOp) Inject(frame defines.MainFrame) {
	o.Frame = frame
	if !utils.IsFile(o.Frame.GetRelativeFileName(o.AuthFile)) {
		pterm.Warning.Printf("没有检测到伪OP权限文件,将在 %v 下展开默认权限文件\n", o.Frame.GetRelativeFileName(o.AuthFile))
		err := o.Frame.WriteFileData(o.AuthFile, defaultFakeOP)
		if err != nil {
			panic(err)
		}
	}
	err := o.Frame.GetJsonData(o.AuthFile, &o.Auth)
	if err != nil {
		panic(err)
	}
	if o.Auth == nil {
		o.Auth = map[string]map[string][]string{}
	}
	o.Frame.GetGameListener().SetGameChatInterceptor(o.onChat)
}
