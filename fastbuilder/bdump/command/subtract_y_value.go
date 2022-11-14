package command

import (
	"io"
)

type SubtractYValue struct {}

func (_ *SubtractYValue) ID() uint16 {
	return 17
}

func (_ *SubtractYValue) Name() string {
	return "SubtractYValueCommand"
}

func (_ *SubtractYValue) Marshal(_ io.Writer) error {
	return nil
}

func (_ *SubtractYValue) Unmarshal(_ io.Reader) error {
	return nil
}