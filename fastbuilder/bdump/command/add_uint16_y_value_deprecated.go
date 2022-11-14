package command

import (
	"io"
	"encoding/binary"
)

type AddUint16YValueDeprecated struct {
	Value uint16
}

func (_ *AddUint16YValueDeprecated) ID() uint16 {
	return 4
}

func (_ *AddUint16YValueDeprecated) Name() string {
	return "AddUint16YValueDeprecatedCommand"
}

func (cmd *AddUint16YValueDeprecated) Marshal(writer io.Writer) error {
	buf:=make([]byte, 2)
	binary.BigEndian.PutUint16(buf, cmd.Value)
	_, err:=writer.Write(buf)
	return err
}

func (cmd *AddUint16YValueDeprecated) Unmarshal(reader io.Reader) error {
	buf:=make([]byte, 2)
	_, err:=io.ReadAtLeast(reader, buf, 2)
	if err!=nil {
		return err
	}
	cmd.Value=binary.BigEndian.Uint16(buf)
	return nil
}