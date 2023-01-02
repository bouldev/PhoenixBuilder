package command

import (
	"io"
	"encoding/binary"
	"phoenixbuilder/fastbuilder/types"
)

type PlaceBlockWithChestData struct {
	BlockConstantStringID uint16
	BlockData uint16
	ChestSlots []types.ChestSlot
}

func (_ *PlaceBlockWithChestData) ID() uint16 {
	return 40
}

func (_ *PlaceBlockWithChestData) Name() string {
	return "PlaceBlockWithChestDataCommand"
}

func (cmd *PlaceBlockWithChestData) Marshal(writer io.Writer) error {
	uint16_buf:=make([]byte, 2)
	binary.BigEndian.PutUint16(uint16_buf, cmd.BlockConstantStringID)
	_, err:=writer.Write(uint16_buf)
	if err!=nil {
		return err
	}
	binary.BigEndian.PutUint16(uint16_buf, cmd.BlockData)
	_, err=writer.Write(uint16_buf)
	if err!=nil {
		return err
	}
	_, err=writer.Write([]byte{uint8(len(cmd.ChestSlots))})
	for _, slot := range cmd.ChestSlots {
		binary.BigEndian.PutUint16(uint16_buf, slot.Damage)
		tmpbuf:=append([]byte(slot.Name), []byte{0, slot.Count}...)
		tmpbuf=append(tmpbuf, append(uint16_buf, slot.Slot)...)
		_, err=writer.Write(tmpbuf)
		if err!=nil {
			return err
		}
	}
	return nil
}

func (cmd *PlaceBlockWithChestData) Unmarshal(reader io.Reader) error {
	buf:=make([]byte, 4)
	_, err:=io.ReadAtLeast(reader, buf, 4)
	if err != nil {
		return err
	}
	cmd.BlockConstantStringID=binary.BigEndian.Uint16(buf[0:2])
	cmd.BlockData=binary.BigEndian.Uint16(buf[2:])
	chestSlotsLenBuf:=make([]byte, 1)
	_, err = io.ReadAtLeast(reader, chestSlotsLenBuf, 1)
	if err!=nil {
		return err
	}
	chestSlotsLen:=int(chestSlotsLenBuf[0])
	cmd.ChestSlots=make([]types.ChestSlot, chestSlotsLen)
	for i:=0;i<chestSlotsLen;i++ {
		item_name, err:=readString(reader)
		if err!=nil {
			return err
		}
		cmd.ChestSlots[i].Name=item_name
		cds_buf:=make([]byte, 4)
		_, err=io.ReadAtLeast(reader, cds_buf, 4)
		if err!=nil {
			return err
		}
		cmd.ChestSlots[i].Count=cds_buf[0]
		cmd.ChestSlots[i].Damage=binary.BigEndian.Uint16(cds_buf[1:3])
		cmd.ChestSlots[i].Slot=cds_buf[3]
	}
	return nil
}