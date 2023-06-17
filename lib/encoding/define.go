package encoding

import (
	"io"
)

const StringLengthMaxLimited = 65535     // 单个字符串的最大长度上限
const SliceLengthMaxLimited = 4294967295 // 单个切片的最大长度上限
const MapLengthMaxLimited = 65535        // 单个 Map 的最大长度上限

const UUIDConstantLength = 16 // 单个 uuid.UUID 的固定长度

// 用于读取二进制切片的阅读器
type Reader struct {
	r interface{ io.Reader }
}

// 用于写入二进制切片的写入者
type Writer struct {
	w interface{ io.Writer }
}
