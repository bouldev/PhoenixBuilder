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

// Some deprecated commands may not be placed in this directory
// as I think we do not have to make them work

func readString(reader io.Reader) (string, error) {
	buf:=make([]byte, 1)
	str:=""
	for {
		_, err:=io.ReadAtLeast(reader, buf, 1)
		if err!=nil {
			return "", err
		}
		if buf[0]==0 {
			return str, nil
		}
		str+=string(buf[0])
	}
	// This should not happen
	return str, nil
}