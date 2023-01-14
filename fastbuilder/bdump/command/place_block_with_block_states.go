package command

import (
	"io"
	"encoding/binary"
)

type PlaceBlockWithBlockStates struct {
	BlockConstantStringID uint16
	BlockStatesConstantStringID uint16
}

func (_ *PlaceBlockWithBlockStates) ID() uint16 {
	return 5
}

func (_ *PlaceBlockWithBlockStates) Name() string {
	return "PlaceBlockWithBlockStatesCommand"
}

func (cmd *PlaceBlockWithBlockStates) Marshal(writer io.Writer) error {
	buf:=make([]byte, 2)
	binary.BigEndian.PutUint16(buf, cmd.BlockConstantStringID)
	_, err:=writer.Write(buf)
	if err!=nil {
		return err
	}
	binary.BigEndian.PutUint16(buf, cmd.BlockStatesConstantStringID)
	_, err=writer.Write(buf)
	return err
}

func (cmd *PlaceBlockWithBlockStates) Unmarshal(reader io.Reader) error {
	buf:=make([]byte, 2)
	_, err:=io.ReadAtLeast(reader, buf, 2)
	if err!=nil {
		return err
	}
	cmd.BlockConstantStringID=binary.BigEndian.Uint16(buf)
	_, err=io.ReadAtLeast(reader, buf, 2)
	if err!=nil {
		return err
	}
	cmd.BlockStatesConstantStringID=binary.BigEndian.Uint16(buf)
	return nil
}
