package luaFrame

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/pterm/pterm"
)

const (
	HEADLUA    = "luas"
	HEADRELOAD = "reload"
	HEADSTART  = "start"
)
const (
	BINDINGFILE = "Binding.json"
)

// 指令信息 必须遵循 HEAD BEHAVIOR
type CmdMsg struct {
	isCmd    bool
	Head     string
	Behavior string
	args     []string
}
type PrintMsg struct {
	Type string
	Body interface{}
}

// 绑定函数 "名字":"逻辑实现的文件名"
type MappedBinding struct {
	Map map[string]string `json:"绑定"`
}

// 获取插件路径绝对路径 文件名字/插件名字
func GetComponentPath() []string {
	nameList := []string{}
	dirPath := OMGPATH + SEPA + "lua"
	fileExt := ".lua"
	// 读取目录下的所有文件
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		// 如果目录不存在，则创建它
		err := os.MkdirAll(dirPath, os.ModePerm)
		if err != nil {
			panic(err)
		}
		fmt.Println("Directory created:", dirPath)

	}
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		panic(err)
	}

	// 遍历目录下的所有文件名
	for _, file := range files {
		// 如果文件后缀名为 .lua，则打印文件名（去掉后缀名）
		if strings.HasSuffix(file.Name(), fileExt) {
			nameList = append(nameList, file.Name())
		}

	}
	return nameList
}

// 打印指定消息
func PrintInfo(str PrintMsg) {
	pterm.Info.Printfln("[%v][%v]: %v ", time.Now().YearDay(), str.Type, str.Body)
}

// 构造一个输出函数
func NewPrintMsg(typeName string, BodyString interface{}) PrintMsg {
	return PrintMsg{
		Type: typeName,
		Body: BodyString,
	}
}

// 获取data的相对位置omega_storage\\data
func GetRootPath() string {
	return OMGDATAPATH
}

// 获取"omega_storage\\data"
func GetDataPath() string {
	return OMGDATAPATH
}

// "omega_storage\\配置"
func GetOmgConfigPath() string {
	return OMGCONFIGPATH
}

// 针对binding.json文件进行的各种包装
// 获取binding.json的路径
func GetBindingPath() string {
	return GetRootPath() + SEPA + "lua" + SEPA + BINDINGFILE
}

// 获取data/lua/config
func GetConfigPath() string {
	return GetRootPath() + SEPA + "lua" + SEPA + "config"
}

// 安全地删除指定文件
func DelectFile(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	} else {
		// 文件存在，删除文件
		err := os.Remove(path)
		if err != nil {
			return err
		} else {
			return nil
		}
	}
}

// 格式化处理指令
func FormateCmd(str string) CmdMsg {

	words := strings.Fields(str)
	if len(words) < 3 {
		return CmdMsg{isCmd: false}
	}
	if words[0] != "lua" {
		return CmdMsg{isCmd: false}
	}
	head := words[1]
	//如果不属于任何指令则返回空cmdmsg
	if head != HEADLUA && head != HEADRELOAD && head != HEADSTART {
		return CmdMsg{isCmd: false}
	}
	behavior := words[2]
	args := []string{}
	if len(words) >= 3 {
		args = words[3:]
	}
	return CmdMsg{
		Head:     head,
		Behavior: behavior,
		args:     args,
		isCmd:    true,
	}
}

/*

文件管理系统

*/

type FileControl struct {
	//文件锁
	FileLock *FileLock
}

// 文件锁类型
type FileLock struct {
	mu sync.RWMutex
}

// 获取文件锁
func (lock *FileLock) Lock() {
	lock.mu.Lock()
}

// 释放文件锁
func (lock *FileLock) Unlock() {
	lock.mu.Unlock()
}

// 获取文件读锁
func (lock *FileLock) RLock() {
	lock.mu.RLock()
}

// 释放文件读锁
func (lock *FileLock) RUnlock() {
	lock.mu.RUnlock()
}

// 创建一个新的文件锁
func NewFileLock() *FileLock {
	return &FileLock{}
}

// 安全写入文件
func (f *FileControl) Write(filename string, data []byte) error {
	// 获取写锁
	lock := f.FileLock
	lock.Lock()
	defer lock.Unlock()

	// 写入数据
	if err := ioutil.WriteFile(filename, data, 0644); err != nil {
		return err
	}

	return nil
}

// 安全读取文件
func (f *FileControl) Read(filename string) ([]byte, error) {
	// 使用 os.Open 打开文件。
	file, err := os.Open(filename)
	if err != nil {
		return []byte{}, err
	}
	defer file.Close()

	// 获取文件信息，以确定要读取的字节数。
	_, err = file.Stat()
	if err != nil {
		return []byte{}, err
	}

	// 使用 ioutil.ReadAll 从文件中读取内容。
	content, err := ioutil.ReadAll(file)
	if err != nil {
		return content, err
	}
	return content, nil
}

// 读取并返回结构体
func (f *FileControl) ReadConfig(path string) (LuaCommpoentConfig, error) {
	newConfig := LuaCommpoentConfig{
		Disabled: true, //默认关闭
	}
	data, err := f.Read(path)
	if err != nil {
		return newConfig, err
	}

	err = json.Unmarshal(data, &newConfig)
	if err != nil {
		return newConfig, err
	}

	return newConfig, nil
}

// 删除插件
func (f *FileControl) DelectCompoentFile(name string) error {
	f.DeleteSubDir(name)
	//关闭相关内容
	PrintInfo(NewPrintMsg("提示", fmt.Sprintf("%v已经删除 干净了", name)))

	return nil
}

// deleteSubDir 函数接受一个父目录路径和一个子目录名称作为参数，
// 并安全地删除指定的子目录及其所有子文件。
func (f *FileControl) DeleteSubDir(subDirName string) error {
	parentDir := OMGCONFIGPATH
	subDir := filepath.Join(parentDir, subDirName)

	// 检查子目录是否存在。
	if !f.fileExists(subDir) {
		return nil
	}

	// 删除子目录及其所有子文件。
	err := os.RemoveAll(subDir)
	if err != nil {
		return err
	}

	return nil
}

// Result 结构体用于存储 JSON 文件和 Lua 文件的路径。
type Result struct {
	JsonFile   string
	LuaFile    string
	JsonConfig LuaCommpoentConfig
}

// GetLuaComponentPath返回一个包含同名字 JSON 文件和 Lua 文件路径的字典。
func (f *FileControl) GetLuaComponentData() (map[string]Result, error) {
	dir := OMGCONFIGPATH
	results := make(map[string]Result)

	// 使用 filepath.Walk 遍历指定目录及其子目录。
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 如果当前路径是一个目录，则检查是否存在与目录名同名的 JSON 和 Lua 文件。
		if info.IsDir() {
			dirName := info.Name()

			jsonFile := filepath.Join(path, dirName+".json")
			luaFile := filepath.Join(path, dirName+".lua")

			// 如果找到 JSON 和 Lua 文件，将它们的路径添加到结果字典中。
			if f.fileExists(jsonFile) && f.fileExists(luaFile) {
				//读取json文件
				config, err := f.ReadConfig(jsonFile)
				if err != nil {
					PrintInfo(NewPrintMsg("警告", err))
				}
				results[dirName] = Result{
					JsonFile:   jsonFile,
					LuaFile:    luaFile,
					JsonConfig: config,
				}
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return results, nil
}

// fileExists 函数接受一个文件路径作为参数，如果文件存在则返回 true，否则返回 false。
func (f *FileControl) fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// 在指定目录下创建具有指定名称的子目录，并在子目录中创建同名的 JSON 和 Lua 文件。
func (f *FileControl) CreateDirAndFiles(name string) error {
	// 创建子目录。
	dir := GetOmgConfigPath()
	data := LuaCommpoentConfig{
		Name:     name,
		Usage:    "",
		Version:  "0.0.1",
		Source:   "Lua-Component",
		Disabled: true,
		Author:   "",
		Config:   make(map[string]interface{}),
	}
	luaCode := `--根据注释初步了解如何书写代码
	--gameCtrol = skynet.GetControl()初始化操作机器人游戏行为的
	--gameCtrol.SendWsCmd("/say hellow") 发送指令
	--gameCtrol.SendCmdAndInvokeOnResponse("/say hellow") 发送指令并且返回一个表 内有是否成功 和返回信息两个值
	--h =gameCtrol.SendCmdAndInvokeOnResponse("/say hellow")
	--print(h.Success,"是否成功")
	--print(h.outputmsg,"输出信息")
	--listener = skynet.GetListener()
	--MsgListener = listener.GetMsgListner()
	--while true do
	--    print(MsgListener:NextMsg())   nextMsg可以读取玩家说话 如果玩家没有说话 那么就会堵塞直到玩家说话
	--end
	
	`

	subDir := filepath.Join(dir, name)
	// 检查目录是否已经存在，如果存在则返回错误。
	if _, err := os.Stat(subDir); !os.IsNotExist(err) {
		return errors.New("该名字的lua组件已经存在")
	}
	err := os.Mkdir(subDir, 0755)
	if err != nil {
		return err
	}

	// 创建 JSON 文件并根据指定结构体进行初始化。
	jsonFile := filepath.Join(subDir, name+".json")
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(jsonFile, jsonData, 0644)
	if err != nil {
		return err
	}

	// 创建 Lua 文件并根据指定字符串进行初始化。
	luaFile := filepath.Join(subDir, name+".lua")
	err = ioutil.WriteFile(luaFile, []byte(luaCode), 0644)
	if err != nil {
		return err
	}

	return nil
}
