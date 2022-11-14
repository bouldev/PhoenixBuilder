package command

import (
	"io"
	"encoding/binary"
)

type PlaceRuntimeBlock struct {
	BlockRuntimeID uint16
}

func (_ *PlaceRuntimeBlock) ID() uint16 {
	return 32
}

func (_ *PlaceRuntimeBlock) Name() string {
	return "PlaceRuntimeBlockCommand"
}

func (cmd *PlaceRuntimeBlock) Marshal(writer io.Writer) error {
	buf:=make([]byte, 2)
	binary.BigEndian.PutUint16(buf, cmd.BlockRuntimeID)
	_, err:=writer.Write(buf)
	return err
}

func (cmd *PlaceRuntimeBlock) Unmarshal(reader io.Reader) error {
	buf:=make([]byte, 2)
	_, err:=io.ReadAtLeast(reader, buf, 2)
	if err!=nil {
		return err
	}
	cmd.BlockRuntimeID=binary.BigEndian.Uint16(buf)
	return nil
}