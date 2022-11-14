package command

import (
	"io"
	"encoding/binary"
)

type AddInt32YValue struct {
	Value int32
}

func (_ *AddInt32YValue) ID() uint16 {
	return 23
}

func (_ *AddInt32YValue) Name() string {
	return "AddInt32YValueCommand"
}

func (cmd *AddInt32YValue) Marshal(writer io.Writer) error {
	buf:=make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(cmd.Value))
	_, err:=writer.Write(buf)
	return err
}

func (cmd *AddInt32YValue) Unmarshal(reader io.Reader) error {
	buf:=make([]byte, 4)
	_, err:=io.ReadAtLeast(reader, buf, 4)
	if err!=nil {
		return err
	}
	cmd.Value=int32(binary.BigEndian.Uint32(buf))
	return nil
}