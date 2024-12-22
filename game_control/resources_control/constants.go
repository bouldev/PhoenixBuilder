package ResourcesControl

/*
 * This file is part of PhoenixBuilder.

 * PhoenixBuilder is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License.

 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.

 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.

 * Copyright (C) 2021-2025 Bouldev
 */

import "time"

// 描述命令请求的原始类型
const (
	// 标准命令
	CommandTypeStandard = "packet.CommandRequest"
	// 网易魔法指令
	CommandTypeAICommand = "packet.PyRpc/C2SModEvent/Minecraft/aiCommand/ExecuteCommandEvent"
)

// 描述与命令请求相关的信号
const (
	// 魔法指令 的命令请求在收到
	// 标准响应体后会使用的信号
	SignalRespondReceived = uint8(iota)
	// 命令响应体可以被加载时会使用的信号
	SignalCouldLoadRespond
)

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
