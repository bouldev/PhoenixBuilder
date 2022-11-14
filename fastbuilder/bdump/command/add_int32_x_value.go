package command

import (
	"io"
	"encoding/binary"
)

type AddInt32XValue struct {
	Value int32
}

func (_ *AddInt32XValue) ID() uint16 {
	return 21
}

func (_ *AddInt32XValue) Name() string {
	return "AddInt32XValueCommand"
}

func (cmd *AddInt32XValue) Marshal(writer io.Writer) error {
	buf:=make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(cmd.Value))
	_, err:=writer.Write(buf)
	return err
}

func (cmd *AddInt32XValue) Unmarshal(reader io.Reader) error {
	buf:=make([]byte, 4)
	_, err:=io.ReadAtLeast(reader, buf, 4)
	if err!=nil {
		return err
	}
	cmd.Value=int32(binary.BigEndian.Uint32(buf))
	return nil
}