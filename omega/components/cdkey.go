package components

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"time"
)

type cdKeyRecord struct {
	Total               int         `json:"可领取次数"`
	AllowRetake         bool        `json:"可重复领取"`
	DeviceIDCheck       bool        `json:"是否检查设备ID"`
	CmdsIn              interface{} `json:"指令"`
	TotalTaken          int
	Cmds                []defines.Cmd
	PlayerUUIDTaken     map[string][]*cdKeyTakenRecord
	PlayerDeviceIDTaken map[string][]*cdKeyTakenRecord
}

type cdKeyTakenRecord struct {
	Time string
	Name string
}

type CDkey struct {
	*defines.BasicComponent
	Usage               string                  `json:"菜单提示"`
	HintOnInvalid       string                  `json:"兑换码无效时提示"`
	HintOnRetake        string                  `json:"不可重复兑换时提示"`
	HintOnDeviceRetake  string                  `json:"不可同设备重复兑换时提示"`
	HintOnRateLimit     string                  `json:"领取次数到达上限时提示"`
	HintOnRequireInput  string                  `json:"要求输入兑换码时提示"`
	FileName            string                  `json:"兑换码领取记录文件"`
	Triggers            []string                `json:"触发词"`
	CDKeys              map[string]*cdKeyRecord `json:"兑换码"`
	fileChange          bool
	needConvertDataFile bool
}

func (o *CDkey) Init(cfg *defines.ComponentConfig) {
	if cfg.Version == "0.0.1" {
		cfg.Configs["要求输入兑换码时提示"] = "请输入兑换码"
		cfg.Configs["不可同设备重复兑换时提示"] = "当前设备已经领取过了，不能重复领取"
		for key := range cfg.Configs["兑换码"].(map[string]any) {
			cfg.Configs["兑换码"].(map[string]any)[key].(map[string]any)["是否检查设备ID"] = false
		}
		cfg.Version = "0.0.2"
		cfg.Upgrade()
		o.needConvertDataFile = true
	}
	m, _ := json.Marshal(cfg.Configs)
	if err := json.Unmarshal(m, o); err != nil {
		panic(err)
	}
	var err error
	for _, r := range o.CDKeys {
		r.PlayerUUIDTaken = make(map[string][]*cdKeyTakenRecord)
		r.PlayerDeviceIDTaken = make(map[string][]*cdKeyTakenRecord)
		if r.Cmds, err = utils.ParseAdaptiveCmd(r.CmdsIn); err != nil {
			panic(err)
		}
		if r.Total < 0 {
			fmt.Println(r, " 的可领取次数不能为负数，如果希望能无限次领取可以设为0")
		}
	}
}

func (o *CDkey) doRedeem(player string, cmds []defines.Cmd, current, total int) {
	res := "无限"
	totalS := fmt.Sprintf("%d", total)
	if total == 0 {
		totalS = "无限"
	} else {
		res = fmt.Sprintf("%d", total-current)
	}
	mapping := map[string]interface{}{
		"[player]":  "\"" + player + "\"",
		"[current]": current,
		"[total]":   totalS,
		"[res]":     res,
	}
	go utils.LaunchCmdsArray(o.Frame.GetGameControl(), cmds, mapping, o.Frame.GetBackendDisplay())
}

func (o *CDkey) redeem(chat *defines.GameChat) bool {
	// 没有输入兑换码时, 要求输入
	if len(chat.Msg) <= 0 {
		if o.Frame.GetGameControl().SetOnParamMsg(chat.Name, func(c *defines.GameChat) (catch bool) {
			o.redeem(c)
			return true
		}) == nil {
			o.Frame.GetGameControl().SayTo(chat.Name, o.HintOnRequireInput)
		}
		return true
	}
	key := chat.Msg[0]
	if redeem, hasK := o.CDKeys[key]; hasK {
		if redeem.Total > 0 && redeem.Total == redeem.TotalTaken {
			o.Frame.GetGameControl().SayTo(chat.Name, o.HintOnRateLimit)
			return true
		}
		uq := o.Frame.GetGameControl().GetPlayerKit(chat.Name).GetRelatedUQ()
		if uq == nil {
			fmt.Println("Cannot obatin player relatedUQ")
			return true
		}
		deviceID := uq.DeviceID
		// 如果开启了设备ID检查但没有获取到设备ID, 此次兑换将不再进行
		if redeem.DeviceIDCheck && deviceID == "" {
			fmt.Println("Cannot obatin player deviceID")
			return true
		}
		// 如果成功获取到UQ, 应该是必定带有uuid的
		uuid := uq.UUID.String()
		uuidTakes, hasUUIDkey := redeem.PlayerUUIDTaken[uuid]
		if !hasUUIDkey {
			uuidTakes = []*cdKeyTakenRecord{}
		}
		if !redeem.AllowRetake && len(uuidTakes) > 0 {
			o.Frame.GetGameControl().SayTo(chat.Name, o.HintOnRetake)
			return true
		}
		if deviceID != "" {
			deviceIdTakes, hasdeviceIDkey := redeem.PlayerDeviceIDTaken[deviceID]
			if !hasdeviceIDkey {
				deviceIdTakes = []*cdKeyTakenRecord{}
			}
			// 如果不允许重复领取且开启了设备ID检查, 会拒绝此设备ID再次兑换CDK
			if !redeem.AllowRetake && redeem.DeviceIDCheck && len(deviceIdTakes) > 0 {
				o.Frame.GetGameControl().SayTo(chat.Name, o.HintOnDeviceRetake)
				return true
			}
			redeem.PlayerDeviceIDTaken[deviceID] = append(deviceIdTakes, &cdKeyTakenRecord{
				Time: utils.TimeToString(time.Now()),
				Name: chat.Name,
			})
		}
		redeem.PlayerUUIDTaken[uuid] = append(uuidTakes, &cdKeyTakenRecord{
			Time: utils.TimeToString(time.Now()),
			Name: chat.Name,
		})
		redeem.TotalTaken++
		o.Frame.GetBackendDisplay().Write(fmt.Sprintf("玩家 %v 兑换 %v (%v/%v)", chat.Name, key, redeem.TotalTaken, redeem.Total))
		o.fileChange = true
		o.doRedeem(chat.Name, redeem.Cmds, redeem.TotalTaken, redeem.Total)
	} else {
		o.Frame.GetGameControl().SayTo(chat.Name, o.HintOnInvalid)
	}
	return true
}

func (o *CDkey) Inject(frame defines.MainFrame) {
	o.Frame = frame
	playerTaken := map[string]map[string]map[string][]*cdKeyTakenRecord{}
	// 数据文件转换 (0.0.1 -> 0.0.2)
	if o.needConvertDataFile {
		oldPlayerTaken := map[string]map[string][]*cdKeyTakenRecord{}
		if err := frame.GetJsonData(o.FileName, &oldPlayerTaken); err != nil {
			panic(err)
		}
		for cdkey, cdkeyData := range oldPlayerTaken {
			playerTaken[cdkey] = make(map[string]map[string][]*cdKeyTakenRecord)
			playerTaken[cdkey]["UUIDs"] = make(map[string][]*cdKeyTakenRecord)
			playerTaken[cdkey]["DevideIDs"] = make(map[string][]*cdKeyTakenRecord)
			for uuid, redeemDetails := range cdkeyData {
				playerTaken[cdkey]["UUIDs"][uuid] = redeemDetails
			}
			playerTaken[cdkey]["DevideIDs"] = map[string][]*cdKeyTakenRecord{}
		}
		o.Frame.WriteJsonDataWithTMP(o.FileName, ".ckpt", playerTaken)
	} else {
		err := frame.GetJsonData(o.FileName, &playerTaken)
		if err != nil {
			panic(err)
		}
	}
	// 初始化与计数, 设备ID部分不参与计数
	for cdkey, cdkeyData := range playerTaken {
		if r, hasK := o.CDKeys[cdkey]; hasK {
			r.PlayerUUIDTaken = cdkeyData["UUIDs"]
			r.PlayerDeviceIDTaken = cdkeyData["DeviceIDs"]
			totalTaken := 0
			for _, p := range cdkeyData["UUIDs"] {
				totalTaken += len(p)
			}
			r.TotalTaken = totalTaken
		}
	}
	o.Frame.GetGameListener().SetGameMenuEntry(&defines.GameMenuEntry{
		MenuEntry: defines.MenuEntry{
			Triggers:     o.Triggers,
			ArgumentHint: "[兑换码]",
			FinalTrigger: false,
			Usage:        o.Usage,
		},
		OptionalOnTriggerFn: o.redeem,
	})
}

func (o *CDkey) saveToDataFile() error {
	if o.fileChange {
		playerTaken := map[string]map[string]map[string][]*cdKeyTakenRecord{}
		for cdKey, record := range o.CDKeys {
			playerTaken[cdKey] = make(map[string]map[string][]*cdKeyTakenRecord)
			playerTaken[cdKey]["UUIDs"] = make(map[string][]*cdKeyTakenRecord)
			playerTaken[cdKey]["DeviceIDs"] = make(map[string][]*cdKeyTakenRecord)
			playerTaken[cdKey]["UUIDs"] = record.PlayerUUIDTaken
			playerTaken[cdKey]["DeviceIDs"] = record.PlayerDeviceIDTaken
		}
		o.fileChange = false
		return o.Frame.WriteJsonDataWithTMP(o.FileName, ".ckpt", playerTaken)
	}
	return nil
}

func (o *CDkey) Signal(signal int) error {
	switch signal {
	case defines.SIGNAL_DATA_CHECKPOINT:
		return o.saveToDataFile()
	}
	return nil
}

func (o *CDkey) Stop() error {
	fmt.Println("正在保存 " + o.FileName)
	return o.saveToDataFile()
}
