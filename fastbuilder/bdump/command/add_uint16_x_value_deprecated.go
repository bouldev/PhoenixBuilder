package command

import (
	"io"
	"encoding/binary"
)

type AddUint16XValueDeprecated struct {
	Value uint16
}

func (_ *AddUint16XValueDeprecated) ID() uint16 {
	return 2
}

func (_ *AddUint16XValueDeprecated) Name() string {
	return "AddUint16XValueDeprecatedCommand"
}

func (cmd *AddUint16XValueDeprecated) Marshal(writer io.Writer) error {
	buf:=make([]byte, 2)
	binary.BigEndian.PutUint16(buf, cmd.Value)
	_, err:=writer.Write(buf)
	return err
}

func (cmd *AddUint16XValueDeprecated) Unmarshal(reader io.Reader) error {
	buf:=make([]byte, 2)
	_, err:=io.ReadAtLeast(reader, buf, 2)
	if err!=nil {
		return err
	}
	cmd.Value=binary.BigEndian.Uint16(buf)
	return nil
}