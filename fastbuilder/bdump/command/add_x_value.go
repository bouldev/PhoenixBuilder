package command

import (
	"io"
)

type AddXValue struct {}

func (_ *AddXValue) ID() uint16 {
	return 14
}

func (_ *AddXValue) Name() string {
	return "AddXValueCommand"
}

func (_ *AddXValue) Marshal(_ io.Writer) error {
	return nil
}

func (_ *AddXValue) Unmarshal(_ io.Reader) error {
	return nil
}