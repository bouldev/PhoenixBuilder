package encoding

import (
	"encoding/binary"
	"fmt"
	"phoenixbuilder/minecraft/protocol"

	"github.com/google/uuid"
)

// 从阅读器阅读 length 个字节
func (r *Reader) ReadBytes(length int) ([]byte, error) {
	ans := make([]byte, length)
	_, err := r.r.Read(ans)
	if err != nil {
		return nil, fmt.Errorf("ReadBytes: %v", err)
	}
	return ans, nil
}

// 从阅读器阅读一个二进制切片并返回到 x 上
func (r *Reader) Slice(x *[]byte) error {
	var length uint32
	err := binary.Read(r.r, binary.BigEndian, &length)
	if err != nil {
		return fmt.Errorf("(r *Reader) Slice: %v", err)
	}
	// get the length of the target string
	slice, err := r.ReadBytes(int(length))
	if err != nil {
		return fmt.Errorf("(r *Reader) Slice: %v", err)
	}
	*x = slice
	// get the target slice
	return nil
	// return
}

// 从阅读器阅读一个字符串并返回到 x 上
func (r *Reader) String(x *string) error {
	var length uint16
	err := binary.Read(r.r, binary.BigEndian, &length)
	if err != nil {
		return fmt.Errorf("(r *Reader) String: %v", err)
	}
	// get the length of the target string
	stringBytes, err := r.ReadBytes(int(length))
	if err != nil {
		return fmt.Errorf("(r *Reader) String: %v", err)
	}
	*x = string(stringBytes)
	// get the target string
	return nil
	// return
}

// 从阅读器阅读一个字符串切片并返回到 x 上
func (r *Reader) StringSlice(x *[]string) error {
	var length uint32
	err := binary.Read(r.r, binary.BigEndian, &length)
	if err != nil {
		return fmt.Errorf("(r *Reader) StringSlice: %v", err)
	}
	// get the length of the target slice
	*x = make([]string, length)
	// make
	for i := 0; i < int(length); i++ {
		err = r.String(&(*x)[i])
		if err != nil {
			return fmt.Errorf("(r *Reader) StringSlice: %v", err)
		}
	}
	// get the target slice
	return nil
	// return
}

// 从阅读器阅读一个 map[string][]byte 并返回到 x 上
func (r *Reader) Map(x *map[string][]byte) error {
	var length uint16
	err := binary.Read(r.r, binary.BigEndian, &length)
	if err != nil {
		return fmt.Errorf("(r *Reader) Map: %v", err)
	}
	// get the length of the target map
	for i := 0; i < int(length); i++ {
		key := ""
		value := []byte{}
		err = r.String(&key)
		if err != nil {
			return fmt.Errorf("(r *Reader) Map: %v", err)
		}
		err = r.Slice(&value)
		if err != nil {
			return fmt.Errorf("(r *Reader) Map: %v", err)
		}
		(*x)[key] = value
	}
	// read map and unmarshal it into x
	return nil
	// return
}

// 从阅读器阅读一个布尔值并返回到 x 上
func (r *Reader) Bool(x *bool) error {
	ans, err := r.ReadBytes(1)
	if err != nil {
		return fmt.Errorf("(r *Reader) Bool: %v", err)
	}
	switch ans[0] {
	case 0:
		*x = false
	case 1:
		*x = true
	case 2:
		return fmt.Errorf("(r *Reader) Bool: Unexpected value %#v was find", ans)
	}
	return nil
}

// 从阅读器阅读一个 uint8 并返回到 x 上
func (r *Reader) Uint8(x *uint8) error {
	ans, err := r.ReadBytes(1)
	if err != nil {
		return fmt.Errorf("(r *Reader) Uint8: %v", err)
	}
	*x = ans[0]
	return nil
}

// 从阅读器阅读一个 int8 并返回到 x 上
func (r *Reader) Int8(x *int8) error {
	err := binary.Read(r.r, binary.BigEndian, x)
	if err != nil {
		return fmt.Errorf("(r *Reader) Int8: %v", err)
	}
	return nil
}

// 从阅读器阅读一个 uint16 并返回到 x 上
func (r *Reader) Uint16(x *uint16) error {
	err := binary.Read(r.r, binary.BigEndian, x)
	if err != nil {
		return fmt.Errorf("(r *Reader) Uint16: %v", err)
	}
	return nil
}

// 从阅读器阅读一个 int16 并返回到 x 上
func (r *Reader) Int16(x *int16) error {
	err := binary.Read(r.r, binary.BigEndian, x)
	if err != nil {
		return fmt.Errorf("(r *Reader) Int16: %v", err)
	}
	return nil
}

// 从阅读器阅读一个 uint32 并返回到 x 上
func (r *Reader) Uint32(x *uint32) error {
	err := binary.Read(r.r, binary.BigEndian, x)
	if err != nil {
		return fmt.Errorf("(r *Reader) Uint32: %v", err)
	}
	return nil
}

// 从阅读器阅读一个 int32 并返回到 x 上
func (r *Reader) Int32(x *int32) error {
	err := binary.Read(r.r, binary.BigEndian, x)
	if err != nil {
		return fmt.Errorf("(r *Reader) Int32: %v", err)
	}
	return nil
}

// 从阅读器阅读一个 uint64 并返回到 x 上
func (r *Reader) Uint64(x *uint64) error {
	err := binary.Read(r.r, binary.BigEndian, x)
	if err != nil {
		return fmt.Errorf("(r *Reader) Uint64: %v", err)
	}
	return nil
}

// 从阅读器阅读一个 int64 并返回到 x 上
func (r *Reader) Int64(x *int64) error {
	err := binary.Read(r.r, binary.BigEndian, x)
	if err != nil {
		return fmt.Errorf("(r *Reader) Int64: %v", err)
	}
	return nil
}

// 从阅读器阅读一个 uuid.UUID 并返回到 x 上
func (r *Reader) UUID(x *uuid.UUID) error {
	p, err := r.ReadBytes(UUIDConstantLength)
	if err != nil {
		return fmt.Errorf("(r *Reader) UUID: %v", err)
	}
	// get binary of uuid
	err = x.UnmarshalBinary(p)
	if err != nil {
		return fmt.Errorf("(r *Reader) UUID: %v", err)
	}
	// unmarshal
	return nil
	// return
}

// 从阅读器阅读一个 protocol.CommandOutputMessage 并返回到 x 上
func (r *Reader) CommandOutputMessage(x *protocol.CommandOutputMessage) error {
	err := r.Bool(&x.Success)
	if err != nil {
		return fmt.Errorf("(r *Reader) CommandOutputMessage: %v", err)
	}
	// Success
	err = r.String(&x.Message)
	if err != nil {
		return fmt.Errorf("(r *Reader) CommandOutputMessage: %v", err)
	}
	// Message
	if x.Parameters == nil {
		x.Parameters = make([]string, 0)
	}
	err = r.StringSlice(&x.Parameters)
	if err != nil {
		return fmt.Errorf("(r *Reader) CommandOutputMessage: %v", err)
	}
	// Parameters
	return nil
	// return
}

// 从阅读器阅读一个 []protocol.CommandOutputMessage 并返回到 x 上
func (r *Reader) CommandOutputMessageSlice(x *[]protocol.CommandOutputMessage) error {
	var length uint32
	err := binary.Read(r.r, binary.BigEndian, &length)
	if err != nil {
		return fmt.Errorf("(r *Reader) CommandOutputMessageSlice: %v", err)
	}
	// get length
	*x = make([]protocol.CommandOutputMessage, length)
	// make
	for i := 0; i < int(length); i++ {
		err = r.CommandOutputMessage(&(*x)[i])
		if err != nil {
			return fmt.Errorf("(r *Reader) CommandOutputMessageSlice: %v", err)
		}
	}
	// read data
	return nil
	// return
}
