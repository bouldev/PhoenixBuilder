package command

import (
	"io"
)

type AddXValueDeprecated struct {}

func (_ *AddXValueDeprecated) ID() uint16 {
	return 3
}

func (_ *AddXValueDeprecated) Name() string {
	return "AddXValueDeprecatedCommand"
}

func (_ *AddXValueDeprecated) Marshal(_ io.Writer) error {
	return nil
}

func (_ *AddXValueDeprecated) Unmarshal(_ io.Reader) error {
	return nil
}