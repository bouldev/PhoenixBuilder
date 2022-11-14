package command

import (
	"io"
	"encoding/binary"
)

type AddInt16XValue struct {
	Value int16
}

func (_ *AddInt16XValue) ID() uint16 {
	return 20
}

func (_ *AddInt16XValue) Name() string {
	return "AddInt16XValueCommand"
}

func (cmd *AddInt16XValue) Marshal(writer io.Writer) error {
	buf:=make([]byte, 2)
	binary.BigEndian.PutUint16(buf, uint16(cmd.Value))
	_, err:=writer.Write(buf)
	return err
}

func (cmd *AddInt16XValue) Unmarshal(reader io.Reader) error {
	buf:=make([]byte, 2)
	_, err:=io.ReadAtLeast(reader, buf, 2)
	if err!=nil {
		return err
	}
	cmd.Value=int16(binary.BigEndian.Uint16(buf))
	return nil
}