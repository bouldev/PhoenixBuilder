package monk

import (
	"phoenixbuilder/solutions/omega_lua/omega_lua/mux_pumper"
	"time"

	"github.com/google/uuid"
)

// 终端输入输出管理中心
var inputPumperMux *mux_pumper.InputPumperMux

// 开始终端资源分配中心
func startInputSource() {
	// 创建一个新的 inputPumperMux 实例
	inputPumperMux = mux_pumper.NewInputPumperMux()
	// 创建一个定时任务，每隔两秒向 inputPumperMux 发送一条随机字符串
	go func() {
		for {
			time.Sleep(time.Second * 2)
			input := "hello: " + uuid.New().String()
			inputPumperMux.PumpInput(input)
		}
	}()
}

// 在程序启动时调用 startInputSource 和 startGamePacketSource 启动资源分配中心
func init() {
	go startInputSource()
}

type MonkSystem struct {
}

func NewMonkSystem() *MonkSystem {
	return &MonkSystem{}
}

func (m *MonkSystem) Print(msg string) {
	println(msg)
}

// UserInputChan 返回一个只读的字符串通道，用于监听用户输入
func (m *MonkSystem) UserInputChan() <-chan string {
	return inputPumperMux.NewListener()
}
