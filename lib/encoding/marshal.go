package encoding

import (
	"encoding/binary"
	"fmt"
	"phoenixbuilder/minecraft/protocol"

	"github.com/google/uuid"
)

// 向写入者写入字节切片 p
func (w *Writer) WriteBytes(p []byte) error {
	_, err := w.w.Write(p)
	if err != nil {
		return fmt.Errorf("WriteBytes: %v", err)
	}
	return nil
}

// 向写入者写入二进制切片 x
func (w *Writer) Slice(x *[]byte) error {
	if len(*x) > SliceLengthMaxLimited {
		return fmt.Errorf("(w *Writer) Slice: The length of the target slice is out of the max limited %v", SliceLengthMaxLimited)
	}
	// check length
	err := binary.Write(w.w, binary.BigEndian, uint32(len(*x)))
	if err != nil {
		return fmt.Errorf("(w *Writer) Slice: %v", err)
	}
	// write the length of the target slice
	err = w.WriteBytes(*x)
	if err != nil {
		return fmt.Errorf("(w *Writer) Slice: %v", err)
	}
	// write slice
	return nil
	// return
}

// 向写入者写入字符串 x
func (w *Writer) String(x *string) error {
	if len(*x) > StringLengthMaxLimited {
		return fmt.Errorf("(w *Writer) String: The length of the target string is out of the max limited %v", StringLengthMaxLimited)
	}
	// check length
	err := binary.Write(w.w, binary.BigEndian, uint16(len(*x)))
	if err != nil {
		return fmt.Errorf("(w *Writer) String: %v", err)
	}
	// write the length of the target string
	err = w.WriteBytes([]byte(*x))
	if err != nil {
		return fmt.Errorf("(w *Writer) String: %v", err)
	}
	// write string
	return nil
	// return
}

// 向写入者写入字符串切片 x
func (w *Writer) StringSlice(x *[]string) error {
	if len(*x) > SliceLengthMaxLimited {
		return fmt.Errorf("(w *Writer) StringSlice: The length of the target slice is out of the max limited %v", SliceLengthMaxLimited)
	}
	// check length
	err := binary.Write(w.w, binary.BigEndian, uint32(len(*x)))
	if err != nil {
		return fmt.Errorf("(w *Writer) StringSlice: %v", err)
	}
	// write the length of the target slice
	for _, value := range *x {
		err = w.String(&value)
		if err != nil {
			return fmt.Errorf("(w *Writer) StringSlice: %v", err)
		}
	}
	// write slice
	return nil
	// return
}

// 向写入者写入 map[string][]byte
func (w *Writer) Map(x *map[string][]byte) error {
	if len(*x) > MapLengthMaxLimited {
		return fmt.Errorf("(w *Writer) Map: The length of the target map is out of the max limited %v", MapLengthMaxLimited)
	}
	// check length
	err := binary.Write(w.w, binary.BigEndian, uint16(len(*x)))
	if err != nil {
		return fmt.Errorf("(w *Writer) Map: %v", err)
	}
	// write the length of the target map
	for key, value := range *x {
		err := w.String(&key)
		if err != nil {
			return fmt.Errorf("(w *Writer) Map: %v", err)
		}
		err = w.Slice(&value)
		if err != nil {
			return fmt.Errorf("(w *Writer) Map: %v", err)
		}
	}
	// write map
	return nil
	// return
}

// 向写入者写入布尔值 x
func (w *Writer) Bool(x *bool) error {
	if *x {
		err := w.WriteBytes([]byte{1})
		if err != nil {
			return fmt.Errorf("(w *Writer) Bool: %v", err)
		}
	} else {
		err := w.WriteBytes([]byte{0})
		if err != nil {
			return fmt.Errorf("(w *Writer) Bool: %v", err)
		}
	}
	return nil
}

// 向写入者写入 x(uint8)
func (w *Writer) Uint8(x *uint8) error {
	err := w.WriteBytes([]byte{*x})
	if err != nil {
		return fmt.Errorf("(w *Writer) Uint8: %v", err)
	}
	return nil
}

// 向写入者写入 x(int8)
func (w *Writer) Int8(x *int8) error {
	err := binary.Write(w.w, binary.BigEndian, *x)
	if err != nil {
		return fmt.Errorf("(w *Writer) Int8: %v", err)
	}
	return nil
}

// 向写入者写入 x(uint16)
func (w *Writer) Uint16(x *uint16) error {
	err := binary.Write(w.w, binary.BigEndian, *x)
	if err != nil {
		return fmt.Errorf("(w *Writer) Uint16: %v", err)
	}
	return nil
}

// 向写入者写入 x(int16)
func (w *Writer) Int16(x *int16) error {
	err := binary.Write(w.w, binary.BigEndian, *x)
	if err != nil {
		return fmt.Errorf("(w *Writer) Int16: %v", err)
	}
	return nil
}

// 向写入者写入 x(uint32)
func (w *Writer) Uint32(x *uint32) error {
	err := binary.Write(w.w, binary.BigEndian, *x)
	if err != nil {
		return fmt.Errorf("(w *Writer) Uint32: %v", err)
	}
	return nil
}

// 向写入者写入 x(int32)
func (w *Writer) Int32(x *int32) error {
	err := binary.Write(w.w, binary.BigEndian, *x)
	if err != nil {
		return fmt.Errorf("(w *Writer) Int32: %v", err)
	}
	return nil
}

// 向写入者写入 x(uint64)
func (w *Writer) Uint64(x *uint64) error {
	err := binary.Write(w.w, binary.BigEndian, *x)
	if err != nil {
		return fmt.Errorf("(w *Writer) Uint64: %v", err)
	}
	return nil
}

// 向写入者写入 x(int64)
func (w *Writer) Int64(x *int64) error {
	err := binary.Write(w.w, binary.BigEndian, *x)
	if err != nil {
		return fmt.Errorf("(w *Writer) Int64: %v", err)
	}
	return nil
}

// 向写入者写入 x(uuid.UUID)
func (w *Writer) UUID(x *uuid.UUID) error {
	p, err := x.MarshalBinary()
	if err != nil {
		return fmt.Errorf("(w *Writer) UUID: %v", err)
	}
	// get binary of uuid
	if len(p) != UUIDConstantLength {
		return fmt.Errorf("(w *Writer) UUID: Unexpected length %v was find, but expected to be %v", len(p), UUIDConstantLength)
	}
	// check length
	err = w.WriteBytes(p)
	if err != nil {
		return fmt.Errorf("(w *Writer) UUID: %v", err)
	}
	// writer binary of uuid
	return nil
	// return
}

// 向写入者写入 x(protocol.CommandOutputMessage)
func (w *Writer) CommandOutputMessage(x *protocol.CommandOutputMessage) error {
	err := w.Bool(&x.Success)
	if err != nil {
		return fmt.Errorf("(w *Writer) CommandOutputMessage: %v", err)
	}
	// Success
	err = w.String(&x.Message)
	if err != nil {
		return fmt.Errorf("(w *Writer) CommandOutputMessage: %v", err)
	}
	// Message
	if x.Parameters == nil {
		err = w.WriteBytes([]byte{0, 0, 0, 0})
		if err != nil {
			return fmt.Errorf("(w *Writer) CommandOutputMessage: %v", err)
		}
	} else {
		err = w.StringSlice(&x.Parameters)
		if err != nil {
			return fmt.Errorf("(w *Writer) CommandOutputMessage: %v", err)
		}
	}
	// Parameters
	return nil
	// return
}

// 向写入者写入 x([]protocol.CommandOutputMessage)
func (w *Writer) CommandOutputMessageSlice(x *[]protocol.CommandOutputMessage) error {
	if len(*x) > SliceLengthMaxLimited {
		return fmt.Errorf("(w *Writer) CommandOutputMessageSlice: The length of x is out of the max limited %v", SliceLengthMaxLimited)
	}
	// check length
	err := binary.Write(w.w, binary.BigEndian, uint32(len(*x)))
	if err != nil {
		return fmt.Errorf("(w *Writer) CommandOutputMessageSlice: %v", err)
	}
	// write length
	for _, value := range *x {
		err = w.CommandOutputMessage(&value)
		if err != nil {
			return fmt.Errorf("(w *Writer) CommandOutputMessageSlice: %v", err)
		}
	}
	// write data
	return nil
	// return
}
