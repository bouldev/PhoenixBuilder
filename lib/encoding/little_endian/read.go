package little_endian

import "math"

type CanReadOutBytes interface {
	ReadOut(len int) (b []byte, err error)
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
func Int32(r CanReadOutBytes) (int32, error) {
	b, err := r.ReadOut(4)
	if err != nil {
		return 0, err
	}
	return int32(uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24), nil
}

// Int64 ...
func Int64(r CanReadOutBytes) (int64, error) {
	b, err := r.ReadOut(8)
	if err != nil {
		return 0, err
	}
	return int64(uint64(b[0]) | uint64(b[1])<<8 | uint64(b[2])<<16 | uint64(b[3])<<24 |
		uint64(b[4])<<32 | uint64(b[5])<<40 | uint64(b[6])<<48 | uint64(b[7])<<56), nil
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
func String(r CanReadOutBytes) (string, error) {
	b, err := r.ReadOut(2)
	if err != nil {
		return "", err
	}
	stringLength := int(uint16(b[0]) | uint16(b[1])<<8)
	data, err := r.ReadOut(stringLength)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
