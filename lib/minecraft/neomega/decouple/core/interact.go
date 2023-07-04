package core

import (
	"phoenixbuilder/lib/minecraft/neomega/omega"
	"phoenixbuilder/minecraft"
	"phoenixbuilder/minecraft/protocol/packet"
)

type InteractCore struct {
	*minecraft.Conn
}

func (i *InteractCore) SendPacket(packet packet.Packet) {
	i.WritePacket(packet)
}

//	func (i *InteractCore) SendPacketBytes(packetID uint32, packet []byte) {
//		i.WritePacketBytes(packetID, packet)
//	}
func NewInteractCore(conn *minecraft.Conn) omega.InteractCore {
	return &InteractCore{Conn: conn}
}

type RemoteInteractCore struct {
	//sendPacketBytes func(packetID uint32, packet []byte)
	//shieldID        int32
	sendPacket func(pk packet.Packet)
}

func (i *RemoteInteractCore) SendPacket(pk packet.Packet) {
	i.sendPacket(pk)
	//writer := bytes.NewBuffer(nil)
	//w := protocol.NewWriter(writer, i.shieldID)
	//pk.Marshal(w)
	//i.sendPacketBytes(pk.ID(), writer.Bytes())

}

//func (i *RemoteInteractCore) SendPacketBytes(packetID uint32, packet []byte) {
//	//i.sendPacketBytes(packetID, packet)
//}
//
//func (i *RemoteInteractCore) UpdateShieldID(id int32) {
//	//i.shieldID = id
//}

func NewRemoteInteractCore(sendPacket func(pk packet.Packet)) *RemoteInteractCore {
	return &RemoteInteractCore{sendPacket: sendPacket}
}
