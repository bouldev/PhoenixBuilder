package external

import (
	"fmt"
	"phoenixbuilder/fastbuilder/environment"
	"phoenixbuilder/fastbuilder/external/connection"
	"phoenixbuilder/fastbuilder/external/packet"
	"phoenixbuilder/fastbuilder/uqHolder"
	"phoenixbuilder/minecraft"
	"time"
)

type ExternalConnectionHandler struct {
	listener              *connection.KCPConnectionServerHandler
	env                   *environment.PBEnvironment
	PacketChannel         chan []byte
	DistributeChannel     chan []byte
	LeaveConsumerChannel  chan int
	NewConsumerChannel    chan int
	AcceptConsumerChannel chan interface{}
}

func (handler *ExternalConnectionHandler) acceptConnection(conn connection.ReliableConnection) {
	env := handler.env
	allAlive := true
	handler.NewConsumerChannel <- (1)
	<-handler.AcceptConsumerChannel
	pingDeadline := time.Now().Add(time.Second * 3)
	bufferChan := make(chan []byte, 1024)
	go func() {
		for {
			gamePacket := <-bufferChan
			if !allAlive {
				return
			}
			// fmt.Println("send: ", gamePacket[0])
			packet.SerializeAndSend(&packet.GamePacket{
				Content: gamePacket,
			}, conn)
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
			pingDeadline = time.Now().Add(time.Second * 3)
			switch p := pkt.(type) {
			case *packet.PingPacket:
				pingDeadline = time.Now().Add(time.Second * 3)
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
			case *packet.GamePacket:
				(env.Connection).(*minecraft.Conn).Write(p.Content)
			case *packet.UQHolderRequestPacket:
				//q:=string(p.QueryString)
				//if q=="*"
				packet.SerializeAndSend(&packet.UQHolderResponsePacket{
					Content: (env.UQHolder).(*uqHolder.UQHolder).Marshal(),
				}, conn)
			}
		}
	}()
	go func() {
		for {
			// fmt.Println("buffering, now", len(bufferChan))
			pkt := <-handler.DistributeChannel
			if len(bufferChan) > 512 || pingDeadline.Before(time.Now()) {
				// fmt.Println("kick client")
				allAlive = false
			}
			if !allAlive {
				// fmt.Println("notify consumer leave")
				handler.LeaveConsumerChannel <- (-1)
				return
			}
			bufferChan <- pkt
			handler.LeaveConsumerChannel <- 0
		}
	}()
}

func (_ *ExternalConnectionHandler) acceptConnectionFail(err error) {
	fmt.Printf("External connection handler: Failed to accept a connection: %v\n", err)
}

func (_ *ExternalConnectionHandler) downError(_ interface{}) {
	fmt.Printf("ERROR: External connection handler's server stopped unexpectedly.\n")
}

func (e *ExternalConnectionHandler) epoll() {
	consumers := 0
	for {
		select {
		case <-e.NewConsumerChannel:
			consumers++
			e.AcceptConsumerChannel <- true
			// fmt.Println("consumers come, current: ", consumers)
		case packet := <-e.PacketChannel:
			if consumers > 0 {
				for i := 0; i < consumers; i++ {
					// fmt.Println("distributing...")
					e.DistributeChannel <- packet
				}
				consumerChange := 0
				for i := 0; i < consumers; i++ {
					// s := <-e.LeaveConsumerChannel
					// fmt.Println("recv consumer", i, " status ", s)
					consumerChange += <-e.LeaveConsumerChannel
				}
				consumers += consumerChange
				// if consumerChange != 0 {
				// 	fmt.Println("consumers leave, current: ", consumers)
				// }
			}
		}
	}
}

func ListenExt(env *environment.PBEnvironment, address string) {
	handlerStruct := &ExternalConnectionHandler{
		listener:              &connection.KCPConnectionServerHandler{},
		PacketChannel:         make(chan []byte, 0),
		DistributeChannel:     make(chan []byte),
		LeaveConsumerChannel:  make(chan int, 0),
		NewConsumerChannel:    make(chan int, 0),
		AcceptConsumerChannel: make(chan interface{}, 0),
		env:                   env,
	}
	go handlerStruct.epoll()
	env.ExternalConnectionHandler = handlerStruct
	env.Destructors = append(env.Destructors, func() {
		close(handlerStruct.PacketChannel)
		close(handlerStruct.LeaveConsumerChannel)
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
