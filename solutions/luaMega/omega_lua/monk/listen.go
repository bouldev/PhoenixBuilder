package monk

import (
	"context"
	"math/rand"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/solutions/luaMega/omega_lua/mux_pumper"
	"time"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/google/uuid"
)

var inputPumperMux *mux_pumper.InputPumperMux
var gamePacketPumperMux *mux_pumper.GamePacketPumperMux

func startInputSource() {
	inputPumperMux = mux_pumper.NewInputPumperMux()
	go func() {
		for {
			time.Sleep(time.Second * 2)
			input := "hello: " + uuid.New().String()
			inputPumperMux.PumpInput(input)
		}
	}()
}

func startGamePacketSource() {
	gamePacketPumperMux = mux_pumper.NewGamePacketPumperMux()
	go func() {
		for {
			<-time.After(time.Second / 10)
			// random decide a packet
			var pk packet.Packet
			pkType := rand.Intn(3)
			switch pkType {
			case 0:
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
				pk = &packet.Text{
					TextType:   packet.TextTypeChat,
					Message:    "hello: " + uuid.New().String(),
					SourceName: "monk",
				}
			case 2:
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
			gamePacketPumperMux.PumpGamePacket(pk)
		}
	}()
}

func init() {
	go startInputSource()
	go startGamePacketSource()
}

type MonkListen struct {
	packetQueueSize int
}

func NewMonkListen(packetQueueSize int) *MonkListen {
	return &MonkListen{
		packetQueueSize: packetQueueSize,
	}
}

func (m *MonkListen) UserInputChan() <-chan string {
	return inputPumperMux.NewListener()
}

func (m *MonkListen) MakeMCPacketFeeder(ctx context.Context, wants []string) <-chan packet.Packet {
	feeder := make(chan packet.Packet, m.packetQueueSize)
	pumper := mux_pumper.MakeMCPacketNoBlockFeeder(ctx, feeder)
	gamePacketPumperMux.AddNewPumper(wants, pumper)
	return feeder
}

func (m *MonkListen) GetMCPacketNameIDMapping() mux_pumper.MCPacketNameIDMapping {
	return gamePacketPumperMux.GetMCPacketNameIDMapping()
}
