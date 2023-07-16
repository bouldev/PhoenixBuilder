package main

import (
	"context"
	"io/ioutil"
	"path"
	"phoenixbuilder/fastbuilder/lib/utils/file_wrapper"
	"phoenixbuilder/solutions/omega_lua/monk"
	"phoenixbuilder/solutions/omega_lua/omega_lua"
	"phoenixbuilder/solutions/omega_lua/omega_lua/concurrent"
	"regexp"
	"strconv"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

func ReadOutAllExamplesHelper(targetDir string) map[int]string {
	// 从example目录下读取所有的lua文件, 文件名规则为 01xxx.lua,02xxx.lua
	// 输出类似于 map[int]string{1:"01xxx.lua",2:"02xxx.lua"}
	fs, err := ioutil.ReadDir(targetDir)
	if err != nil {
		panic(err)
	}
	result := make(map[int]string)
	for _, f := range fs {
		fileName := f.Name()
		if !strings.HasSuffix(fileName, ".lua") {
			continue
		}
		// 使用正则匹配读取文件名开头的数字
		r, _ := regexp.Compile(`^\d+`)
		idx := r.FindString(fileName)
		tmp, err := strconv.Atoi(idx)
		if err != nil {
			panic(err)
		}
		code, _ := file_wrapper.GetFileData(path.Join(targetDir, fileName))
		result[tmp] = string(code)
	}
	return result
}

func CreateLuaEnv(ctx context.Context) (ac concurrent.AsyncCtrl, L *lua.LState) {
	L = lua.NewState()
	ac = concurrent.NewAsyncCtrl(ctx)
	// go implements
	// 1. monk system
	goSystem := monk.NewMonkSystem()
	goPackets := monk.NewMonkPackets(128)
	goCmdSender := monk.NewMonkCmdSender()
	// lua wrapper
	return omega_lua.CreateOmegaLuaEnv(ctx, &omega_lua.GoImplements{
		GoSystem:         goSystem,
		GoPackets:        goPackets,
		GoPacketProvider: goPackets,
		GoCmdSender:      goCmdSender,
	})
}

func main() {
	// read lua
	//测试用读取的packet.lua
	allCodes := ReadOutAllExamplesHelper("examples")
	if len(allCodes) == 0 {
		panic("examples not found, check your current work dir")
	}
	ac, L := CreateLuaEnv(context.Background())
	exampleIdx := 4 // 选择要运行的示例, 1,2,3,4,...
	errChan := concurrent.FireLuaCodeInGoRoutine(ac, L, allCodes[exampleIdx])
	// wait for lua code to finish
	err := <-errChan
	if err != nil {
		panic(err)
	}
	ac.Wait()
}
