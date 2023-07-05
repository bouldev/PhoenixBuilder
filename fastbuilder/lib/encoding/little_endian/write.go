package little_endian

import (
	"math"
	"unsafe"
)

type CanWriteBytes interface {
	Write([]byte) error
}

// WriteInt16 ...
func WriteInt16(w CanWriteBytes, x int16) error {
	return w.Write([]byte{byte(x), byte(x >> 8)})
}

// WriteInt32 ...
func WriteInt32(w CanWriteBytes, x int32) error {
	return w.Write([]byte{byte(x), byte(x >> 8), byte(x >> 16), byte(x >> 24)})
}

// WriteInt64 ...
func WriteInt64(w CanWriteBytes, x int64) error {
	return w.Write([]byte{byte(x), byte(x >> 8), byte(x >> 16), byte(x >> 24),
		byte(x >> 32), byte(x >> 40), byte(x >> 48), byte(x >> 56)})
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
func WriteString(w CanWriteBytes, x string) error {
	if len(x) > math.MaxInt16 {
		return ErrStringLengthExceeds
	}
	length := int16(len(x))
	if err := w.Write([]byte{byte(length), byte(length >> 8)}); err != nil {
		return err
	}
	// Use unsafe conversion from a string to a byte slice to prevent copying.
	if err := w.Write(*(*[]byte)(unsafe.Pointer(&x))); err != nil {
		return err
	}
	return nil
}
