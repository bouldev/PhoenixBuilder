package command

import "io"

type NoOperation struct {}

func (_ *NoOperation) ID() uint16 {
	return 9
}

func (_ *NoOperation) Name() string {
	return "NoOperationCommand"
}

func (_ *NoOperation) Marshal(_ io.Writer) error {
	return nil
}

func (_ *NoOperation) Unmarshal(_ io.Reader) error {
	return nil
}