package ResourcesControl

import "time"

// 描述命令请求的最长截止时间。
// 当超过此时间后，将会返回超时错误
const CommandRequestDeadLine = time.Second

// 描述命令请求中响应体的错误类型
const (
	CommandRequestOK = byte(iota)
	ErrCommandRequestNotRecord
	ErrCommandRequestConversionFailure
	ErrCommandRequestTimeOut
	ErrCommandRequestOthers
)
