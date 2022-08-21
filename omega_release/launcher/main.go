package main

import (
	"bufio"
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"net/http"
	"net/url"
	"omega_launcher/embed_binary"
	"omega_launcher/utils"
	. "omega_launcher/variants"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/pterm/pterm"
)

var STOARGE_REPO = "https://omega.fastbuilder.pro/binary"

type BotConfig struct {
	Code   string `json:"租赁服号"`
	Passwd string `json:"租赁服密码"`
	Token  string `json:"FBToken"`
}

func LookForPossibleRepo() error {
	connectSuccess := false
	successCtx := make(chan bool)
	for _, url := range []string{
		"https://omega.fastbuilder.pro/binary",
	} {
		_url := url
		go func() {
			if _, err := http.Get(_url + "/TIME_STAMP"); err == nil {
				if !connectSuccess {
					connectSuccess = true
					close(successCtx)
					STOARGE_REPO = _url
					return
				}
			}
		}()
		time.Sleep(3 * time.Millisecond)
	}
	select {
	case <-successCtx:
		return nil
	case <-time.NewTimer(time.Second).C:
	}
	if !connectSuccess {
		return fmt.Errorf("无法连接到更新服务器")
	}
	return nil
}

func main() {
	defer func() {
		r := recover()
		if r != nil {
			pterm.Error.Println("运行中出现错误: ", r)
			b := make([]byte, 1)
			fmt.Println("按下任意键退出")
			os.Stdin.Read(b)
		}
	}()
	PrintVariant()
	pterm.Info.Println("当前路径: " + GetCurrentDir())
	if err := LookForPossibleRepo(); err != nil {
		pterm.Info.Printf("无法连接更新服务器，要尝试使用本地的程序吗? 要请输入 y 不要请输入 n ")
		accept := utils.GetInputYN()
		if !accept {
			panic(err)
		}
	} else {
		pterm.Info.Println("已连接到更新服务器, 开始检查更新")
		targetHash := GetRemoteOmegaHash()
		currentHash := GetCurrentOmegaHash()
		if targetHash == currentHash {
			pterm.Success.Println("太好了，你的程序已经是最新的了!")
		} else {
			pterm.Warning.Println("我们将为你下载最新程序, 请保持耐心...")
			DownloadOmega()
		}
	}

	pterm.Info.Printf("这次是要使用 Omega 还是仅仅使用 FB ? \n要使用 Omega 请输入 y 要使用 FB 请输入 n :")
	accept := utils.GetInputYN()
	if accept {
		if err := os.Chdir(GetCurrentDir()); err != nil {
			panic(err)
		}
			CQHttpEnablerHelper()
		StartOmegaHelper()
	} else {
		RunFB()
	}
}

type QGroupLink struct {
	Address                   string                        `json:"CQHTTP正向Websocket代理地址"`
	GameMessageFormat         string                        `json:"游戏消息格式化模版"`
	QQMessageFormat           string                        `json:"Q群消息格式化模版"`
	Groups                    map[string]int64              `json:"链接的Q群"`
	Selector                  string                        `json:"游戏内可以听到QQ消息的玩家的选择器"`
	NoBotMsg                  bool                          `json:"不要转发机器人的消息"`
	ChatOnly                  bool                          `json:"只转发聊天消息"`
	MuteIgnored               bool                          `json:"屏蔽其他群的消息"`
	FilterQQToServerMsgByHead string                        `json:"仅仅转发开头为以下特定字符的消息到服务器"`
	FilterServerToQQMsgByHead string                        `json:"仅仅转发开头为以下特定字符的消息到QQ"`
	AllowedCmdExecutor        map[int64]bool                `json:"允许这些人透过QQ执行命令"`
	AllowdFakeCmdExecutor     map[int64]map[string][]string `json:"允许这些人透过QQ执行伪命令"`
	DenyCmds                  map[string]string             `json:"屏蔽这些指令"`
}

type ComponentConfig struct {
	Name        string      `json:"名称"`
	Description string      `json:"描述"`
	Disabled    bool        `json:"是否禁用"`
	Version     string      `json:"版本"`
	Source      string      `json:"来源"`
	Configs     *QGroupLink `json:"配置"`
	upgradeFn   func(*ComponentConfig) error
}

func (c *ComponentConfig) Upgrade() error {
	return c.upgradeFn(c)
}

func (c *ComponentConfig) SetUpgradeFn(fn func(*ComponentConfig) error) (ok bool) {
	if c.upgradeFn == nil {
		c.upgradeFn = fn
		return true
	} else {
		return false
	}
}

var QGroupLinkCoinfigPath string

func AcquireQGroupLinkConfig() string {
	if QGroupLinkCoinfigPath != "" {
		return QGroupLinkCoinfigPath
	}
	cfgPath := path.Join(GetOmegaStorageDir(), "配置")
	if !utils.IsDir(cfgPath) {
		return ""
	}
	if err := filepath.Walk(cfgPath, func(filePath string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
	}
		if runtime.GOOS == "windows" {
			filePath = strings.ReplaceAll(filePath, "\\", "/")
		}
		fileBaseName := path.Base(filePath)
		if !strings.HasPrefix(fileBaseName, "组件") || !strings.HasSuffix(fileBaseName, ".json") {
			return nil
		}
		c := &ComponentConfig{}
		if err := utils.GetJsonData(filePath, c); err != nil {
			return fmt.Errorf("处理[" + filePath + "]时出错" + err.Error())
		}
		if c.Name == "群服互通" {
			QGroupLinkCoinfigPath = filePath
	}
		return nil
	}); err != nil {
		panic(err)
}
	return QGroupLinkCoinfigPath
}

func GetOmegaStorageDir() string {
	if GetPlantform() == embed_binary.Android_arm64 {
		if utils.IsDir("/sdcard/Download/omega_storage") {
			return "/sdcard/Download/omega_storage"
		}
	}
	return path.Join(GetCurrentDir(), "omega_storage")
}

func GetCQHttpDir() string {
	return path.Join(GetCurrentDir(), "cqhttp_storage")
}

//go:embed config.yml
var defaultConfigBytes []byte

func CQHttpLoadHelper() (isImport bool) {
	fileName := path.Join(GetOmegaStorageDir(), "上传这个文件到云服务器以使用云服务器的群服互通.data")
	if utils.IsFile(fileName) {
		var fp *os.File
		defer func() {
			if fp != nil {
				fp.Close()
			}
		}()
		unzipSize, err := utils.GetUnZipSize(fileName)
		if err != nil {
			panic(err)
		}
		fp, err = os.OpenFile(fileName, os.O_RDONLY, 0755)
		if err != nil {
			panic(err)
		}
		uuidBytes := make([]byte, 36)
		if _, err := fp.Read(uuidBytes); err != nil {
			panic(err)
		}
		fmt.Println(string(uuidBytes))
		uuidFile := path.Join(GetCQHttpDir(), "uuid")
		if utils.IsFile(uuidFile) {
			if thisUUidBytes, err := utils.GetFileData(uuidFile); err == nil {
				if string(thisUUidBytes) == string(uuidBytes) {
					return false
				}
			}
		}
		pterm.Info.Printf("可以从 %v 中导入群服互通配置，要导入吗? 要请输入 y 不要请输入 n ", fileName)
		accept := utils.GetInputYN()
		if accept {
			isImport = true
			os.RemoveAll(GetCQHttpDir())
			zipData, err := ioutil.ReadAll(fp)
			if err != nil {
				panic(err)
			}
			if err := utils.UnZip(bytes.NewReader(zipData), unzipSize, GetCQHttpDir()); err != nil {
				panic(err)
			}
			if _, err := utils.CopyFile(path.Join(GetCQHttpDir(), "组件-群服互通.json"), AcquireQGroupLinkConfig()); err != nil {
				panic(err)
			}
			pterm.Success.Println("导入应该成功了")
		}

	}
	return false
}

func CQHttpEnablerHelper() {
	qGroupLinkCfgPath := AcquireQGroupLinkConfig()
	if qGroupLinkCfgPath == "" {
		return
	}
	qGroupLinkCfg := &ComponentConfig{}
	if err := utils.GetJsonData(qGroupLinkCfgPath, &qGroupLinkCfg); err != nil {
		pterm.Error.Println(err)
		return
	}
	if qGroupLinkCfgPath == "" {
		pterm.Info.Println("群服互通辅助配置程序将在第二次成功启动 Omega 时启用，现在是第一次启动 Omega")
		return
	}
	isImport := CQHttpLoadHelper()
	pterm.Info.Printf("要启用群服互通吗 要请输入 y 不要请输入 n ")
	accept := utils.GetInputYN()
	if !accept {
		utils.WriteJsonData(qGroupLinkCfgPath, qGroupLinkCfg)
		return
	}
	if !utils.IsFile(GetCqHttpExec()) {
		if err := utils.WriteFileData(GetCqHttpExec(), GetCqHttpBinary()); err != nil {
			panic(err)
		}
	}
	utils.MakeDirP(GetCQHttpDir())
	configFile := path.Join(GetCQHttpDir(), "config.yml")
	accept = true
	if utils.IsFile(configFile) {
		pterm.Info.Printf("要接受现有QQ号配置吗？要请输入 y 修改请输入 n ")
		accept = utils.GetInputYN()
	}
	if IsMCSM() {
		os.RemoveAll(path.Join(GetCQHttpDir(), "data"))
		os.RemoveAll(path.Join(GetCQHttpDir(), "logs"))
		if !utils.IsFile(configFile) || !accept {
			pterm.Error.Println("对于面板服，你只能使用上传登录文件的方式登录，详情请见群文件")
			return
		} else {
			pterm.Warning.Println("将使用 " + configFile + " 的配置进行 QQ 登录，请不要修改这份文件，否则会出错")
			pterm.Warning.Println("将使用 " + qGroupLinkCfgPath + " 的配置进行群服互通，您可以自行修改这份文件")
		}
	} else {
		pterm.Success.Println("如果你需要群服互通，请保证这行字能完整显示在一行中（你可以双指捏合缩放）-->|")
		pterm.Error.Println("如果你正在手机上使用群服互通，请务必用另一部手机扫码登录，不能截图！")
		if !utils.IsFile(configFile) || !accept {
			os.RemoveAll(GetCQHttpDir())
			utils.MakeDirP(GetCQHttpDir())
			pterm.Info.Printf("请输入QQ账号: ")
			Code := utils.GetValidInput()
			pterm.Info.Printf("请输入QQ密码（想扫码登录则留空）: ")
			Passwd := utils.GetInput()
			if Passwd == "" {
				Passwd = "''"
			}
			defaultConfigStr := string(defaultConfigBytes)
			cfgStr := strings.ReplaceAll(defaultConfigStr, "[QQ账号]", Code)
			cfgStr = strings.ReplaceAll(cfgStr, "[QQ密码]", Passwd)
			utils.WriteFileData(configFile, []byte(cfgStr))
			pterm.Info.Printf("请输入想链接的群号: ")
			GroupCode := utils.GetValidInt()
			for k, _ := range qGroupLinkCfg.Configs.Groups {
				qGroupLinkCfg.Configs.Groups[k] = int64(GroupCode)
			}
			utils.WriteJsonData(qGroupLinkCfgPath, qGroupLinkCfg)
		}
		pterm.Warning.Println("将使用 " + configFile + " 的配置进行 QQ 登录，您可以自行修改这份文件")
		pterm.Warning.Println("将使用 " + qGroupLinkCfgPath + " 的配置进行群服互通，您可以自行修改这份文件")
	}

	if portNumber, err := utils.GetAvailablePort(); err != nil {
		panic(err)
	} else {
		RunCQHttp(isImport, portNumber)
	}

}

func WaitConnect(portNumber int) {
	for {
		u := url.URL{Scheme: "ws", Host: fmt.Sprintf("127.0.0.1:%v", portNumber)}
		var err error
		_, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			time.Sleep(time.Second)
			continue
		} else {
			return
		}
	}
}

func PackCQHttpRunAuth(isImport bool) {
	_uuid, _ := uuid.NewUUID()
	uuid := _uuid.String()
	uuidFile := path.Join(GetCQHttpDir(), "uuid")
	if !isImport {
		if err := utils.WriteFileData(uuidFile, []byte(uuid)); err != nil {
			panic(err)
		}
		omegaConfigFile := AcquireQGroupLinkConfig()
		if _, err := utils.CopyFile(omegaConfigFile, path.Join(GetCQHttpDir(), "组件-群服互通.json")); err != nil {
			panic(err)
		}
		fileName := path.Join(GetOmegaStorageDir(), "上传这个文件到云服务器以使用云服务器的群服互通.data")
		fp, err := os.OpenFile(fileName, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
		fp.Write([]byte(uuid))
		if err != nil {
			panic(err)
		}
		if err := utils.Zip(GetCQHttpDir(), fp, []string{"data", "logs"}); err != nil {
			panic(err)
		}
		fp.Close()

		for i := 0; i < 3; i++ {
			pterm.Success.Printfln("你可以将文件[%v]上传到云服务器的 omega_storage 中以便云服务器使用群服互通", fileName)
		}
		time.Sleep(2)
	}
}

func AlterPort(portNumber int) {
	qGroupLinkCfgPath := AcquireQGroupLinkConfig()
	CQConfigFile := path.Join(GetCQHttpDir(), "config.yml")
	qGroupLinkCfg := &ComponentConfig{}
	if err := utils.GetJsonData(qGroupLinkCfgPath, &qGroupLinkCfg); err != nil {
				panic(err)
			}
	qGroupLinkCfg.Configs.Address = fmt.Sprintf("127.0.0.1:%v", portNumber)
	if err := utils.WriteJsonData(qGroupLinkCfgPath, qGroupLinkCfg); err != nil {
		panic(err)
	}
	if xp, err := regexp.Compile(`port.*:.*\d+`); err == nil {
		if srcBytes, err := ioutil.ReadFile(CQConfigFile); err == nil {
			dstBytes := xp.ReplaceAll(srcBytes, []byte(fmt.Sprintf("port: %v", portNumber)))
			if err := ioutil.WriteFile(CQConfigFile, dstBytes, 0755); err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	} else {
		panic(err)
	}
}

func RunCQHttp(isImport bool, portNumber int) {
	if !isImport {
		pterm.Info.Println("如果你扫码有困难，也可以 直接扫码当前文件夹下的 qrcode.png 它们是一样的")
	}
	AlterPort(portNumber)
	cmd := exec.Command(GetCqHttpExec(), "-faststart")
	cmd.Dir = GetCQHttpDir()
	cqHttpOut, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	connected := false
	go func() {
		reader := bufio.NewReader(cqHttpOut)
		for {
			readString, err := reader.ReadString('\n')
			if err != nil || err == io.EOF {
				fmt.Print(readString)
				pterm.Warning.Println("CQHTTP 已退出")
				return
			}
			if !connected {
				fmt.Print(readString)
			} else {
				fmt.Print("CQ:" + readString)
			}
		}
	}()
	cqHttpErr, err := cmd.StderrPipe()
	go func() {
		reader := bufio.NewReader(cqHttpErr)
		for {
			readString, err := reader.ReadString('\n')
			if err != nil || err == io.EOF {
				pterm.Error.Println("CQHTTP 出现错误: " + readString)
				pterm.Error.Println("CQHTTP 已退出")
				return
			}
			pterm.Error.Print("CQHTTP 出现错误: " + readString)
		}
	}()
	go func() {
		err = cmd.Start()
		if err != nil {
			fmt.Println(err)
		}
		err = cmd.Wait()
		if err != nil {
			fmt.Println(err)
		}
	}()
	WaitConnect(portNumber)
	connected = true
	pterm.Info.Println("CQ-Http 已成功登录 QQ")
	pterm.Success.Println("CQ-Http已经成功启动了！")
	PackCQHttpRunAuth(isImport)
}

func LoadCurrentFBToken() string {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	fbconfigdir := filepath.Join(homedir, ".config/fastbuilder")
	token := filepath.Join(fbconfigdir, "fbtoken")
	if utils.IsFile(token) {
		if data, err := utils.GetFileData(token); err == nil {
			return string(data)
		}
	}
	return ""
}

func RequestToken() string {
	currentFbToken := LoadCurrentFBToken()
	if currentFbToken != "" && strings.HasPrefix(currentFbToken, "w9/BeLNV/9") {
		pterm.Info.Printf("要使用现有的FB账户登录吗?  使用现有账户请输入 y ,使用新账户请输入 n: ")
		accept := utils.GetInputYN()
		if accept {
			return currentFbToken
		}
	}
	pterm.Info.Printf("请输入FB账号/或者输入 Token: ")
	Code := utils.GetValidInput()
	if strings.HasPrefix(Code, "w9/BeLNV/9") {
		pterm.Success.Printf("您输入的是 Token, 因此无需输入密码了")
		time.Sleep(time.Second)
		return Code
	}
	pterm.Info.Printf("请输入FB密码: ")
	Passwd := utils.GetValidInput()
	tokenstruct := &map[string]interface{}{
		"encrypt_token": true,
		"username":      Code,
		"password":      Passwd,
	}
	if token, err := json.Marshal(tokenstruct); err != nil {
		panic(err)
	} else {
		return string(token)
	}
}

func FBTokenSetup(cfg *BotConfig) {
	if cfg.Token != "" {
		pterm.Info.Printf("要使用上次的 FB 账号登录吗? 要请输入 y ,需要修改请输入 n: ")
		accept := utils.GetInputYN()
		if accept {
			return
		}
	}
	newToken := RequestToken()
	cfg.Token = newToken
}

func StartOmegaHelper() {
	pterm.Success.Println("开始配置Omega登录")
	botConfig := &BotConfig{}
	reconfigFlag := true
	fbTokenSetted := false
	if err := utils.GetJsonData(path.Join(GetCurrentDir(), "服务器登录配置.json"), botConfig); err == nil && botConfig.Code != "" {
		FBTokenSetup(botConfig)
		fbTokenSetted = true
		pwd := " 密码为空"
		if botConfig.Passwd != "" {
			pwd = " 密码为: " + botConfig.Passwd
		}
		pterm.Info.Println("租赁服账号为: " + botConfig.Code + pwd)
		pterm.Info.Printf("接受这个登录配置请输入 y ,需要修改请输入 n: ")
		accept := utils.GetInputYN()
		if accept {
			reconfigFlag = false
		}
	}
	if !fbTokenSetted {
		FBTokenSetup(botConfig)
	}
	if reconfigFlag {
		pterm.Info.Printf("请输入租赁服账号: ")
		botConfig.Code = utils.GetValidInput()
		pterm.Info.Printf("请输入租赁服密码（没有则留空）: ")
		botConfig.Passwd = utils.GetInput()
	}
	if err := utils.WriteJsonData(path.Join(GetCurrentDir(), "服务器登录配置.json"), botConfig); err != nil {
		pterm.Error.Println("无法记录租赁服配置，不过可能不是什么大问题")
	}
	RunOmega(botConfig)
}

func RunFB() {
	var cmd *exec.Cmd
	args := []string{"--no-update-check"}
	// readC := make(chan string)
	cmd = exec.Command(GetOmegaExecName(), args...)
	cmd.Stdin = os.Stdin
	fb_out, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	go func() {
		reader := bufio.NewReader(fb_out)
		io.Copy(os.Stdout, reader)
	}()
	fb_err, err := cmd.StderrPipe()
	if err != nil {
		panic(err)
	}
	go func() {
		reader := bufio.NewReader(fb_err)
		io.Copy(os.Stderr, reader)
	}()
	cmd.Run()
}

func RunOmega(cfg *BotConfig) {
	// fmt.Println(cfg.Token)
	omegaRunning := false
	doExit := false
	var cmd *exec.Cmd
	args := []string{"-M", "-O", "--plain-token", cfg.Token, "--no-update-check", "-c", cfg.Code}
	if cfg.Passwd != "" {
		args = append(args, "-p")
		args = append(args, cfg.Passwd)
	}
	readC := make(chan string)
	go func() {
		for {
			s := utils.GetInput()
			readC <- s
		}
	}()
	signalchannel := make(chan os.Signal)
	signal.Notify(signalchannel, os.Interrupt)
	signal.Notify(signalchannel, syscall.SIGTERM)
	signal.Notify(signalchannel, syscall.SIGQUIT)
	var omega_in io.WriteCloser
	go func() {
		<-signalchannel
		doExit = true
		if omegaRunning {
			pterm.Warning.Println("正在等待 Omega 完成数据保存工作")
			omega_in.Write([]byte("stop\n"))
		} else {
			os.Exit(0)
		}
	}()
	restartTime := 0
	for {
		startTime := time.Now()
		cmd = exec.Command(GetOmegaExecName(), args...)
		omega_out, err := cmd.StdoutPipe()
		if err != nil {
			panic(err)
		}
		omega_in, err = cmd.StdinPipe()
		if err != nil {
			panic(err)
		}
		omega_error, err := cmd.StderrPipe()
		if err != nil {
			panic(err)
		}
		pterm.Success.Println("如果Omega崩溃了，它会在最长 30 秒后自动重启")
		omegaRunning = true

		go func() {
			io.Copy(os.Stdout, omega_out)
		}()
		go func() {
			io.Copy(os.Stderr, omega_error)
		}()
		cmd.Stdin = os.Stdin

		err = cmd.Start()
		if err != nil {
			fmt.Println(err)
		}
		err = cmd.Wait()
		if err != nil {
			fmt.Println(err)
		}
		if time.Since(startTime) > time.Minute*3 {
			restartTime = 0
		} else {
			restartTime++
		}
		if doExit {
			pterm.Success.Println("上方错误可忽略")
			pterm.Success.Println("保存完毕，程序退出")
			os.Exit(0)
			break
		}
		omegaRunning = false
		var sleepTime time.Duration
		if restartTime == 0 {
			sleepTime = time.Second*30 - time.Since(startTime)
			pterm.Warning.Printfln("Omega将在 %v 秒后自动重启", sleepTime.Seconds())
			time.Sleep(sleepTime)
		} else {
			sleepTime = (1 << restartTime) * 30 * time.Second
			if sleepTime > time.Minute*30 {
				sleepTime = time.Minute * 30
			}
			pterm.Warning.Printfln("程序连续第 %v 次崩溃，Omega将在 %v 秒后自动重启", restartTime, sleepTime.Seconds())
			time.Sleep(sleepTime)
		}
	}
}

func GetCqHttpBinary() []byte {
	compressedData := embed_binary.GetCqHttpBinary()
	var execBytes []byte
	var err error
	if execBytes, err = ioutil.ReadAll(brotli.NewReader(bytes.NewReader(compressedData))); err != nil {
		panic(err)
	}
	return execBytes
}

func GetOmegaExecName() string {
	omega := "fastbuilder"
	if GetPlantform() == embed_binary.WINDOWS_x86_64 {
		omega = "fastbuilder.exe"
	}
	omega = path.Join(GetCurrentDir(), omega)
	p, err := filepath.Abs(omega)
	if err != nil {
		panic(err)
	}
	return p
}

func GetCqHttpExec() string {
	cqhttp := "cqhttp"
	if GetPlantform() == embed_binary.WINDOWS_x86_64 {
		cqhttp = "cqhttp.exe"
	}
	cqhttp = path.Join(GetCurrentDir(), cqhttp)
	p, err := filepath.Abs(cqhttp)
	if err != nil {
		panic(err)
	}
	return p
}

func GetPlantform() string {
	return embed_binary.GetPlantform()
}

func GetRemoteOmegaHash() string {
	url := ""
	switch GetPlantform() {
	case embed_binary.WINDOWS_x86_64:
		url = STOARGE_REPO + "/fastbuilder-windows.hash"
	case embed_binary.Linux_x86_64:
		url = STOARGE_REPO + "/fastbuilder-linux.hash"
	case embed_binary.MACOS_x86_64:
		url = STOARGE_REPO + "/fastbuilder-macos.hash"
	case embed_binary.Android_arm64:
		url = STOARGE_REPO + "/fastbuilder-android.hash"
	default:
		panic("未知平台" + GetPlantform())
	}
	// fmt.Println(url)
	hashBytes := utils.DownloadMicroContent(url)
	return string(hashBytes)
}

func GetFileHash(fname string) string {
	if utils.IsFile(fname) {
		fileData, err := utils.GetFileData(fname)
		if err != nil {
			panic(err)
		}
		return utils.GetBinaryHash(fileData)
	}
	return ""
}

func GetCurrentOmegaHash() string {
	exec := GetOmegaExecName()
	return GetFileHash(exec)
}

func GetCQHttpHash() string {
	exec := GetCqHttpExec()
	return GetFileHash(exec)
}

func GetEmbeddedCQHttpHash() string {
	return utils.GetBinaryHash(GetCqHttpBinary())
}

func DownloadOmega() {
	exec := GetOmegaExecName()
	url := ""
	switch GetPlantform() {
	case embed_binary.WINDOWS_x86_64:
		url = STOARGE_REPO + "/fastbuilder-windows.exe.brotli"
	case embed_binary.Linux_x86_64:
		url = STOARGE_REPO + "/fastbuilder-linux.brotli"
	case embed_binary.MACOS_x86_64:
		url = STOARGE_REPO + "/fastbuilder-macos.brotli"
	case embed_binary.Android_arm64:
		url = STOARGE_REPO + "/fastbuilder-android.brotli"
	default:
		panic("未知平台" + GetPlantform())
	}
	// fmt.Println(url)
	compressedData := utils.DownloadSmallContent(url)
	var execBytes []byte
	var err error
	if execBytes, err = ioutil.ReadAll(brotli.NewReader(bytes.NewReader(compressedData))); err != nil {
		panic(err)
	}
	if err := utils.WriteFileData(exec, execBytes); err != nil {
		panic(err)
	}
}
