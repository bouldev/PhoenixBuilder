package binary_read_write

type BinaryWriter interface {
	Write(b []byte) (err error)
	WriteByte(byte) error
}

type BinaryReader interface {
	Read(b []byte) (err error)
	ReadOut(len int) (b []byte, err error)
	ReadByte() (b byte, err error)
}
