package command

import (
	"io"
	"encoding/binary"
)

type PlaceBlockWithBlockStatesDeprecated struct {
	BlockConstantStringID uint16
	BlockStatesString string
}

func (_ *PlaceBlockWithBlockStatesDeprecated) ID() uint16 {
	return 13
}

func (_ *PlaceBlockWithBlockStatesDeprecated) Name() string {
	return "PlaceBlockWithBlockStatesDeprecatedCommand"
}

func (cmd *PlaceBlockWithBlockStatesDeprecated) Marshal(writer io.Writer) error {
	buf:=make([]byte, 2)
	binary.BigEndian.PutUint16(buf, cmd.BlockConstantStringID)
	_, err:=writer.Write(buf)
	if err!=nil {
		return err
	}
	_, err=writer.Write(append([]byte(cmd.BlockStatesString), 0))
	return err
}

func (cmd *PlaceBlockWithBlockStatesDeprecated) Unmarshal(reader io.Reader) error {
	buf:=make([]byte, 2)
	_, err:=io.ReadAtLeast(reader, buf, 2)
	if err!=nil {
		return err
	}
	cmd.BlockConstantStringID=binary.BigEndian.Uint16(buf)
	blockStates, err:=readString(reader)
	if err!=nil {
		return err
	}
	cmd.BlockStatesString=blockStates
	return nil
}
