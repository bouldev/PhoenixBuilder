package core

import (
	"bytes"
	"phoenixbuilder/fastbuilder/lib/minecraft/neomega/omega"
	"phoenixbuilder/minecraft"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
)

type InteractCore struct {
	*minecraft.Conn
}

func (i *InteractCore) SendPacket(packet packet.Packet) {
	i.WritePacket(packet)
}

func (i *InteractCore) SendPacketBytes(packetID uint32, packet []byte) {
	i.WritePacketByte(packetID, packet)
}

func NewInteractCore(conn *minecraft.Conn) omega.InteractCore {
	return &InteractCore{Conn: conn}
}

type RemoteInteractCoreSlow struct {
	sendPacket func(pk packet.Packet)
}

func (i *RemoteInteractCoreSlow) SendPacket(pk packet.Packet) {
	i.sendPacket(pk)
}

func (i *RemoteInteractCoreSlow) SendPacketBytes(packetID uint32, packet []byte) {
	panic("not implemented")
}

func NewRemoteInteractCore(sendPacket func(pk packet.Packet)) *RemoteInteractCoreSlow {
	return &RemoteInteractCoreSlow{sendPacket: sendPacket}
}

type RemoteInteractCoreDirect struct {
	sendPacketBytes func(packetID uint32, packet []byte)
	shieldID        int32
	//sendPacket func(pk packet.Packet)
}

func (i *RemoteInteractCoreDirect) SendPacket(pk packet.Packet) {
	writer := bytes.NewBuffer(nil)
	w := protocol.NewWriter(writer, i.shieldID)
	pk.Marshal(w)
	i.sendPacketBytes(pk.ID(), writer.Bytes())
}

func (i *RemoteInteractCoreDirect) SendPacketBytes(packetID uint32, packet []byte) {
	i.sendPacketBytes(packetID, packet)
}

func (i *RemoteInteractCoreDirect) UpdateShieldID(id int32) {
	i.shieldID = id
}

func NewRemoteInteractCoreDirect(sendPacketBytes func(packetID uint32, packet []byte)) *RemoteInteractCoreDirect {
	return &RemoteInteractCoreDirect{sendPacketBytes: sendPacketBytes}
}
