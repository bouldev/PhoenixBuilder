package command

import (
	"io"
	"encoding/binary"
)

type AssignNBTData struct {
	Data []byte
}

func (_ *AssignNBTData) ID() uint16 {
	return 39
}

func (_ *AssignNBTData) Name() string {
	return "AssignNBTDataCommand"
}

func (cmd *AssignNBTData) Marshal(writer io.Writer) error {
	lenBuf:=make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, uint32(len(cmd.Data)))
	_, err:=writer.Write(append(lenBuf, cmd.Data...))
	return err
}

func (cmd *AssignNBTData) Unmarshal(reader io.Reader) error {
	lenBuf:=make([]byte, 4)
	_, err:=io.ReadAtLeast(reader, lenBuf, 4)
	if err!=nil {
		return err
	}
	cmd.Data=make([]byte, int(binary.BigEndian.Uint32(lenBuf)))
	_, err=io.ReadAtLeast(reader, cmd.Data, len(cmd.Data))
	return err
}