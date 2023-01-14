package bdump

import (
	"io"
	"hash"
	"phoenixbuilder/fastbuilder/bdump/command"
)

type BDumpWriter struct {
	writer io.Writer
}

func (w *BDumpWriter) WriteCommand(cmd command.Command) error {
	return command.WriteCommand(cmd, w.writer)
}

type HashedWriter struct {
	writer io.Writer
	hash hash.Hash
}

func (w *HashedWriter) Write(p []byte) (n int, err error) {
	w.hash.Write(p)
	n, err=w.writer.Write(p)
	return
}