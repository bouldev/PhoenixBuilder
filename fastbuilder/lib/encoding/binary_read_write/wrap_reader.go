package binary_read_write

import "io"

type wrappedReader struct {
	underlayReader io.Reader
	readByte       func() (b byte, err error)
}

func (w *wrappedReader) Read(b []byte) (err error) {
	var n int
	n, err = io.ReadAtLeast(w.underlayReader, b, len(b))
	if err != nil {
		return err
	}
	if n != len(b) {
		return ErrFailToFullyRead
	}
	return nil
}

func (w *wrappedReader) ReadOut(len int) (b []byte, err error) {
	holder := make([]byte, len)
	err = w.Read(holder)
	if err != nil {
		return nil, err
	}
	return holder, err
}

// it's ok to do so, since compiler will optimize this
func (w *wrappedReader) ReadByte() (b byte, err error) {
	return w.readByte()
}

func WrapBinaryReader(underlayReader io.Reader) BinaryReader {
	r := &wrappedReader{
		underlayReader: underlayReader,
		readByte:       nil,
	}
	if canReadByte, ok := underlayReader.(interface {
		ReadByte() (b byte, err error)
	}); ok {
		r.readByte = canReadByte.ReadByte
	} else {
		r.readByte = func() (b byte, err error) {
			data := make([]byte, 1)
			err = r.Read(data)
			return data[0], err
		}
	}
	return r
}
