package command

import "io"

type Terminate struct {}

func (_ *Terminate) ID() uint16 {
	return 88
}

func (_ *Terminate) Name() string {
	return "TerminateCommand"
}

func (_ *Terminate) Marshal(_ io.Writer) error {
	return nil
}

func (_ *Terminate) Unmarshal(_ io.Reader) error {
	return nil
}