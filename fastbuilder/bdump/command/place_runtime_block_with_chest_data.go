package command

import (
	"io"
	"encoding/binary"
	"phoenixbuilder/fastbuilder/types"
)

type PlaceRuntimeBlockWithChestData struct {
	BlockRuntimeID uint16
	ChestSlots []types.ChestSlot
}

func (_ *PlaceRuntimeBlockWithChestData) ID() uint16 {
	return 37
}

func (_ *PlaceRuntimeBlockWithChestData) Name() string {
	return "PlaceRuntimeBlockWithChestDataCommand"
}

func (cmd *PlaceRuntimeBlockWithChestData) Marshal(writer io.Writer) error {
	uint16_buf:=make([]byte, 2)
	binary.BigEndian.PutUint16(uint16_buf, cmd.BlockRuntimeID)
	_, err:=writer.Write(append(uint16_buf, uint8(len(cmd.ChestSlots))))
	// They are different parts, but wrote together for convenient
	if err!=nil {
		return err
	}
	for _, slot:=range cmd.ChestSlots {
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

func (cmd *PlaceRuntimeBlockWithChestData) Unmarshal(reader io.Reader) error {
	uint16_buf:=make([]byte, 2)
	_, err:=io.ReadAtLeast(reader, uint16_buf, 2)
	if err!=nil {
		return err
	}
	cmd.BlockRuntimeID=binary.BigEndian.Uint16(uint16_buf)
	uint8_buf:=make([]byte, 1)
	_, err=io.ReadAtLeast(reader, uint8_buf, 1)
	if err!=nil {
		return err
	}
	cmd.ChestSlots=make([]types.ChestSlot, int(uint8_buf[0]))
	for i:=0;i<int(uint8_buf[0]);i++ {
		item_name, err:=readString(reader)
		if err!=nil {
			return err
		}
		cmd.ChestSlots[i].Name=item_name
		count_damage_slot_buf:=make([]byte, 4)
		_, err=io.ReadAtLeast(reader, count_damage_slot_buf, 4)
		if err!=nil {
			return err
		}
		cmd.ChestSlots[i].Count=count_damage_slot_buf[0]
		cmd.ChestSlots[i].Damage=binary.BigEndian.Uint16(count_damage_slot_buf[1:3])
		cmd.ChestSlots[i].Slot=count_damage_slot_buf[3]
	}
	return nil
}