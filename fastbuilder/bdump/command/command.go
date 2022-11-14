package command

import (
	"io"
)

type Command interface {
	ID() uint16 // Extra ID spaces (uint16) may be allocated in the future.
	Name() string
	Marshal(writer io.Writer) error
	Unmarshal(reader io.Reader) error
}
