package main

import "phoenixbuilder/solutions/omega_lua/omega_lua/mux_pumper"

type LuaGoSystem struct {
	inputPumperMux *mux_pumper.InputPumperMux
}

func NewLuaGoSystem(inputPumperMux *mux_pumper.InputPumperMux) *LuaGoSystem {
	return &LuaGoSystem{
		inputPumperMux: inputPumperMux,
	}
}

func (m *LuaGoSystem) Print(msg string) {
	println(msg)
}

// UserInputChan 返回一个只读的字符串通道，用于监听用户输入
func (m *LuaGoSystem) UserInputChan() <-chan string {
	return m.inputPumperMux.NewListener()
}
