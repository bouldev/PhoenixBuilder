package Happy2018new

import (
	"bytes"
	"encoding/json"
	"fmt"
	"phoenixbuilder/GameControl/GlobalAPI"
	"phoenixbuilder/minecraft/nbt"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	"strings"
	"sync"
)

type ExportMCStructure struct {
	*defines.BasicComponent
	apis     GlobalAPI.GlobalAPI
	lockDown sync.Mutex
	Triggers []string `json:"菜单触发词"`
	Usage    string   `json:"菜单项描述"`
}

func (o *ExportMCStructure) Init(cfg *defines.ComponentConfig, storage defines.StorageAndLogProvider) {
	marshal, _ := json.Marshal(cfg.Configs)
	if err := json.Unmarshal(marshal, o); err != nil {
		panic(err)
	}
}

func (o *ExportMCStructure) Inject(frame defines.MainFrame) {
	o.Frame = frame
	o.apis = o.Frame.GetGameControl().GetInteraction()
	o.lockDown = sync.Mutex{}
	o.Frame.GetGameListener().SetGameMenuEntry(&defines.GameMenuEntry{
		MenuEntry: defines.MenuEntry{
			Triggers:     o.Triggers,
			ArgumentHint: "<名称: string> [保存路径: string]",
			FinalTrigger: false,
			Usage:        o.Usage,
		},
		OptionalOnTriggerFn: o.Runner,
	})
}

func (o *ExportMCStructure) Runner(chat *defines.GameChat) bool {
	go func() {
		if !o.lockDown.TryLock() {
			o.Frame.GetGameControl().SayTo(chat.Name, "§c请求过于频繁§f，§c请稍后再试")
			return
		}
		defer o.lockDown.Unlock()
		o.ExportMCStructure(chat)
	}()
	return true
}

func (o *ExportMCStructure) ExportMCStructure(chat *defines.GameChat) {
	if len(chat.Msg) < 1 {
		o.Frame.GetGameControl().SayTo(chat.Name, "§c提供的参数不足§f，§c请检查后重试")
		return
	}
	// check value
	holder := o.apis.Resources.Structure.Occupy()
	resp, err := o.apis.SendStructureRequestWithResponce(
		&packet.StructureTemplateDataRequest{
			StructureName: chat.Msg[0],
			RequestType:   packet.StructureTemplateRequestExportFromLoad,
		},
	)
	if err != nil {
		panic(fmt.Sprintf("ExportMCStructure: %v", err))
	}
	o.apis.Resources.Structure.Release(holder)
	// get mcstructure
	if !resp.Success {
		o.Frame.GetGameControl().SayTo(chat.Name, "§c请求导出的结构不存在")
		return
	}
	// check success states
	writer := bytes.NewBuffer([]byte{})
	err = nbt.NewEncoderWithEncoding(writer, nbt.LittleEndian).Encode(&resp.StructureTemplate)
	if err != nil {
		panic(fmt.Sprintf("ExportMCStructure: %v", err))
	}
	// get binary data of mcstructure
	if len(chat.Msg) < 2 {
		err = o.Frame.WriteFileData(
			fmt.Sprintf("%v.mcstructure", strings.ReplaceAll(chat.Msg[0], ":", "_")),
			writer.Bytes(),
		)
		if err != nil {
			panic(fmt.Sprintf("ExportMCStructure: %v", err))
		}
		o.Frame.GetGameControl().SayTo(chat.Name, fmt.Sprintf("§a已成功导出结构 §b%v §f，§a它已被保存到 §bomega_storage/data/%v.mcstructure", chat.Msg[0], strings.ReplaceAll(chat.Msg[0], ":", "_")))
	} else {
		err = o.Frame.WriteFileData(
			chat.Msg[1],
			writer.Bytes(),
		)
		if err != nil {
			panic(fmt.Sprintf("ExportMCStructure: %v", err))
		}
		o.Frame.GetGameControl().SayTo(chat.Name, fmt.Sprintf("§a已成功导出结构 §b%v §f，§a它已被保存到 §bomega_storage/data/%v", chat.Msg[0], chat.Msg[1]))
	}
	// write mcstructure to file
}
