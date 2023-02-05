package components

import (
	"encoding/json"
	"fmt"
	"os"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"time"

	"github.com/pterm/pterm"
)

type Capture struct {
	*defines.BasicComponent
	FileName     string   `json:"文件名"`
	PacketType   []string `json:"目标数据包类型"`
	packetTypeID []uint32
}

func (o *Capture) Init(cfg *defines.ComponentConfig, storage defines.StorageAndLogProvider) {
	o.BasicComponent.Init(cfg, storage)
	m, _ := json.Marshal(cfg.Configs)
	if err := json.Unmarshal(m, o); err != nil {
		panic(err)
	}
	for _, packetType := range o.PacketType {
		if packetTypeID, hasK := utils.PktIDMapping[packetType]; !hasK {
			panic(fmt.Errorf("no such packet %v", o.PacketType))
		} else {
			o.packetTypeID = append(o.packetTypeID, uint32(packetTypeID))
		}
	}

}

func (o *Capture) Inject(frame defines.MainFrame) {
	o.Frame = frame
	if fp, err := os.OpenFile(o.FileName, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 07555); err != nil {
		panic(err)
	} else {
		go func() {
			for range time.NewTicker(time.Second).C {
				fp.Sync()
			}
		}()

		for i, packetTypeID := range o.packetTypeID {
			name := o.PacketType[i]
			o.Frame.GetGameListener().SetOnTypedPacketCallBack(packetTypeID, func(p packet.Packet) {
				if v, err := json.Marshal(p); err != nil {
					pterm.Error.Println(err)
				} else {
					fp.Write([]byte(name + "\n"))
					fp.Write(v)
					fp.Write([]byte("\n"))
				}
			})
		}
	}

}

// scoreboard objectives setdisplay list time
// scoreboard objectives setdisplay list time2
