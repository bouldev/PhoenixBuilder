package omega_side

import (
	"fmt"
	"phoenixbuilder/omega/utils"

	"github.com/pterm/pterm"
)

var site = "localhost:8083/"

type deployItemType int

// since the file in the server is specifically designed for this program
// so, it's the file name of webpage follow the definition here but not
// the entry here follow the files in the server
const (
	basicOmegaRuntimeLibsAndDevelopGuide = deployItemType(iota) // 每次运行时都从缓存强制部署
	omegaPythonPluginExamples                                   // omega_python_plugins 文件夹消失时 且启动 omega python 时部署
	dotcsPluginExamples                                         // 与上面的类似
	pythonExecEnv                                               // python 运行环境
)

// entry name, platform code
var deployList = map[deployItemType]map[utils.PREPARED_PLATFORM_MARK]string{
	basicOmegaRuntimeLibsAndDevelopGuide: {
		utils.PLATFORM_ALL: site + "basic_structure_and_runtime_libs.zip",
	},
	omegaPythonPluginExamples: {
		utils.PLATFORM_ALL: site + "omega_python_plugins.zip",
	},
	dotcsPluginExamples: {
		utils.PLATFORM_ALL: site + "dotcs_plugins.zip",
	},
	pythonExecEnv: {
		utils.PLATFORM_LINUX_AMD64:   site + "linux_amd64.python.zip",
		utils.PLATFORM_MACOS_AMD64:   site + "macos_amd64.python.zip",
		utils.PLATFORM_MACOS_ARM64:   site + "macos_arm64.python.zip",
		utils.PLATFORM_WINDOWS_AMD64: site + "windows_amd64.python.zip",
	},
}

func getDeployResourceURL(item deployItemType) (url string, found bool) {
	platformSpecificURL := deployList[item]
	if url, hasK := platformSpecificURL[utils.PLATFORM_ALL]; hasK {
		return url, true
	} else if url, hasK := platformSpecificURL[utils.PLATFORM_MARK_FOR_PREPARED]; hasK {
		return url, true
	}
	if item == pythonExecEnv && utils.PLATFORM_MARK_FOR_PREPARED == utils.PLATFORM_ANDROID_ARM64 {
		pterm.Warning.Println("很遗憾，受限于Termux的运行机制，我们无法自动为你准备 Python， 但是你仍有办法自行安装 Python，现在，请仔细阅读以下说明")
		fmt.Println("请按顺序，在退出 omega 和 fb 后，输入以下三段代码")
		fmt.Println()
		fmt.Println(`sed -i 's@^\(deb.*stable main\)$@#\1\ndeb https://mirrors.tuna.tsinghua.edu.cn/termux/apt/termux-main stable main@' $PREFIX/etc/apt/sources.list`)
		fmt.Println()
		fmt.Println(`apt update && apt upgrade`)
		fmt.Println()
		fmt.Println(`pkg install python`)
		fmt.Println()
		pterm.Warning.Println("很遗憾，受限于Termux的运行机制，我们无法自动为你准备 Python， 但是你仍有办法自行安装 Python，现在，请仔细阅读以上说明")
		panic("请阅读上述安卓 Python 安装说明，并仔细按说明操作,祝你好运！")
	}
	return "", false
}
