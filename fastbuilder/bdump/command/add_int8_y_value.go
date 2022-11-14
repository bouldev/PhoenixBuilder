package command

import (
	"io"
)

type AddInt8YValue struct {
	Value int8
}

func (_ *AddInt8YValue) ID() uint16 {
	return 29
}

func (_ *AddInt8YValue) Name() string {
	return "AddInt8YValueCommand"
}

func (cmd *AddInt8YValue) Marshal(writer io.Writer) error {
	buf:=[]byte{uint8(cmd.Value)}
	_, err:=writer.Write(buf)
	return err
}

func (cmd *AddInt8YValue) Unmarshal(reader io.Reader) error {
	buf:=make([]byte, 1)
	_, err:=io.ReadAtLeast(reader, buf, 1)
	if err!=nil {
		return err
	}
	cmd.Value=int8(buf[0])
	return nil
}