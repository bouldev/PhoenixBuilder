package command

import (
	"encoding/binary"
	"io"
	"phoenixbuilder/minecraft/nbt"
)

type PlaceBlockWithNBTData struct {
	BlockConstantStringID       uint16
	BlockStatesConstantStringID uint16
	BlockNBT_bytes              []byte
	BlockNBT                    map[string]interface{}
}

func (_ *PlaceBlockWithNBTData) ID() uint16 {
	return 41
}

func (_ *PlaceBlockWithNBTData) Name() string {
	return "PlaceBlockWithNBTDataCommand"
}

func (cmd *PlaceBlockWithNBTData) Marshal(writer io.Writer) error {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, cmd.BlockConstantStringID)
	_, err := writer.Write(buf)
	if err != nil {
		return err
	}
	binary.BigEndian.PutUint16(buf, cmd.BlockStatesConstantStringID)
	_, err = writer.Write(buf)
	if err != nil {
		return err
	}
	_, err = writer.Write(append(buf, cmd.BlockNBT_bytes...)) // cmd.BlockNBT_bytes 以 nbt.LittleEndian 编码
	return err
}

func (cmd *PlaceBlockWithNBTData) Unmarshal(reader io.Reader) error {
	buf := make([]byte, 2)
	_, err := io.ReadAtLeast(reader, buf, 2)
	if err != nil {
		return err
	}
	cmd.BlockConstantStringID = binary.BigEndian.Uint16(buf)
	buf = make([]byte, 2)
	_, err = io.ReadAtLeast(reader, buf, 2)
	if err != nil {
		return err
	}
	cmd.BlockStatesConstantStringID = binary.BigEndian.Uint16(buf)
	_, err = io.ReadAtLeast(reader, buf, 2)
	if err != nil {
		return err
	}
	err = nbt.NewDecoderWithEncoding(reader, nbt.LittleEndian).Decode(&cmd.BlockNBT)
	return err
}
