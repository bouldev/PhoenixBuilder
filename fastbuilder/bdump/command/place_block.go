package command

import (
	"io"
	"encoding/binary"
)

type PlaceBlock struct {
	BlockConstantStringID uint16
	BlockData uint16
}

func (_ *PlaceBlock) ID() uint16 {
	return 7
}

func (_ *PlaceBlock) Name() string {
	return "PlaceBlockCommand"
}

func (cmd *PlaceBlock) Marshal(writer io.Writer) error {
	buf:=make([]byte, 2)
	binary.BigEndian.PutUint16(buf, cmd.BlockConstantStringID)
	_, err:=writer.Write(buf)
	if err!=nil {
		return err
	}
	binary.BigEndian.PutUint16(buf, cmd.BlockData)
	_, err=writer.Write(buf)
	return err
}

func (cmd *PlaceBlock) Unmarshal(reader io.Reader) error {
	buf:=make([]byte, 4)
	_, err:=io.ReadAtLeast(reader, buf, 4)
	if err!=nil {
		return err
	}
	cmd.BlockConstantStringID=binary.BigEndian.Uint16(buf[0:2])
	cmd.BlockData=binary.BigEndian.Uint16(buf[2:])
	return nil
}