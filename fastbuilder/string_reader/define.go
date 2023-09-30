package string_reader

// 描述一个字符串阅读器
type StringReader struct {
	ptr int     // 指代当前的阅读进度
	ctx *string // 指代该阅读器所包含的底层字符串
}

// 返回以 content 为底层的字符串阅读器
func NewStringReader(content *string) *StringReader {
	reader := StringReader{}
	reader.Reset(content)
	return &reader
}
