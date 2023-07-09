package components

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/pterm/pterm"
)

type BanTime struct {
	*defines.BasicComponent
	OnOmegaTakeOverIn    interface{} `json:"Omega接管时指令"`
	KickCmdIn            interface{} `json:"踢出指令"`
	KickCmdForDeviceIn   interface{} `json:"设备封禁时踢出指令"`
	AfterOmegaTakeOverIn interface{} `json:"到达封禁时间Omega结束接管时指令"`
	Selector             string      `json:"选择器"`
	ScoreboardName       string      `json:"读取封禁时间的计分板名"`
	FileName             string      `json:"文件名"`
	LoginDelay           int         `json:"登录时延迟发送"`
	Duration             int         `json:"检查周期"`
	KickDelay            int         `json:"延迟踢出时间"`
	EnableDeviceCheck    bool        `json:"是否检查设备ID"`
	OnOmegaTakeOver      []defines.Cmd
	KickCmd              []defines.Cmd
	KickCmdForDevice     []defines.Cmd
	AfterOmegaTakeOver   []defines.Cmd
	fileChange           bool
	Data                 *data
	mu                   sync.Mutex
}

type data struct {
	OldData        map[string]string                 `json:"旧版本数据_0.0.1"`
	BannedUUID     map[uuid.UUID]*bannedUUIDDetails  `json:"玩家封禁数据"`
	BannedDeviceID map[string]*bannedDeviceIDDetails `json:"设备封禁数据"`
}

type bannedUUIDDetails struct {
	BanTime        string `json:"解封时间"`
	PlayerName     string `json:"封禁时玩家名"`
	PlayerDeviceID string `json:"封禁时设备ID"`
}

type bannedDeviceIDDetails struct {
	BanTime    string `json:"解封时间"`
	PlayerName string `json:"封禁时玩家名"`
}

func (o *BanTime) Init(cfg *defines.ComponentConfig, storage defines.StorageAndLogProvider) {
	if cfg.Version == "0.0.1" {
		// 转换数据文件
		filename := cfg.Configs["文件名"].(string)
		updateData := &data{}
		if err := storage.GetJsonData(filename, &updateData.OldData); err == nil {
			storage.WriteJsonDataWithTMP(filename, ".ckpt", updateData)
		}
		// 升级配置
		cfg.Configs["设备封禁时踢出指令"] = []string{"kick [player] 当前设备已被封禁\n剩余时间: [day]天[hour]时[min]分[sec]秒"}
		cfg.Configs["是否检查设备ID"] = false
		cfg.Configs["设备ID相关说明"] = "机器人仅能获取到附近玩家的设备ID, 建议在玩家上线时将机器人传送至其位置"
		cfg.Version = "0.0.2"
		cfg.Upgrade()
	}
	m, _ := json.Marshal(cfg.Configs)
	var err error
	if err = json.Unmarshal(m, o); err != nil {
		panic(err)
	}
	if o.OnOmegaTakeOver, err = utils.ParseAdaptiveCmd(o.OnOmegaTakeOverIn); err != nil {
		panic(err)
	}
	if o.KickCmd, err = utils.ParseAdaptiveCmd(o.KickCmdIn); err != nil {
		panic(err)
	}
	if o.AfterOmegaTakeOver, err = utils.ParseAdaptiveCmd(o.AfterOmegaTakeOverIn); err != nil {
		panic(err)
	}
	if o.EnableDeviceCheck {
		if o.KickCmdForDevice, err = utils.ParseAdaptiveCmd(o.KickCmdForDeviceIn); err != nil {
			panic(err)
		}
	}
	o.mu = sync.Mutex{}
}

func (o *BanTime) Inject(frame defines.MainFrame) {
	o.Frame = frame
	var err error

	o.mu.Lock()
	defer o.mu.Unlock()

	if err = frame.GetJsonData(o.FileName, &o.Data); err != nil {
		panic(err)
	}
	if o.Data == nil {
		o.Data = &data{}
	}
	if o.Data.OldData == nil {
		o.Data.OldData = make(map[string]string)
	}
	if o.Data.BannedUUID == nil {
		o.Data.BannedUUID = make(map[uuid.UUID]*bannedUUIDDetails)
	}
	if o.Data.BannedDeviceID == nil {
		o.Data.BannedDeviceID = make(map[string]*bannedDeviceIDDetails)
	}

	o.Frame.GetGameListener().AppendLoginInfoCallback(func(entry protocol.PlayerListEntry) {
		// 旧数据 + UUID = 新数据
		if value, ok := o.Data.OldData[entry.Username]; ok {
			o.Data.BannedUUID[entry.UUID] = &bannedUUIDDetails{
				PlayerName: entry.Username,
				BanTime:    value,
			}
			delete(o.Data.OldData, entry.Username)
			o.fileChange = true
		}
		// 处理UUID封禁
		if value, ok := o.Data.BannedUUID[entry.UUID]; ok {
			banTime, err := utils.StringToTimeWithLocal(value.BanTime + " +0800 CST")
			if err != nil {
				panic(err)
			}
			if banTime.After(time.Now()) {
				go func() {
					<-time.NewTimer(time.Duration(o.KickDelay) * time.Second).C
					o.Frame.GetBackendDisplay().Write(fmt.Sprintf("尝试踢出玩家：%v", entry.Username))
					o.kick(entry.Username, o.KickCmd, banTime)
				}()
			} else {
				delete(o.Data.BannedUUID, entry.UUID)
				o.fileChange = true
				go func() {
					<-time.NewTimer(time.Duration(o.LoginDelay) * time.Second).C
					utils.LaunchCmdsArray(o.Frame.GetGameControl(), o.AfterOmegaTakeOver, map[string]interface{}{
						"[player]": entry.Username,
					}, o.Frame.GetBackendDisplay())
				}()
			}
		}
	})
	// 处理DeviceID封禁
	if o.EnableDeviceCheck {
		o.Frame.GetGameListener().SetOnTypedPacketCallBack(packet.IDAddPlayer, func(p packet.Packet) {
			pkt := p.(*packet.AddPlayer)
			if value, ok := o.Data.BannedDeviceID[pkt.DeviceID]; ok {
				banTime, err := utils.StringToTimeWithLocal(value.BanTime + " +0800 CST")
				if err != nil {
					panic(err)
				}
				if banTime.After(time.Now()) {
					go func() {
						<-time.NewTimer(time.Duration(o.KickDelay) * time.Second).C
						o.Frame.GetBackendDisplay().Write(fmt.Sprintf("尝试踢出玩家：%v", pkt.Username))
						o.kick(pkt.Username, o.KickCmdForDevice, banTime)
					}()
				} else {
					o.mu.Lock()
					delete(o.Data.BannedDeviceID, pkt.DeviceID)
					o.mu.Unlock()
					o.fileChange = true
				}
			}
		})
	}
}

func (o *BanTime) Signal(signal int) (err error) {
	switch signal {
	case defines.SIGNAL_DATA_CHECKPOINT:
		if o.fileChange {
			err = o.Frame.WriteJsonDataWithTMP(o.FileName, ".ckpt", o.Data)
			o.fileChange = false
		}
	}
	return err
}

func (o *BanTime) Stop() error {
	fmt.Printf("正在保存 %v\n", o.FileName)
	return o.Frame.WriteJsonDataWithTMP(o.FileName, ".final", o.Data)
}

func (o *BanTime) kick(name string, cmd []defines.Cmd, banTime time.Time) {
	duration := time.Until(banTime)
	go utils.LaunchCmdsArray(o.Frame.GetGameControl(), cmd, map[string]interface{}{
		"[player]": name,
		"[day]":    int(duration.Hours()) / 24,
		"[hour]":   int(duration.Hours()) % 24,
		"[min]":    int(duration.Minutes()) % 60,
		"[sec]":    int(duration.Seconds()) % 60,
	}, o.Frame.GetBackendDisplay())
}

func (o *BanTime) takeOver(name string) {
	utils.GetPlayerScore(o.Frame.GetGameControl(), name, o.ScoreboardName, func(val int, err error) {
		if err != nil {
			pterm.Error.Printfln("无法获取封禁时间信息 %v %v %v", name, o.ScoreboardName, err)
		} else if val < 0 {
			pterm.Error.Printfln("封禁时间指令设计配置有问题，如果封禁时间小于等于 0，则不应该被选择器选中 %v %v", name, o.ScoreboardName)
		} else {
			playerUQ := o.Frame.GetGameControl().GetPlayerKit(name).GetRelatedUQ()
			duration := time.Second * time.Duration(val)
			banTime := time.Now().Add(duration)
			banTimeStr := utils.TimeToString(banTime)
			o.mu.Lock()
			defer o.mu.Unlock()
			// 写入UUID封禁数据
			o.Data.BannedUUID[playerUQ.UUID] = &bannedUUIDDetails{
				BanTime:        banTimeStr,
				PlayerName:     name,
				PlayerDeviceID: playerUQ.DeviceID,
			}
			// 写入DeviceID封禁数据
			if o.EnableDeviceCheck && playerUQ.DeviceID != "" {
				o.Data.BannedDeviceID[playerUQ.DeviceID] = &bannedDeviceIDDetails{
					BanTime:    banTimeStr,
					PlayerName: name,
				}
			}
			o.fileChange = true
			go func() {
				utils.LaunchCmdsArray(o.Frame.GetGameControl(), o.OnOmegaTakeOver, map[string]interface{}{
					"[player]": name,
					"[day]":    int(duration.Hours()) / 24,
					"[hour]":   int(duration.Hours()) % 24,
					"[min]":    int(duration.Minutes()) % 60,
					"[sec]":    int(duration.Seconds()) % 60,
				}, o.Frame.GetBackendDisplay())
				o.kick(name, o.KickCmd, banTime)
			}()
		}
	})
}
func (o *BanTime) Activate() {
	t := time.NewTicker(time.Second * time.Duration(o.Duration))
	for {
		o.Frame.GetGameControl().SendCmdAndInvokeOnResponse(fmt.Sprintf("testfor %v", o.Selector), func(output *packet.CommandOutput) {
			if output.SuccessCount > 0 && len(output.OutputMessages) > 0 {
				ban := &Banned{Victim: []string{}}
				err := json.Unmarshal([]byte(output.DataSet), &ban)
				if err != nil {
					o.Frame.GetBackendDisplay().Write(fmt.Sprintf("踢出玩家时遇到问题：%v", err.Error()))
				} else {
					o.Frame.GetBackendDisplay().Write(fmt.Sprintf("尝试踢出玩家：%v", ban.Victim))
					for _, v := range ban.Victim {
						o.takeOver(v)
					}
				}
			}
		})
		<-t.C
	}
}
