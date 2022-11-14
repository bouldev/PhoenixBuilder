package command

import (
	"io"
)

type AddZValue0 struct {}

func (_ *AddZValue0) ID() uint16 {
	return 8
}

func (_ *AddZValue0) Name() string {
	return "AddZValue0Command"
}

func (_ *AddZValue0) Marshal(_ io.Writer) error {
	return nil
}

func (_ *AddZValue0) Unmarshal(_ io.Reader) error {
	return nil
}