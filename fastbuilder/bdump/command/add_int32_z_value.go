package command

import (
	"io"
	"encoding/binary"
)

type AddInt32ZValue struct {
	Value int32
}

func (_ *AddInt32ZValue) ID() uint16 {
	return 25
}

func (_ *AddInt32ZValue) Name() string {
	return "AddInt32ZValueCommand"
}

func (cmd *AddInt32ZValue) Marshal(writer io.Writer) error {
	buf:=make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(cmd.Value))
	_, err:=writer.Write(buf)
	return err
}

func (cmd *AddInt32ZValue) Unmarshal(reader io.Reader) error {
	buf:=make([]byte, 4)
	_, err:=io.ReadAtLeast(reader, buf, 4)
	if err!=nil {
		return err
	}
	cmd.Value=int32(binary.BigEndian.Uint32(buf))
	return nil
}