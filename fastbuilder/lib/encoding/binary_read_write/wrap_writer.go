package binary_read_write

import "io"

type wrappedWriter struct {
	underlayWriter io.Writer
	writeByte      func(b byte) error
}

func (w *wrappedWriter) Write(b []byte) (err error) {
	var n int
	n, err = w.underlayWriter.Write(b)
	if err != nil {
		return err
	}
	if n != len(b) {
		return ErrFailToFullyWrite
	}
	return nil
}

// it's ok to do so, since compiler will optimize this
func (w *wrappedWriter) WriteByte(b byte) error {
	return w.writeByte(b)
}

func WrapBinaryWriter(underlayWriter io.Writer) BinaryWriter {
	w := &wrappedWriter{
		underlayWriter: underlayWriter,
		writeByte:      nil,
	}
	if canWriteByte, ok := underlayWriter.(interface {
		WriteByte(b byte) error
	}); ok {
		w.writeByte = canWriteByte.WriteByte
	} else {
		w.writeByte = func(b byte) error {
			return w.Write([]byte{b})
		}
	}
	return w
}
