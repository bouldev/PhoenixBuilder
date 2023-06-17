package encoding

import (
	"bytes"
	"fmt"
)

// 创建一个新的阅读器
func NewReader(reader *bytes.Buffer) IO {
	return &Reader{r: reader}
}

// 创建一个新的写入者
func NewWriter(writer *bytes.Buffer) IO {
	return &Writer{w: writer}
}

// 取得阅读器的底层切片
func (r *Reader) GetBuffer() (*bytes.Buffer, error) {
	ans, success := r.r.(*bytes.Buffer)
	if !success {
		return nil, fmt.Errorf("(r *Reader) GetBuffer: Failed to convert r.r into *bytes.Buffer; r.r = %#v", r.r)
	}
	return ans, nil
}

// 取得写入者的底层切片
func (w *Writer) GetBuffer() (*bytes.Buffer, error) {
	ans, success := w.w.(*bytes.Buffer)
	if !success {
		return nil, fmt.Errorf("(w *Writer) GetBuffer: Failed to convert w.w into *bytes.Buffer; w.w = %#v", w.w)
	}
	return ans, nil
}
