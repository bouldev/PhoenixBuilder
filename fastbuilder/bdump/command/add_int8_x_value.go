package command

import (
	"io"
)

type AddInt8XValue struct {
	Value int8
}

func (_ *AddInt8XValue) ID() uint16 {
	return 28
}

func (_ *AddInt8XValue) Name() string {
	return "AddInt8XValueCommand"
}

func (cmd *AddInt8XValue) Marshal(writer io.Writer) error {
	buf:=[]byte{uint8(cmd.Value)}
	_, err:=writer.Write(buf)
	return err
}

func (cmd *AddInt8XValue) Unmarshal(reader io.Reader) error {
	buf:=make([]byte, 1)
	_, err:=io.ReadAtLeast(reader, buf, 1)
	if err!=nil {
		return err
	}
	cmd.Value=int8(buf[0])
	return nil
}