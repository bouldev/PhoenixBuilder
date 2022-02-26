package path

import "io"

type FileWriter interface {
	io.Writer
	Close() error
}
