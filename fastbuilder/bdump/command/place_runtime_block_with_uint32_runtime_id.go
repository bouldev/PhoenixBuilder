package command

import (
	"io"
	"encoding/binary"
)

type PlaceRuntimeBlockWithUint32RuntimeID struct {
	BlockRuntimeID uint32
}

func (_ *PlaceRuntimeBlockWithUint32RuntimeID) ID() uint16 {
	return 33
}

func (_ *PlaceRuntimeBlockWithUint32RuntimeID) Name() string {
	return "PlaceRuntimeBlockUint32RuntimeIDCommand"
}

func (cmd *PlaceRuntimeBlockWithUint32RuntimeID) Marshal(writer io.Writer) error {
	buf:=make([]byte, 4)
	binary.BigEndian.PutUint32(buf, cmd.BlockRuntimeID)
	_, err:=writer.Write(buf)
	return err
}

func (cmd *PlaceRuntimeBlockWithUint32RuntimeID) Unmarshal(reader io.Reader) error {
	buf:=make([]byte, 4)
	_, err:=io.ReadAtLeast(reader, buf, 4)
	if err!=nil {
		return err
	}
	cmd.BlockRuntimeID=binary.BigEndian.Uint32(buf)
	return nil
}