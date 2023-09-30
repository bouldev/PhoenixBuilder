package string_reader

import "fmt"

// 返回阅读器上的底层字符串。
// 如果底层为空指针，则返回空字符串
func (s *StringReader) String() string {
	if s.ctx == nil {
		return ""
	}
	return *s.ctx
}

// 返回当前的阅读进度
func (s *StringReader) Pointer() int {
	return s.ptr
}

// 将底层字符串重设为 str 并清空阅读进度
func (s *StringReader) Reset(str *string) {
	s.ctx = str
	s.ptr = 0
}

// 设置当前阅读进度为 new 。
// 若底层字符串为空指针，
// 或提供的新进度大于底层字符串的长度，
// 则将导致崩溃
func (s *StringReader) SetPtr(new int) {
	if s.ctx == nil || new > len(*s.ctx) {
		switch {
		case s.ctx == nil:
			panic(fmt.Sprintf("SetPtr: Failed to set pointer to %d because c.ctx is nil", new))
		case new > len(*s.ctx):
			panic(fmt.Sprintf("SetPtr: Failed to set pointer to %d because of EOF error", new))
		}
	}
	s.ptr = new
}

// 从底层字符串阅读 length 个字符。
// 即使阅读量超出阅读器剩余的可读量，
// 亦不会造成崩溃
func (s *StringReader) Sentence(length int) string {
	if length < 0 {
		panic(fmt.Sprintf("Sentence: The length provided is less than 0; length = %d", length))
	}
	if s.ctx == nil || s.ptr > len(*s.ctx) {
		panic("Sentence: Pointer was broken")
	}
	if l := len(*s.ctx); s.ptr+length > l {
		res := (*s.ctx)[s.ptr:]
		s.ptr = l
		return res
	}
	str := (*s.ctx)[s.ptr : s.ptr+length]
	s.ptr = s.ptr + length
	return str
}

// 尝试从底层字符串阅读一个字符，
// 但可能会因为已抵达阅读上限而失败，
// 此时若 avoidEOF 为假，
// 则将抛出名为 EOF 的崩溃
func (s *StringReader) Next(avoidEOF bool) string {
	if str := s.Sentence(1); !avoidEOF && len(str) == 0 {
		panic("Next: EOF")
	} else {
		return str
	}
}

// 以 startPoint 为起始，
// 截取其至当前阅读进度处的字符并返回
func (s *StringReader) CutSentence(startPoint int) string {
	length := s.ptr - startPoint
	s.SetPtr(startPoint)
	return s.Sentence(length)
}
