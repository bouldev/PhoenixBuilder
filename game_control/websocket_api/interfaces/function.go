package interfaces

import (
	GameInterface "phoenixbuilder/game_control/game_interface"
)

// 指代要向函数传入的所有参数
type Input interface {
	// 用于执行当次请求的实现
	Run(env *GameInterface.GameInterface) Return
	// 自动化 解码/编码 实现
	AutoMarshal(io IO)
}

// 指代函数的返回值
type Return interface {
	// 自动化 解码/编码 实现
	AutoMarshal(io IO)
}
