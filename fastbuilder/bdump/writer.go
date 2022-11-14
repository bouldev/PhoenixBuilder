package bdump

import (
	"io"
	"phoenixbuilder/fastbuilder/bdump/command"
)

type BDumpWriter struct {
	writer io.Writer
}

func (w *BDumpWriter) WriteCommand(cmd command.Command) error {
	return command.WriteCommand(cmd, w.writer)
}