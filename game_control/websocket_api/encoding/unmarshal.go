package encoding

import (
	"encoding/binary"
	"fmt"
	"phoenixbuilder/minecraft/protocol"

	"github.com/google/uuid"
)

// 从阅读器阅读一个 uint8 并返回到 x 上
func (r *Reader) Uint8(x *uint8) {
	if err := binary.Read(r.r, binary.BigEndian, x); err != nil {
		panic(fmt.Sprintf("(r *Reader) Uint8: %v", err))
	}
}

// 从阅读器阅读一个 uint16 并返回到 x 上
func (r *Reader) Uint16(x *uint16) {
	if err := binary.Read(r.r, binary.BigEndian, x); err != nil {
		panic(fmt.Sprintf("(r *Reader) Uint16: %v", err))
	}
}

// 从阅读器阅读一个 uint32 并返回到 x 上
func (r *Reader) Uint32(x *uint32) {
	if err := binary.Read(r.r, binary.BigEndian, x); err != nil {
		panic(fmt.Sprintf("(r *Reader) Uint32: %v", err))
	}
}

// 从阅读器阅读一个 uint64 并返回到 x 上
func (r *Reader) Uint64(x *uint64) {
	if err := binary.Read(r.r, binary.BigEndian, x); err != nil {
		panic(fmt.Sprintf("(r *Reader) Uint64: %v", err))
	}
}

// 从阅读器阅读一个 int8 并返回到 x 上
func (r *Reader) Int8(x *int8) {
	if err := binary.Read(r.r, binary.BigEndian, x); err != nil {
		panic(fmt.Sprintf("(r *Reader) Int8: %v", err))
	}
}

// 从阅读器阅读一个 int16 并返回到 x 上
func (r *Reader) Int16(x *int16) {
	if err := binary.Read(r.r, binary.BigEndian, x); err != nil {
		panic(fmt.Sprintf("(r *Reader) Int16: %v", err))
	}
}

// 从阅读器阅读一个 int32 并返回到 x 上
func (r *Reader) Int32(x *int32) {
	if err := binary.Read(r.r, binary.BigEndian, x); err != nil {
		panic(fmt.Sprintf("(r *Reader) Int32: %v", err))
	}
}

// 从阅读器阅读一个 int64 并返回到 x 上
func (r *Reader) Int64(x *int64) {
	if err := binary.Read(r.r, binary.BigEndian, x); err != nil {
		panic(fmt.Sprintf("(r *Reader) Int64: %v", err))
	}
}

// 从阅读器阅读一个布尔值并返回到 x 上
func (r *Reader) Bool(x *bool) {
	ans, err := r.ReadBytes(1)
	if err != nil {
		panic(fmt.Sprintf("(r *Reader) Bool: %v", err))
	}
	// get boolean
	switch ans[0] {
	case 0:
		*x = false
	case 1:
		*x = true
	default:
		panic(fmt.Sprintf("(r *Reader) Bool: Unexpected boolean %#v was find", ans[0]))
	}
	// set values
}

// 从阅读器阅读一个字符串并返回到 x 上
func (r *Reader) String(x *string) {
	var length uint16
	// init values
	if err := binary.Read(r.r, binary.BigEndian, &length); err != nil {
		panic(fmt.Sprintf("(r *Reader) String: %v", err))
	}
	// get the length of the target string
	p, err := r.ReadBytes(int(length))
	if err != nil {
		panic(fmt.Sprintf("(r *Reader) String: %v", err))
	}
	*x = string(p)
	// get the target string
}

// 从阅读器阅读一个字符串切片并返回到 x 上
func (r *Reader) StringSlice(x *[]string) {
	var length uint32
	// init values
	p, err := r.ReadBytes(1)
	if err != nil {
		panic(fmt.Sprintf("(r *Reader) StringSlice: %v", err))
	}
	if p[0] == 0 {
		*x = nil
	}
	// read header
	if err := binary.Read(r.r, binary.BigEndian, &length); err != nil {
		panic(fmt.Sprintf("(r *Reader) StringSlice: %v", err))
	}
	// get the length of the target slice
	*x = make([]string, length)
	for i := 0; i < int(length); i++ {
		r.String(&(*x)[i])
	}
	// get the target slice
}

// 从阅读器阅读一个 uuid.UUID 并返回到 x 上
func (r *Reader) UUID(x *uuid.UUID) {
	p, err := r.ReadBytes(16)
	if err != nil {
		panic(fmt.Sprintf("(r *Reader) UUID: %v", err))
	}
	// get the binary form of the uuid
	err = x.UnmarshalBinary(p)
	if err != nil {
		panic(fmt.Sprintf("(r *Reader) UUID: %v", err))
	}
	// unmarshal
}

// 从阅读器阅读一个 protocol.CommandOutputMessage 并返回到 x 上
func (r *Reader) CommandOutputMessage(x *protocol.CommandOutputMessage) {
	r.Bool(&x.Success)
	r.String(&x.Message)
	r.StringSlice(&x.Parameters)
}

// 从阅读器阅读一个 protocol.CommandOutputMessage 切片并返回到 x 上
func (r *Reader) CommandOutputMessageSlice(x *[]protocol.CommandOutputMessage) {
	var length uint32
	// init values
	p, err := r.ReadBytes(1)
	if err != nil {
		panic(fmt.Sprintf("(r *Reader) CommandOutputMessageSlice: %v", err))
	}
	if p[0] == 0 {
		*x = nil
	}
	// read header
	if err := binary.Read(r.r, binary.BigEndian, &length); err != nil {
		panic(fmt.Sprintf("(r *Reader) CommandOutputMessageSlice: %v", err))
	}
	// get the length of the target slice
	*x = make([]protocol.CommandOutputMessage, length)
	for i := 0; i < int(length); i++ {
		r.CommandOutputMessage(&(*x)[i])
	}
	// get the target slice
}
