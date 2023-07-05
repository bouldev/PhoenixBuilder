package uqholder

import (
	"bytes"
	"phoenixbuilder/fastbuilder/lib/encoding/binary_read_write"
	LE "phoenixbuilder/fastbuilder/lib/encoding/little_endian"
	"phoenixbuilder/fastbuilder/lib/minecraft/neomega/omega"
	"phoenixbuilder/minecraft"
	"phoenixbuilder/minecraft/protocol/packet"
)

func init() {
	if false {
		func(omega.BotBasicInfoHolder) {}(&BotBasicInfoHolder{})
	}
}

type BotBasicInfoHolder struct {
	BotName      string
	BotRuntimeID uint64
	BotUniqueID  int64
	BotIdentity  string
}

func (b *BotBasicInfoHolder) Marshal() (data []byte, err error) {
	basicWriter := bytes.NewBuffer(nil)
	writer := binary_read_write.WrapBinaryWriter(basicWriter)
	err = LE.WriteString(writer, b.BotName)
	if err != nil {
		return nil, err
	}
	err = LE.WriteInt64(writer, int64(b.BotRuntimeID))
	if err != nil {
		return nil, err
	}
	err = LE.WriteInt64(writer, int64(b.BotUniqueID))
	if err != nil {
		return nil, err
	}
	err = LE.WriteString(writer, b.BotIdentity)
	if err != nil {
		return nil, err
	}
	return basicWriter.Bytes(), err
}

func (b *BotBasicInfoHolder) Unmarshal(data []byte) (err error) {
	basicReader := bytes.NewReader(data)
	reader := binary_read_write.WrapBinaryReader(basicReader)
	b.BotName, err = LE.String(reader)
	if err != nil {
		return err
	}
	var i64 int64
	i64, err = LE.Int64(reader)
	if err != nil {
		return err
	}
	b.BotRuntimeID = uint64(i64)
	i64, err = LE.Int64(reader)
	if err != nil {
		return err
	}
	b.BotUniqueID = i64
	b.BotIdentity, err = LE.String(reader)
	if err != nil {
		return err
	}
	return nil
}

func (b *BotBasicInfoHolder) UpdateFromPacket(packet packet.Packet) {
}

func (b *BotBasicInfoHolder) GetBotName() string {
	return b.BotName
}

func (b *BotBasicInfoHolder) GetBotRuntimeID() uint64 {
	return b.BotRuntimeID
}

func (b *BotBasicInfoHolder) GetBotUniqueID() int64 {
	return b.BotUniqueID
}

func (b *BotBasicInfoHolder) GetBotIdentity() string {
	return b.BotIdentity
}

func NewBotInfoHolder(conn *minecraft.Conn) omega.BotBasicInfoHolder {
	h := &BotBasicInfoHolder{}
	gd := conn.GameData()
	h.BotRuntimeID = gd.EntityRuntimeID
	h.BotUniqueID = gd.EntityUniqueID
	h.BotName = conn.IdentityData().DisplayName
	h.BotIdentity = conn.IdentityData().Identity
	return h
}
