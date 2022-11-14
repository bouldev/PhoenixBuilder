package command

import (
	"io"
)

type AddInt8ZValue struct {
	Value int8
}

func (_ *AddInt8ZValue) ID() uint16 {
	return 30
}

func (_ *AddInt8ZValue) Name() string {
	return "AddInt8ZValueCommand"
}

func (cmd *AddInt8ZValue) Marshal(writer io.Writer) error {
	buf:=[]byte{uint8(cmd.Value)}
	_, err:=writer.Write(buf)
	return err
}

func (cmd *AddInt8ZValue) Unmarshal(reader io.Reader) error {
	buf:=make([]byte, 1)
	_, err:=io.ReadAtLeast(reader, buf, 1)
	if err!=nil {
		return err
	}
	cmd.Value=int8(buf[0])
	return nil
}