package encoding

import (
	"encoding/binary"
	"fmt"
	"phoenixbuilder/minecraft/protocol"

	"github.com/google/uuid"
)

// 向写入者写入 x(uint8)
func (w *Writer) Uint8(x *uint8) {
	if err := binary.Write(w.w, binary.BigEndian, *x); err != nil {
		panic(fmt.Sprintf("(w *Writer) Uint8: %v", err))
	}
}

// 向写入者写入 x(uint16)
func (w *Writer) Uint16(x *uint16) {
	if err := binary.Write(w.w, binary.BigEndian, *x); err != nil {
		panic(fmt.Sprintf("(w *Writer) Uint16: %v", err))
	}
}

// 向写入者写入 x(uint32)
func (w *Writer) Uint32(x *uint32) {
	if err := binary.Write(w.w, binary.BigEndian, *x); err != nil {
		panic(fmt.Sprintf("(w *Writer) Uint32: %v", err))
	}
}

// 向写入者写入 x(uint64)
func (w *Writer) Uint64(x *uint64) {
	if err := binary.Write(w.w, binary.BigEndian, *x); err != nil {
		panic(fmt.Sprintf("(w *Writer) Uint64: %v", err))
	}
}

// 向写入者写入 x(int8)
func (w *Writer) Int8(x *int8) {
	if err := binary.Write(w.w, binary.BigEndian, *x); err != nil {
		panic(fmt.Sprintf("(w *Writer) Int8: %v", err))
	}
}

// 向写入者写入 x(int16)
func (w *Writer) Int16(x *int16) {
	if err := binary.Write(w.w, binary.BigEndian, *x); err != nil {
		panic(fmt.Sprintf("(w *Writer) Int16: %v", err))
	}
}

// 向写入者写入 x(int32)
func (w *Writer) Int32(x *int32) {
	if err := binary.Write(w.w, binary.BigEndian, *x); err != nil {
		panic(fmt.Sprintf("(w *Writer) Int32: %v", err))
	}
}

// 向写入者写入 x(int64)
func (w *Writer) Int64(x *int64) {
	if err := binary.Write(w.w, binary.BigEndian, *x); err != nil {
		panic(fmt.Sprintf("(w *Writer) Int64: %v", err))
	}
}

// 向写入者写入布尔值 x
func (w *Writer) Bool(x *bool) {
	if *x {
		if err := w.WriteBytes([]byte{1}); err != nil {
			panic(fmt.Sprintf("(w *Writer) Bool: %v", err))
		}
	} else {
		if err := w.WriteBytes([]byte{0}); err != nil {
			panic(fmt.Sprintf("(w *Writer) Bool: %v", err))
		}
	}
}

// 向写入者写入字符串 x
func (w *Writer) String(x *string) {
	if len(*x) > StringLengthMaxLimited {
		panic(fmt.Sprintf("(w *Writer) String: The length of the target string is out of the max limited %v", StringLengthMaxLimited))
	}
	// check length
	if err := binary.Write(w.w, binary.BigEndian, uint16(len(*x))); err != nil {
		panic(fmt.Sprintf("(w *Writer) String: %v", err))
	}
	// write the length of the target string
	if err := w.WriteBytes([]byte(*x)); err != nil {
		panic(fmt.Sprintf("(w *Writer) String: %v", err))
	}
	// write string
}

// 向写入者写入字符串切片 x
func (w *Writer) StringSlice(x *[]string) {
	if x == nil {
		if err := w.WriteBytes([]byte{0}); err != nil {
			panic(fmt.Sprintf("(w *Writer) StringSlice: %v", err))
		}
		return
	}
	// check pointer
	if len(*x) > SliceLengthMaxLimited {
		panic(fmt.Sprintf("(w *Writer) StringSlice: The length of the target slice is out of the max limited %v", SliceLengthMaxLimited))
	}
	// check length
	if err := w.WriteBytes([]byte{1}); err != nil {
		panic(fmt.Sprintf("(w *Writer) StringSlice: %v", err))
	}
	// write header
	if err := binary.Write(w.w, binary.BigEndian, uint32(len(*x))); err != nil {
		panic(fmt.Sprintf("(w *Writer) StringSlice: %v", err))
	}
	// write the length of the target slice
	for _, value := range *x {
		w.String(&value)
	}
	// write slice
}

// 向写入者写入 x(uuid.UUID)
func (w *Writer) UUID(x *uuid.UUID) {
	p, err := x.MarshalBinary()
	if err != nil {
		panic(fmt.Sprintf("(w *Writer) UUID: %v", err))
	}
	// marshal
	err = w.WriteBytes(p)
	if err != nil {
		panic(fmt.Sprintf("(w *Writer) UUID: %v", err))
	}
	// write result
}

// 向写入者写入 x(protocol.CommandOutputMessage)
func (w *Writer) CommandOutputMessage(x *protocol.CommandOutputMessage) {
	w.Bool(&x.Success)
	w.String(&x.Message)
	w.StringSlice(&x.Parameters)
}

// 向写入者写入 protocol.CommandOutputMessage 切片 x
func (w *Writer) CommandOutputMessageSlice(x *[]protocol.CommandOutputMessage) {
	if x == nil {
		if err := w.WriteBytes([]byte{0}); err != nil {
			panic(fmt.Sprintf("(w *Writer) CommandOutputMessageSlice: %v", err))
		}
		return
	}
	// check pointer
	if len(*x) > SliceLengthMaxLimited {
		panic(fmt.Sprintf("(w *Writer) CommandOutputMessageSlice: The length of the target slice is out of the max limited %v", SliceLengthMaxLimited))
	}
	// check length
	if err := w.WriteBytes([]byte{1}); err != nil {
		panic(fmt.Sprintf("(w *Writer) CommandOutputMessageSlice: %v", err))
	}
	// write header
	if err := binary.Write(w.w, binary.BigEndian, uint32(len(*x))); err != nil {
		panic(fmt.Sprintf("(w *Writer) CommandOutputMessageSlice: %v", err))
	}
	// write the length of the target slice
	for _, value := range *x {
		w.CommandOutputMessage(&value)
	}
	// write slice
}
