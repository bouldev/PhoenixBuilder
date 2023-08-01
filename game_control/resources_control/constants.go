package ResourcesControl

import "time"

// 描述请求的最长截止时间
const (
	// 描述命令请求的最长截止时间。
	// 当超过此时间后，将会返回超时错误
	CommandRequestNoDeadLine      = 0
	CommandRequestDefaultDeadLine = time.Second
	// 描述容器操作(打开/关闭)的最长截止时间。
	// 当超过此时间后，将不再等待
	ContainerOperationDeadLine = time.Second
)

// 描述命令请求中响应体的错误类型
const (
	CommandRequestOK = byte(iota)
	ErrCommandRequestNotRecord
	ErrCommandRequestConversionFailure
	ErrCommandRequestTimeOut
	ErrCommandRequestOthers
)

// 描述单个数据包监听器中允许的最大协程运行数量
const MaximumCoroutinesRunningCount int32 = 255
