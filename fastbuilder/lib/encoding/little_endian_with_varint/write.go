package little_endian_with_varint

import (
	"math"
	"unsafe"
)

type CanWriteBytes interface {
	Write([]byte) error
}
type CanWriteByte interface {
	WriteByte(byte) error
}

type CanWriteBoth interface {
	CanWriteByte
	CanWriteBytes
}

// WriteInt16 ...
func WriteInt16(w CanWriteBytes, x int16) error {
	return w.Write([]byte{byte(x), byte(x >> 8)})
}

// WriteInt32 ...
func WriteInt32(w CanWriteBoth, x int32) error {
	ux := uint32(x) << 1
	if x < 0 {
		ux = ^ux
	}
	for ux >= 0x80 {
		if err := w.WriteByte(byte(ux) | 0x80); err != nil {
			return err
		}
		ux >>= 7
	}
	if err := w.WriteByte(byte(ux)); err != nil {
		return err
	}
	return nil
}

// WriteInt64 ...
func WriteInt64(w CanWriteBoth, x int64) error {
	ux := uint64(x) << 1
	if x < 0 {
		ux = ^ux
	}
	for ux >= 0x80 {
		if err := w.WriteByte(byte(ux) | 0x80); err != nil {
			return err
		}
		ux >>= 7
	}
	if err := w.WriteByte(byte(ux)); err != nil {
		return err
	}
	return nil
}

// WriteFloat32 ...
func WriteFloat32(w CanWriteBytes, x float32) error {
	bits := math.Float32bits(x)
	return w.Write([]byte{byte(bits), byte(bits >> 8), byte(bits >> 16), byte(bits >> 24)})
}

// WriteFloat64 ...
func WriteFloat64(w CanWriteBytes, x float64) error {
	bits := math.Float64bits(x)
	return w.Write([]byte{byte(bits), byte(bits >> 8), byte(bits >> 16), byte(bits >> 24),
		byte(bits >> 32), byte(bits >> 40), byte(bits >> 48), byte(bits >> 56)})
}

// WriteString ...
func WriteString(w CanWriteBoth, x string) error {
	if len(x) > math.MaxInt16 {
		return ErrStringLengthExceeds
	}
	ux := uint32(len(x))
	for ux >= 0x80 {
		if err := w.WriteByte(byte(ux) | 0x80); err != nil {
			return err
		}
		ux >>= 7
	}
	if err := w.WriteByte(byte(ux)); err != nil {
		return err
	}
	// Use unsafe conversion from a string to a byte slice to prevent copying.
	if err := w.Write(*(*[]byte)(unsafe.Pointer(&x))); err != nil {
		return err
	}
	return nil
}
