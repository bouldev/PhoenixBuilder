package external

import (
	"fmt"
	"phoenixbuilder/fastbuilder/environment"
	"phoenixbuilder/fastbuilder/external/connection"
	"phoenixbuilder/fastbuilder/external/packet"
)

type ExternalConnectionHandler struct {
	listener      *connection.KCPConnectionServerHandler
	env           *environment.PBEnvironment
	PacketChannel chan []byte
}

func (handler *ExternalConnectionHandler) acceptConnection(conn connection.ReliableConnection) {
	env := handler.env
	allAlive := true
	go func() {
		for {
			gamePacket := <-handler.PacketChannel
			if !allAlive {
				return
			}
			packet.SerializeAndSend(&packet.GamePacket{
				Content: gamePacket,
			}, conn)
			select {
			case handler.PacketChannel <- gamePacket:
			default:
			}
			// Send the packet to the next receiver(connection)
		}
	}()
	go func() {
		for {
			rawPacket, err := conn.RecvFrame()
			if err != nil || !allAlive {
				allAlive = false
				return
			}
			pkt, canParse := packet.Deserialize(rawPacket)
			if !canParse {
				packet.SerializeAndSend(&packet.PacketViolationWarningPacket{
					Text: "Unparsable packet received!",
				}, conn)
			}
			switch p := pkt.(type) {
			case *packet.PingPacket:
				packet.SerializeAndSend(&packet.PongPacket{}, conn)
			case *packet.PongPacket:
				break
			case *packet.ByePacket:
				packet.SerializeAndSend(&packet.ByePacket{}, conn)
			case *packet.PacketViolationWarningPacket:
				break
			case *packet.EvalPBCommandPacket:
				handler.env.FunctionHolder.Process(p.Command)
			case *packet.GameCommandPacket:
				if p.CommandType == packet.CommandTypeSettings {
					env.CommandSender.SendSizukanaCommand(p.Command)
					break
				} else if p.CommandType == packet.CommandTypeNormal {
					env.CommandSender.SendWSCommand(p.Command, p.UUID)
				} else {
					env.CommandSender.SendCommand(p.Command, p.UUID)
				}
			}
		}
	}()
}

func (_ *ExternalConnectionHandler) acceptConnectionFail(err error) {
	fmt.Printf("External connection handler: Failed to accept a connection: %v\n", err)
}

func (_ *ExternalConnectionHandler) downError(_ interface{}) {
	fmt.Printf("ERROR: External connection handler's server stopped unexpectedly.\n")
}

func ListenExt(env *environment.PBEnvironment, address string) {
	handlerStruct := &ExternalConnectionHandler{
		listener:      &connection.KCPConnectionServerHandler{},
		PacketChannel: make(chan []byte),
		env:           env,
	}
	env.ExternalConnectionHandler = handlerStruct
	env.Destructors = append(env.Destructors, func() {
		close(handlerStruct.PacketChannel)
	})
	listener := handlerStruct.listener
	err := listener.Listen(address)
	if err != nil {
		fmt.Printf("Failed to listen on address %s: %v\n", address, err)
		return
	}
	listener.SetOnNewConnection(handlerStruct.acceptConnection)
	listener.SetOnAcceptNewConnectionFail(handlerStruct.acceptConnectionFail)
	listener.SetOnServerDown(handlerStruct.downError)
	fmt.Printf("Listening for external connection on address %s\n", address)
}
