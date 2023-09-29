package mc_command_reader

import "fmt"

// 描述一个单个的命令阅读器
type CommandReader struct {
	ptr int     // 指代当前的阅读进度
	ctx *string // 指代该阅读器所包含的底层字符串
}

// 返回以 content 为底层的命令阅读器
func NewCommandReader(content *string) *CommandReader {
	reader := CommandReader{}
	reader.Reset(content)
	return &reader
}

// 检查阅读器(阅读进度及底层字符串)是否正确
func (c *CommandReader) states_check() {
	if c.ctx == nil && c.ptr > len(*c.ctx) {
		panic("states_check: EOF")
	}
}

// 返回阅读器上的底层字符串。
// 如果底层为空指针，则返回空字符串
func (c *CommandReader) String() string {
	if c.ctx == nil {
		return ""
	}
	return *c.ctx
}

// 返回当前的阅读进度
func (c *CommandReader) Pointer() int {
	return c.ptr
}

// 将底层字符串重设为 str 并清空阅读进度
func (c *CommandReader) Reset(str *string) {
	c.ctx = str
	c.ptr = 0
}

// 设置当前阅读进度为 new 。
// 若底层字符串为空指针，
// 或提供的新进度大于底层字符串的长度，
// 则将导致崩溃
func (c *CommandReader) SetPtr(new int) {
	if c.ctx == nil || new > len(*c.ctx) {
		switch {
		case c.ctx == nil:
			panic(fmt.Sprintf("SetPtr: Failed to set pointer to %d because c.ctx is nil", new))
		case new > len(*c.ctx):
			panic(fmt.Sprintf("SetPtr: Failed to set pointer to %d because of EOF error", new))
		}
	}
	c.ptr = new
}

// 从底层字符串阅读 length 个字符。
// 若阅读进度超出底层字符串长度，
// 则将会造成崩溃
func (c *CommandReader) Sentence(length int) string {
	if length < 0 {
		panic(fmt.Sprintf("Sentence: The length provided is less than 0; length = %d", length))
	}
	c.states_check()
	if c.ptr+length > len(*c.ctx) {
		panic("Sentence: EOF")
	}
	str := (*c.ctx)[c.ptr : c.ptr+length]
	c.ptr = c.ptr + length
	return str
}

// 从底层字符串阅读一个字符。
// 可能会返回 EOF 的崩溃
func (c *CommandReader) Next() string {
	return c.Sentence(1)
}

// 以 startPoint 为起始阅读 length 个字符，
// 然后返回对应的字符串。
// 特别地，如果 length 为 nil ，
// 则返回 startPoint 至当前位置之间的字符串
func (c *CommandReader) SentenceThroughPtr(startPoint int, length *int) string {
	if length == nil {
		new := c.Pointer() - startPoint
		length = &new
	}
	c.SetPtr(startPoint)
	return c.Sentence(*length)
}
