package external

import (
	"fmt"
	"io"
	"phoenixbuilder/fastbuilder/environment"
	"phoenixbuilder/fastbuilder/external/connection"
	"phoenixbuilder/fastbuilder/external/packet"
	"phoenixbuilder/fastbuilder/uqHolder"
	GameInterface "phoenixbuilder/game_control/game_interface"
	"phoenixbuilder/minecraft"
	"phoenixbuilder/minecraft/protocol"
	"time"
)

type ExternalConnectionHandler struct {
	listener              *connection.ConnectionServerHandler
	env                   *environment.PBEnvironment
	PacketChannel         chan []byte
	DistributeChannel     chan []byte
	LeaveConsumerChannel  chan int
	NewConsumerChannel    chan int
	AcceptConsumerChannel chan interface{}
}
type NoEOFByteReader struct {
	s []byte
	i int
}

func (nbr *NoEOFByteReader) Read(b []byte) (n int, err error) {
	if len(b) == 0 {
		return 0, nil
	}
	if nbr.i >= len(nbr.s) {
		return 0, io.EOF
	}
	n = copy(b, nbr.s[nbr.i:])
	nbr.i += n
	return
}

func (nbr *NoEOFByteReader) ReadByte() (b byte, err error) {
	if nbr.i >= len(nbr.s) {
		return 0, io.EOF
	}
	b = nbr.s[nbr.i]
	nbr.i++
	return b, nil
}

func (handler *ExternalConnectionHandler) acceptConnection(conn connection.ReliableConnection) {
	env := handler.env
	allAlive := true
	handler.NewConsumerChannel <- (1)
	<-handler.AcceptConsumerChannel
	pingDeadline := time.Now().Add(time.Second * 5)
	bufferChan := make(chan []byte, 1024*8)
	clientPacketChan := make(chan []byte, 1024)
	skipMap := make(map[uint8]uint8)
	hitMap := make(map[uint8]uint8)
	setSkip := func(ID uint8, possib uint8) {
		skipMap[ID] = possib
		hitMap[ID] = 0
	}
	go func() {
		for {
			select {
			case clientPackets := <-clientPacketChan:
				pingDeadline = time.Now().Add(time.Second * 5)
				pkt, canParse := packet.Deserialize(clientPackets)
				if !canParse {
					packet.SerializeAndSend(&packet.PacketViolationWarningPacket{
						Text: "Unparsable packet received!",
					}, conn)
				}
				switch p := pkt.(type) {
				case *packet.PingPacket:
					pingDeadline = time.Now().Add(time.Second * 5)
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
						env.GameInterface.SendSettingsCommand(p.Command, false)
						break
					} else if p.CommandType == packet.CommandTypeNormal {
						sendCommand(
							env.GameInterface.(*GameInterface.GameInterface),
							p.Command,
							p.UUID,
							protocol.CommandOriginAutomationPlayer,
						)
					} else {
						sendCommand(
							env.GameInterface.(*GameInterface.GameInterface),
							p.Command,
							p.UUID,
							protocol.CommandOriginPlayer,
						)
					}
				case *packet.GamePacket:
					(env.Connection).(*minecraft.Conn).Write(p.Content)
				case *packet.GamePacketReducePacket:
					setSkip(p.PacketID, p.DropBy)
				case *packet.UQHolderRequestPacket:
					//q:=string(p.QueryString)
					//if q=="*"
					packet.SerializeAndSend(&packet.UQHolderResponsePacket{
						Content: (env.UQHolder).(*uqHolder.UQHolder).Marshal(),
					}, conn)
				}
			case gamePacket := <-bufferChan:
				if !allAlive {
					return
				}
				// fmt.Println("send: ", gamePacket[0])
				packet.SerializeAndSend(&packet.GamePacket{
					Content: gamePacket,
				}, conn)
			}
		}
	}()
	go func() {
		for {
			rawPacket, err := conn.RecvFrame()
			if err != nil || !allAlive {
				allAlive = false
				return
			}
			clientPacketChan <- rawPacket
		}
	}()
	go func() {
		//packetCount := 0
		//startTime := time.Now()
		for {
			//fmt.Println("buffering, now", len(bufferChan))
			//packetCount++
			//if packetCount%200 == 0 {
			//	fmt.Println(float32(packetCount) / float32(time.Now().Sub(startTime).Seconds()))
			//	fmt.Println(len(bufferChan))
			//	packetCount = 0
			//	startTime = time.Now()
			//}
			pkt := <-handler.DistributeChannel
			pktID := pkt[0]
			if t, hasK := skipMap[pktID]; hasK {
				hitMap[pktID]++
				if hitMap[pktID] == t {
					hitMap[pktID] = 0
				} else {
					handler.LeaveConsumerChannel <- 0
					continue
				}
			}
			if len(bufferChan) > 1024*7 || pingDeadline.Before(time.Now()) {
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
	//TypePool := mc_packet.NewPool()

	for {
		select {
		case <-e.NewConsumerChannel:
			consumers++
			e.AcceptConsumerChannel <- true
			// fmt.Println("consumers come, current: ", consumers)
		case pkt := <-e.PacketChannel:
			//pkFn, hasK := TypePool[uint32(pkt[0])]
			//
			//if !hasK {
			//	panic(fmt.Errorf("cannot decode packet %v", pkt[0]))
			//}
			//pk := pkFn()
			//pk.Unmarshal(protocol.NewReader(&NoEOFByteReader{s: pkt[1:]}, 0))
			if consumers > 0 {
				for i := 0; i < consumers; i++ {
					// fmt.Println("distributing...")
					e.DistributeChannel <- pkt
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
		listener:              &connection.ConnectionServerHandler{},
		PacketChannel:         make(chan []byte, 1024),
		DistributeChannel:     make(chan []byte, 1024),
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
