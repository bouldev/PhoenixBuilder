package command

import (
	"io"
)

type AddYValue struct {}

func (_ *AddYValue) ID() uint16 {
	return 16
}

func (_ *AddYValue) Name() string {
	return "AddYValueCommand"
}

func (_ *AddYValue) Marshal(_ io.Writer) error {
	return nil
}

func (_ *AddYValue) Unmarshal(_ io.Reader) error {
	return nil
}