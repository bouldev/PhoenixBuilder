package websocket_api

import (
	"phoenixbuilder/game_control/websocket_api/interfaces"
)

// 描述单个请求所对应的响应
type Responce struct {
	// 指代此响应包所对应的请求包的回声。
	// 只有此响应包对应的请求包中提供了此字段时才会存在，
	// 否则为空字符串
	Echo string `json:"request_id"`
	// 指代对应请求的完成时间
	FinishTime string `json:"finish_time"`
	// 指代请求的函数的返回值
	FuncReturn interfaces.Return `json:"function_return"`
}

// Marshal 提供了双向实现，
// 以允许将 Responce 结构体编码/解码为二进制数据，
// 然后在网络上进行传输
func (r *Responce) Marshal(io interfaces.IO) {
	io.String(&r.Echo)
	io.String(&r.FinishTime)
	r.FuncReturn.AutoMarshal(io)
}
