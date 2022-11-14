package command

import (
	"io"
)

type AddYValueDeprecated struct {}

func (_ *AddYValueDeprecated) ID() uint16 {
	return 5
}

func (_ *AddYValueDeprecated) Name() string {
	return "AddYValueDeprecatedCommand"
}

func (_ *AddYValueDeprecated) Marshal(_ io.Writer) error {
	return nil
}

func (_ *AddYValueDeprecated) Unmarshal(_ io.Reader) error {
	return nil
}