package uqholder

import (
	"bytes"
	"errors"
	"phoenixbuilder/fastbuilder/lib/encoding/binary_read_write"
	LE "phoenixbuilder/fastbuilder/lib/encoding/little_endian"
	"phoenixbuilder/fastbuilder/lib/minecraft/neomega/omega"
	"phoenixbuilder/minecraft"
	"phoenixbuilder/minecraft/protocol/packet"
)

func init() {
	if false {
		func(holder omega.MicroUQHolder) {}(&MicroUQHolder{})
	}
}

type MicroUQHolder struct {
	omega.BotBasicInfoHolder
}

func NewMicroUQHolder(conn *minecraft.Conn) *MicroUQHolder {
	uq := &MicroUQHolder{
		NewBotInfoHolder(conn),
	}
	return uq
}

func NewMicroUQHolderFromData(data []byte) (uq *MicroUQHolder, err error) {
	uq = &MicroUQHolder{}
	err = uq.Unmarshal(data)
	if err != nil {
		return nil, err
	}
	return uq, nil
}

func (u *MicroUQHolder) GetBotBasicInfo() omega.BotBasicInfoHolder {
	return u.BotBasicInfoHolder
}

func (u *MicroUQHolder) Marshal() (data []byte, err error) {
	basicWriter := bytes.NewBuffer(nil)
	writer := binary_read_write.WrapBinaryWriter(basicWriter)
	err = LE.WriteString(writer, "BotBasicInfoHolder")
	if err != nil {
		return nil, err
	}
	var subData []byte
	subData, err = u.BotBasicInfoHolder.Marshal()
	if err != nil {
		return nil, err
	}
	err = LE.WriteInt64(writer, int64(len(subData)))
	if err != nil {
		return nil, err
	}
	err = writer.Write(subData)
	if err != nil {
		return nil, err
	}
	return basicWriter.Bytes(), err
}

var ErrInvalidUQHolderEntry = errors.New("invalid uqholder entry")

func (u *MicroUQHolder) Unmarshal(data []byte) error {
	if u.BotBasicInfoHolder == nil {
		u.BotBasicInfoHolder = &BotBasicInfoHolder{}
	}
	basicReader := bytes.NewBuffer(data)
	reader := binary_read_write.WrapBinaryReader(basicReader)
	var name string
	name, err := LE.String(reader)
	if err != nil {
		return err
	}
	if name != "BotBasicInfoHolder" {
		return ErrInvalidUQHolderEntry
	}
	var subData []byte
	var subDataLen int64
	subDataLen, err = LE.Int64(reader)
	if err != nil {
		return err
	}
	subData, err = reader.ReadOut(int(subDataLen))
	if err != nil {
		return err
	}
	return u.BotBasicInfoHolder.Unmarshal(subData)
}

func (u *MicroUQHolder) UpdateFromPacket(packet packet.Packet) {
	u.BotBasicInfoHolder.UpdateFromPacket(packet)
}
