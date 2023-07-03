package little_endian_with_varint

import "errors"

var ErrStringLengthExceeds = errors.New("string length exceeds maximum length prefix")
