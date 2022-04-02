// +build do_not_build_this

package external

import (
	script_engine_bridge "phoenixbuilder/fastbuilder/script_engine/bridge"
	"phoenixbuilder/fastbuilder/external/connection"
	"phoenixbuilder/fastbuilder/external/packet"
	"phoenixbuilder/fastbuilder/environment"
	"phoenixbuilder/fastbuilder/function"
	"crypto/elliptic"
	"crypto/ecdsa"
	"crypto/rand"
	"strings"
	"net"
)

type ExternalConnectionHandler struct {
	listener *connection.KCPConnectionServerHandler
	env *environment.PBEnvironment
}

func (handler *ExternalConnectionHandler) acceptConnection(conn connection.ReliableConnection) {
	go func() {
		for {
			rawPacket, err:=conn.RecvFrame()
			if(err!=nil) {
				return
			}
			packet, canParse:=packet.Deserialize(rawPacket)
			if !canParse {
				conn.SendFrame(packet.PacketViolationWarningPacket {
					Text: "Unparsable packet received!",
				}.Serialize())
			}
			switch p:=packet.(type) {
			case packet.PingPacket:
				conn.SendFrame(packet.PongPacket{}.Serialize())
			case packet.PongPacket:
				break
			case packet.ByePacket:
				conn.SendFrame(packet.ByePacket{}.Serialize())
			case packet.PacketViolationWarningPacket:
				break
			case packet.EvalPBCommandPacket:
				fh:=handler.env.FunctionHolder.(*function.FunctionHolder)
				fh.Process(p.Command)
			case packet.GameCommandPacket:
				gc:=handler.env.Connection.(*minecraft.Conn)
				
		}
	} ()
}

func (_ *ExternalConnectionHandler) acceptConnectionFail(err error) {
	fmt.Printf("External connection handler: Failed to accept a connection: %v\n",err)
}

func (_ *ExternalConnectionHandler) downError(_ interface{}) {
	fmt.Printf("ERROR: External connection handler's server stopped unexpectedly.\n")
}

func ListenExt(env *environment.PBEnvironment, address string) {
	handlerStruct:=&ExternalConnectionHandler {
		listener: &connection.KCPConnectionServerHandler{}
	}
	listener:=handlerStruct.listener
	err:=listener.Listen(address)
	if(err!=nil) {
		fmt.Printf("Failed to listen on address %s: %v\n",address,err)
		return
	}
	listener.SetOnNewConnection(listener.acceptConnection)
	listener.SetOnAcceptNetConnectionFail(listener.acceptConnectionFail)
	listener.SetOnServerDown(listener.downError)
	fmt.Printf("Listening for external connection on address %s\n",address)
}