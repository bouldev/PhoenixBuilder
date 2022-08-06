package omega_side

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	"phoenixbuilder/omega/utils"
	"runtime"
	"strings"

	"github.com/pterm/pterm"
	"golang.org/x/text/encoding/simplifiedchinese"
)

type OmegaSideProcessStartCmd struct {
	Name     string            `json:"旁加载功能名"`
	Cmd      string            `json:"启动指令"`
	Remapper map[string]string `json:"变更选项"`
}

type OmegaSide struct {
	*defines.BasicComponent
	PreferPort               string   `json:"如果可以则使用这个http端口"`
	DebugServerOnly          bool     `json:"只打开用于开发的Websocket端口而不启动任何插件"`
	EnableOmegaPythonRuntime bool     `json:"使用Omega标准Python插件框架"`
	EnableDotCSSimulator     bool     `json:"使用DotCS社区版插件运行模拟器"`
	PossiblePythonExecPath   []string `json:"python解释器搜索路径"`
	autoDeployPython         bool
	pythonPath               string
	StartUpCmds              []OmegaSideProcessStartCmd `json:"启动其他旁加载程序的指令"`
	closeCtx                 chan struct{}
	pushController           *pushController
	fileChange               bool
	FileName                 string `json:"玩家数据文件"`
	PlayerData               map[string]map[string]interface{}
}

func (o *OmegaSide) WaitClose() {
	<-o.closeCtx
}

func (o *OmegaSide) getWorkingDir() string {
	return o.Frame.GetOmegaSideDir()
}
func (o *OmegaSide) getCacheDir() string {
	return path.Join(o.Frame.GetOmegaSideDir(), "cache")
}

func (o *OmegaSide) OnMCPkt(pktID int, data interface{}) {
	o.pushController.pushMCPkt(pktID, data)
}

func (o *OmegaSide) runCmd(subProcessName string, cmdStr string, remapping map[string]string, execDir string) {
	for k, v := range remapping {
		cmdStr = strings.ReplaceAll(cmdStr, k, v)
	}

	cmds := strings.Split(cmdStr, " ")
	execName := ""
	args := []string{}
	i := 0
	for _, frag := range cmds {
		if frag == "" {
			continue
		}
		i++
		if i == 1 {
			execName = frag
		} else {
			args = append(args, frag)
		}
	}
	if execName == "" {
		pterm.Info.Println("启动子进程[" + subProcessName + "]: " + cmdStr + " 失败: 未指定 程序名")
		return
	} else {
		pterm.Info.Println("启动子进程["+subProcessName+"]: "+cmdStr+" => 标准化为", strings.Join([]string{pterm.Yellow(execName), pterm.Blue(strings.Join(args, " "))}, " "))
	}
	cmd := exec.Command(execName, args...)
	if !path.IsAbs(execDir) {
		wd, _ := os.Getwd()
		execDir = path.Join(wd, execDir)
	}
	if runtime.GOOS == "windows" {
		execDir = strings.ReplaceAll(execDir, "\\", "/")
	}
	cmd.Dir = execDir
	// cmd.Env = append(cmd.Env,
	// 	"PATH="+execDir,
	// )
	pterm.Info.Println("工作目录 " + execDir)
	cmdOut, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	Error := pterm.PrefixPrinter{
		MessageStyle: &pterm.ThemeDefault.ErrorMessageStyle,
		Prefix: pterm.Prefix{
			Style: &pterm.ThemeDefault.ErrorPrefixStyle,
			Text:  fmt.Sprintf("%v错误", subProcessName),
		},
	}
	go func() {
		reader := bufio.NewReader(cmdOut)
		replacerRule := utils.GenerateMCColorReplacerRule()
		replacerRule = append(replacerRule, "\n", "\x1b[m\n")
		replacer := strings.NewReplacer(replacerRule...)
		for {
			readString, err := reader.ReadString('\n')
			if err != nil || err == io.EOF {
				// Info.Println("已退出")
				return
			}
			readString = strings.Trim(readString, "\n")
			if runtime.GOOS == "windows" {
				// gbk, _ := simplifiedchinese.GBK.NewEncoder().Bytes([]byte(readString))
				readString, err = simplifiedchinese.GBK.NewDecoder().String(readString)
				if err != nil {
					pterm.Warning.Println(err)
				}
			}
			if readString == "" {
				continue
			}
			fmt.Println(replacer.Replace(readString))
		}
	}()
	cmdErr, err := cmd.StderrPipe()
	go func() {
		reader := bufio.NewReader(cmdErr)
		for {
			readString, err := reader.ReadString('\n')
			if err != nil || err == io.EOF {
				Error.Println("已退出")
				return
			}
			if runtime.GOOS == "windows" {
				// gbk, _ := simplifiedchinese.GBK.NewEncoder().Bytes([]byte(readString))
				readString, err = simplifiedchinese.GBK.NewDecoder().String(readString)
				if err != nil {
					pterm.Warning.Println(err)
				}
			}
			readString = strings.Trim(readString, "\n")
			if readString == "" {
				continue
			}
			Error.Println(readString)
		}
	}()
	go func() {
		err = cmd.Start()
		if err != nil {
			Error.Println(err)
		}
		err = cmd.Wait()
		if err != nil {
			Error.Println(err)
		}
	}()
}

func (o *OmegaSide) Init(cfg *defines.ComponentConfig) {
	m, _ := json.Marshal(cfg.Configs)
	if err := json.Unmarshal(m, o); err != nil {
		panic(err)
	}
}

func (o *OmegaSide) Inject(frame defines.MainFrame) {
	o.Frame = frame
	o.PlayerData = map[string]map[string]interface{}{}
	err := frame.GetJsonData(o.FileName, &o.PlayerData)
	if err != nil {
		panic(err)
	}
}

func (o *OmegaSide) Activate() {
	o.deployBasicLibrary()
	if o.EnableOmegaPythonRuntime || o.EnableDotCSSimulator {
		o.autoDeployPython = true
	}
	if o.EnableOmegaPythonRuntime {
		o.StartUpCmds = append(o.StartUpCmds, OmegaSideProcessStartCmd{
			Name:     "Python",
			Cmd:      "[python] python_plugin_starter.py --server ws://[addr]/omega_side",
			Remapper: map[string]string{},
		})
	}
	if o.EnableDotCSSimulator {
		o.StartUpCmds = append(o.StartUpCmds, OmegaSideProcessStartCmd{
			Name:     "DotCS",
			Cmd:      "[python] dotcs_emulator.py --server ws://[addr]/omega_side",
			Remapper: map[string]string{},
		})
	}
	if o.autoDeployPython {
		needDeployPython := true
		o.PossiblePythonExecPath = append(o.PossiblePythonExecPath, "interpreters/python/bin/python", "interpreters/python/bin/python.exe")
		for _, possiblePath := range o.PossiblePythonExecPath {
			if !path.IsAbs(possiblePath) {
				if _, err := os.Stat(path.Join(o.getWorkingDir(), possiblePath)); err == nil {
					needDeployPython = false
					o.pythonPath = possiblePath
					break
				}
			} else {
				if _, err := os.Stat(possiblePath); err == nil {
					needDeployPython = false
					o.pythonPath = possiblePath
					break
				}
			}
		}
		if needDeployPython {
			o.deployPythonRuntime()
		}
	}
	o.SideUp()
	o.Frame.GetGameListener().SetOnAnyPacketCallBack(func(p packet.Packet) {
		o.pushController.pushMCPkt(int(p.ID()), p)
	})
}
func (o *OmegaSide) Signal(signal int) error {
	switch signal {
	case defines.SIGNAL_DATA_CHECKPOINT:
		if o.fileChange {
			o.fileChange = false
			return o.Frame.WriteJsonDataWithTMP(o.FileName, ".ckpt", o.PlayerData)
		}
	}
	return nil
}

func (o *OmegaSide) Stop() error {
	fmt.Printf("正在保存 %v\n", o.FileName)
	return o.Frame.WriteJsonDataWithTMP(o.FileName, ".final", o.PlayerData)
}
