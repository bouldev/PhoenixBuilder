package monk

import (
	"context"
	"math/rand"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/solutions/omega_lua/omega_lua/mux_pumper"
	"time"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/google/uuid"
)

// 游戏包的管理中心 负责分发gamePacket插件
var gamePacketPumperMux *mux_pumper.GamePacketPumperMux

// 开启游戏包的资源分配中心
func startGamePacketSource() {
	// 创建一个新的 gamePacketPumperMux 实例
	gamePacketPumperMux = mux_pumper.NewGamePacketPumperMux()
	go func() {
		for {
			<-time.After(time.Second / 10)
			// random decide a packet
			// 随机决定生成哪种类型的数据包
			var pk packet.Packet
			pkType := rand.Intn(3)
			switch pkType {
			case 0:
				// 生成移动玩家数据包
				pk = &packet.MovePlayer{
					EntityRuntimeID: 0,
					Position: mgl32.Vec3{
						rand.Float32(),
						rand.Float32(),
						rand.Float32(),
					},
					Yaw:      rand.Float32(),
					Pitch:    rand.Float32(),
					OnGround: rand.Intn(2) == 1,
				}
			case 1:
				// 生成聊天消息数据包
				pk = &packet.Text{
					TextType:   packet.TextTypeChat,
					Message:    "hello: " + uuid.New().String(),
					SourceName: "monk",
				}
			case 2:
				// 生成命令输出数据包
				pk = &packet.CommandOutput{
					OutputType:   packet.CommandOutputTypeDataSet,
					SuccessCount: uint32(rand.Intn(100)),
					OutputMessages: []protocol.CommandOutputMessage{
						protocol.CommandOutputMessage{
							Success:    true,
							Message:    "hello: " + uuid.New().String(),
							Parameters: []string{"1", "2", "3"},
						},
						protocol.CommandOutputMessage{
							Success:    true,
							Message:    "hello2: " + uuid.New().String(),
							Parameters: []string{"4", "5", "6"},
						},
					},
				}

			}
			//发送游戏包
			gamePacketPumperMux.PumpGamePacket(pk)
		}
	}()
}

// 在程序启动时调用 startInputSource 和 startGamePacketSource 启动资源分配中心
func init() {
	go startGamePacketSource()
}

// MonkPackets 表示一个游戏监听器，其中 packetQueueSize 表示数据包通道的缓冲区大小
type MonkPackets struct {
	packetQueueSize int
}

// NewMonkListen 创建一个新的 MonkListen 实例
func NewMonkPackets(packetQueueSize int) *MonkPackets {
	return &MonkPackets{
		packetQueueSize: packetQueueSize,
	}
}

// MakeMCPacketFeeder 根据给定的协议名称列表创建一个数据包通道，并返回该通道。该方法还会创建一个数据包提供者，并将其添加到游戏包的管理中心中
func (m *MonkPackets) MakeMCPacketFeeder(ctx context.Context, wants []string) <-chan packet.Packet {
	feeder := make(chan packet.Packet, m.packetQueueSize)
	pumper := mux_pumper.MakeMCPacketNoBlockFeeder(ctx, feeder)
	gamePacketPumperMux.AddNewPumper(wants, pumper)
	return feeder
}

// GetMCPacketNameIDMapping 返回游戏包的名称和 ID 的映射
func (m *MonkPackets) GetMCPacketNameIDMapping() mux_pumper.MCPacketNameIDMapping {
	return gamePacketPumperMux.GetMCPacketNameIDMapping()
}
