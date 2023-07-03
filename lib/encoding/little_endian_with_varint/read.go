package little_endian_with_varint

import "math"

type CanReadOutBytes interface {
	ReadOut(len int) (b []byte, err error)
}

type CanReadByte interface {
	ReadByte() (b byte, err error)
}

type CanReadBoth interface {
	CanReadByte
	CanReadOutBytes
}

// Int16 ...
func Int16(r CanReadOutBytes) (int16, error) {
	b, err := r.ReadOut(2)
	if err != nil {
		return 0, err
	}
	return int16(uint16(b[0]) | uint16(b[1])<<8), nil
}

// Int32 ...
func Int32(r CanReadBoth) (int32, error) {
	var ux uint32
	for i := uint(0); i < 35; i += 7 {
		b, err := r.ReadByte()
		if err != nil {
			return 0, err
		}
		ux |= uint32(b&0x7f) << i
		if b&0x80 == 0 {
			break
		}
	}
	x := int32(ux >> 1)
	if ux&1 != 0 {
		x = ^x
	}
	return x, nil
}

// Int64 ...
func Int64(r CanReadBoth) (int64, error) {
	var ux uint64
	for i := uint(0); i < 70; i += 7 {
		b, err := r.ReadByte()
		if err != nil {
			return 0, err
		}
		ux |= uint64(b&0x7f) << i
		if b&0x80 == 0 {
			break
		}
	}
	x := int64(ux >> 1)
	if ux&1 != 0 {
		x = ^x
	}
	return x, nil
}

// Float32 ...
func Float32(r CanReadOutBytes) (float32, error) {
	b, err := r.ReadOut(4)
	if err != nil {
		return 0, err
	}
	return math.Float32frombits(uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24), nil
}

// Float64 ...
func Float64(r CanReadOutBytes) (float64, error) {
	b, err := r.ReadOut(8)
	if err != nil {
		return 0, err
	}
	return math.Float64frombits(uint64(b[0]) | uint64(b[1])<<8 | uint64(b[2])<<16 | uint64(b[3])<<24 |
		uint64(b[4])<<32 | uint64(b[5])<<40 | uint64(b[6])<<48 | uint64(b[7])<<56), nil
}

// String ...
func String(r CanReadBoth) (string, error) {
	var length uint32
	for i := uint(0); i < 35; i += 7 {
		b, err := r.ReadByte()
		if err != nil {
			return "", err
		}
		length |= uint32(b&0x7f) << i
		if b&0x80 == 0 {
			break
		}
	}
	if length > math.MaxInt16 {
		return "", ErrStringLengthExceeds
	}
	data, err := r.ReadOut(int(length))
	if err != nil {
		return "", err
	}
	return string(data), nil
}
