package command

import (
	"io"
)

type UseRuntimeIDPool struct {
	ID uint8
}

func (_ *UseRuntimeIDPool) ID() uint16 {
	return 31
}

func (_ *UseRuntimeIDPool) Name() string {
	return "UseRuntimeIDPoolCommand"
}

func (cmd *UseRuntimeIDPool) Marshal(writer io.Writer) error {
	_, err:=writer.Write([]byte{cmd.ID})
	return err
}

func (cmd *UseRuntimeIDPool) Unmarshal(reader io.Reader) error {
	buf:=make([]byte, 1)
	_, err:=io.ReadAtLeast(reader, buf, 1)
	if err!=nil {
		return err
	}
	cmd.ID=buf[0]
	return nil
}