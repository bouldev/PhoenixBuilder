package command

import (
	"io"
)

type CreateConstantString struct {
	ConstantString string
}

func (_ *CreateConstantString) ID() uint16 {
	return 1
}

func (_ *CreateConstantString) Name() string {
	return "CreateConstantStringCommand"
}

func (cmd *CreateConstantString) Marshal(writer io.Writer) error {
	str_c:=append([]byte(cmd.ConstantString), 0)
	_, err:=writer.Write(str_c)
	return err
}

func (cmd *CreateConstantString) Unmarshal(reader io.Reader) error {
	singlebuf:=make([]byte, 1)
	cmd.ConstantString=""
	for {
		_, err:=io.ReadAtLeast(reader, singlebuf, 1)
		if err != nil {
			return err
		}
		if singlebuf[0]==0 {
			break
		}
		cmd.ConstantString+=string(singlebuf[0])
	}
	return nil
}