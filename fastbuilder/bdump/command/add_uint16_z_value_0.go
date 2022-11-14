package command

import (
	"io"
	"encoding/binary"
)

type AddUint16ZValue0 struct {
	Value uint16
}

func (_ *AddUint16ZValue0) ID() uint16 {
	return 6
}

func (_ *AddUint16ZValue0) Name() string {
	return "AddUint16ZValue0Command"
}

func (cmd *AddUint16ZValue0) Marshal(writer io.Writer) error {
	buf:=make([]byte, 2)
	binary.BigEndian.PutUint16(buf, cmd.Value)
	_, err:=writer.Write(buf)
	return err
}

func (cmd *AddUint16ZValue0) Unmarshal(reader io.Reader) error {
	buf:=make([]byte, 2)
	_, err:=io.ReadAtLeast(reader, buf, 2)
	if err!=nil {
		return err
	}
	cmd.Value=binary.BigEndian.Uint16(buf)
	return nil
}