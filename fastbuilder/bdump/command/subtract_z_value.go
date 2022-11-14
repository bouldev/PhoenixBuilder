package command

import (
	"io"
)

type SubtractZValue struct {}

func (_ *SubtractZValue) ID() uint16 {
	return 19
}

func (_ *SubtractZValue) Name() string {
	return "SubtractZValueCommand"
}

func (_ *SubtractZValue) Marshal(_ io.Writer) error {
	return nil
}

func (_ *SubtractZValue) Unmarshal(_ io.Reader) error {
	return nil
}