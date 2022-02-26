package path

import "io"

type FileReader interface {
	io.Reader
	Close() error
}
