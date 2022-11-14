package command

import (
	"io"
)

type AddZValue struct {}

func (_ *AddZValue) ID() uint16 {
	return 18
}

func (_ *AddZValue) Name() string {
	return "AddZValueCommand"
}

func (_ *AddZValue) Marshal(_ io.Writer) error {
	return nil
}

func (_ *AddZValue) Unmarshal(_ io.Reader) error {
	return nil
}