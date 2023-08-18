package encoding

import (
	"bytes"
	"fmt"
	"phoenixbuilder/game_control/websocket_api/interfaces"
)

// 创建一个新的阅读器
func NewReader(reader *bytes.Buffer) interfaces.IO {
	return &Reader{r: reader}
}

// 创建一个新的写入者
func NewWriter(writer *bytes.Buffer) interfaces.IO {
	return &Writer{w: writer}
}

// 从阅读器阅读 length 个字节
func (r *Reader) ReadBytes(length int) ([]byte, error) {
	ans := make([]byte, length)
	_, err := r.r.Read(ans)
	if err != nil {
		return nil, fmt.Errorf("ReadBytes: %v", err)
	}
	return ans, nil
}

// 向写入者写入字节切片 p
func (w *Writer) WriteBytes(p []byte) error {
	_, err := w.w.Write(p)
	if err != nil {
		return fmt.Errorf("WriteBytes: %v", err)
	}
	return nil
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
