package mcstructure

import (
	"strconv"
	"strings"
)

// 一个简单的字符串阅读器
type StringReader struct {
	// 该字符串的完整内容
	Context string
	// 阅读器的指针位置，
	// 用于标识当前的阅读进度
	Pointer int
}

// 初始化一个阅读进度为 0 的新字符串阅读器，
// str 代表该阅读器所要阅读的完整字符串
func NewStringReader(str string) StringReader {
	return StringReader{
		Context: str,
		Pointer: 0,
	}
}

// 获得阅读器底层字符串中 location 处的字符。
// 返回的布尔值代表该字符是否存在
func (s *StringReader) GetCharacter(location int) (string, bool) {
	if location > len(s.Context)-1 {
		return "", false
	}
	return s.Context[location : location+1], true
}

// 获得当前指针处的字符。
// 返回的布尔值代表该字符是否存在
func (s *StringReader) GetCurrentCharacter() (string, bool) {
	if s.Pointer > len(s.Context)-1 {
		return "", false
	}
	return s.Context[s.Pointer : s.Pointer+1], true
}

// 以当前阅读进度为起始，
// 获得阅读器底层字符串中 length 长度的字符串。
// 返回的布尔值代表这样的字符是否存在
func (s *StringReader) GetString(length int) (string, bool) {
	if s.Pointer+int(length) > len(s.Context) {
		return "", false
	}
	return s.Context[s.Pointer : s.Pointer+length], true
}

// 以当前阅读进度为起始，
// 获得阅读器底层字符串中最近的一个非空格字符的位置。
// 特别地，换行符和缩进符也会当中空格处理。
// 返回的布尔值代表这样的字符是否存在
func (s *StringReader) GetCharacterWithNoSpace() (int, bool) {
	temp := s.Pointer
	for {
		current, exist := s.GetCharacter(temp)
		if !exist {
			return -1, false
		}
		if current != " " && current != "\n" && current != "\t" {
			return temp, true
		}
		temp++
	}
}

// 以当前阅读进度为起始，
// 从阅读器底层字符串识别一个字符串的闭合符 “"” ，
// 然后返回它的位置。
// 此函数会考虑转义符如 “\"” 的影响。
// 返回的布尔值则代表这样的闭合符是否存在
func (s *StringReader) GetRightBundary() (int, bool) {
	temp := s.Pointer
	for {
		current, exist := s.GetCharacter(temp)
		if !exist {
			return -1, false
		}
		if current == `\` {
			temp++
		}
		if current == `"` {
			return temp, true
		}
		temp++
	}
}

// 以当前阅读进度为起始，
// 从阅读器底层字符串识别一个整数。
// 阅读器的阅读进度会在读取到有效信息后自行改变。
// 返回的布尔值则代表这样的整数是否存在
func (s *StringReader) GetInt() (int, bool) {
	result := ""
	for {
		current, exist := s.GetCurrentCharacter()
		if !exist {
			break
		}
		if current == "-" || current == "+" || current == "0" || current == "1" || current == "2" || current == "3" || current == "4" || current == "5" || current == "6" || current == "7" || current == "8" || current == "9" {
			result = result + current
		} else {
			break
		}
		s.Pointer++
	}
	if len(result) == 0 {
		return -1, false
	}
	i, _ := strconv.ParseInt(result, 10, 64)
	return int(i), true
}

// 以当前阅读进度为起始，
// 从阅读器底层字符串识别一个布尔值。
// 阅读器的阅读进度会在读取到有效信息后自行改变。
// 返回的第一个布尔值代表解析结果，
// 返回的第二个布尔值代表这样的布尔值是否存在
func (s *StringReader) GetBool() (bool, bool) {
	get, exist := s.GetString(4)
	if !exist {
		return false, false
	}
	get = strings.ToLower(get)
	if get == "true" {
		s.Pointer = s.Pointer + 4
		return true, true
	}
	get, exist = s.GetString(5)
	if !exist {
		return false, false
	}
	get = strings.ToLower(get)
	if get == "false" {
		s.Pointer = s.Pointer + 5
		return false, true
	}
	return false, false
}
