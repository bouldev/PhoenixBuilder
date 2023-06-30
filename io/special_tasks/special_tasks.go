//go:build !is_tweak
// +build !is_tweak

package special_tasks

import (
	"fmt"
	"io"
	"math"
)

type SolidSimplePos struct {
	X int64 `json:"x"`
	Y int64 `json:"y"`
	Z int64 `json:"z"`
}

type SolidRet struct {
	BlockName  string         `json:"blockName"`
	Position   SolidSimplePos `json:"position"`
	StatusCode int64          `json:"statusCode"`
}

type byteAndNormalReader interface {
	io.Reader
	io.ByteReader
}

func readVarint32(reader byteAndNormalReader) (int32, error) {
	// Copied code, from gophertunnel
	var val uint32
	for i := uint(0); i < 35; i += 7 {
		b, err := reader.ReadByte()
		if err != nil {
			return 0, fmt.Errorf("Early EOF")
		}
		val |= uint32(b&0x7f) << i
		if b&0x80 == 0 {
			break
		}
	}
	ret_val := int32(val >> 1)
	if val&1 != 0 {
		ret_val = ^ret_val
	}
	return ret_val, nil
}

func readVarint64(reader byteAndNormalReader) (int64, error) {
	// Copied code, from gophertunnel
	var val uint64
	for i := uint(0); i < 70; i += 7 {
		b, err := reader.ReadByte()
		if err != nil {
			return 0, fmt.Errorf("Early EOF")
		}
		val |= uint64(b&0x7f) << i
		if b&0x80 == 0 {
			break
		}
	}
	rval := int64(val >> 1)
	if val&1 != 0 {
		rval = ^rval
	}
	return rval, nil
}

func readNBTString(reader byteAndNormalReader) (string, error) {
	// Code mainly from gophertunnel
	var length uint32
	for i := uint(0); i < 35; i += 7 {
		b, err := reader.ReadByte()
		if err != nil {
			return "", fmt.Errorf("Early EOF")
		}
		length |= uint32(b&0x7f) << i
		if b&0x80 == 0 {
			break
		}
	}
	if length > math.MaxInt16 {
		return "", fmt.Errorf("Invalid string length")
	}
	buf := make([]byte, length)
	_, err := io.ReadAtLeast(reader, buf, int(length))
	if err != nil {
		return "", fmt.Errorf("Early EOF")
	}
	return string(buf), nil
}
