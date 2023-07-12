package mux_pumper

import (
	"phoenixbuilder/fastbuilder/lib/utils/sync_wrapper"

	"github.com/google/uuid"
)

// InputPumperMux 结构体，表示输入数据的处理器
// 这个处理器的基本思路是，
// 当一个新的输入数据到来时
// 它将被发送到所有的监听器中
// 而每个监听器则可以根据自己的需求对输入数据进行处理
// 多个监听器可以同时监听同一个输入源
// 从而实现了多路输入处理的功能。
type InputPumperMux struct {
	// 管理多个输入监听器的 SyncMap 保证数据包不会堵塞
	pumpers *sync_wrapper.SyncMap[chan string]
}

// NewInputPumperMux 创建一个新的 InputPumperMux 实例
func NewInputPumperMux() *InputPumperMux {
	// 返回一个新的 InputPumperMux 对象，其中 pumpers 字段为一个空的 SyncMap
	return &InputPumperMux{
		pumpers: sync_wrapper.NewInstanceMap[chan string](),
	}
}

// PumpInput 将输入数据传递给所有监听器
func (i *InputPumperMux) PumpInput(input string) {
	// 保存当前的监听器列表到 currentPumper
	currentPumper := i.pumpers
	// 创建一个新的空的监听器列表
	i.pumpers = sync_wrapper.NewInstanceMap[chan string]()
	// 遍历 currentPumper 中的每个监听器
	currentPumper.Iter(func(k string, listener chan string) (continueInter bool) {
		// 将输入数据发送到监听器的通道中
		select {
		case listener <- input:
		default:
		}
		return true
	})
}

// NewListener 创建一个新的监听器，并将其添加到 pumpers 中
func (i *InputPumperMux) NewListener() chan string {
	// 创建一个新的字符串类型的通道作为监听器的值
	listener := make(chan string)
	// 将新创建的监听器添加到 pumpers 中
	i.pumpers.Set(uuid.New().String(), listener)
	// 返回新创建的监听器
	return listener
}
