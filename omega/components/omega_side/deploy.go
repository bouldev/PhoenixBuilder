package omega_side

import (
	"fmt"
	"os"
	"path"
	"phoenixbuilder/omega/utils"
	"runtime"

	"github.com/pterm/pterm"
)

// TODO: get repo url from server
var site = "https://omega.fastbuilder.pro/omega_compoments_deploy_binary/"

type deployItemType int

// since the file in the server is specifically designed for this program
// so, it's the file name of webpage follow the definition here but not
// the entry here follow the files in the server
const (
	basicOmegaRuntimeLibsAndDevelopGuide = deployItemType(iota) // 每次运行时都从缓存强制部署
	omegaPythonPluginExamples                                   // omega_python_plugins 文件夹消失时 且启动 omega python 时部署
	dotcsPluginExamples                                         // 与上面的类似
	pythonExecEnv                                               // python 运行环境
	pureDotcsEnv
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
	pureDotcsEnv: {
		utils.PLATFORM_ALL: site + "pure_dotcs.zip",
	},
	pythonExecEnv: {
		utils.PLATFORM_LINUX_AMD64:   site + "linux_amd64.python.tar.gz",
		utils.PLATFORM_MACOS_AMD64:   site + "macos_amd64.python.tar.gz",
		utils.PLATFORM_MACOS_ARM64:   site + "macos_arm64.python.tar.gz",
		utils.PLATFORM_WINDOWS_AMD64: site + "windows_amd64.python.tar.gz",
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

func (o *OmegaSide) deployBasicLibrary() {
	os.MkdirAll(o.getWorkingDir(), 0755)
	targetDeployDir := o.getWorkingDir()
	cacheDir := o.getCacheDir()
	if o.EnableOmegaPythonRuntime || o.EnableDotCSSimulator {
		if url, hasK := getDeployResourceURL(basicOmegaRuntimeLibsAndDevelopGuide); hasK {
			deployer := utils.SimpleDeployer{
				CacheFilePath:      path.Join(cacheDir, "Omega运行相关库.zip"),
				TargetDeployDir:    targetDeployDir,
				SourceFileURL:      url,
				SourceFileMD5ByURL: url + ".hash",
			}
			if err := deployer.Deploy(); err != nil {
				pterm.Error.Println("自动部署失败: " + err.Error())
			}
		} else {
			pterm.Warning.Println("无法在你的设备中自动部署 Omega Side 旁加载插件运行相关库，你将无法使用 Omega 标准 Python 插件和 DotCS 社区版插件")
		}
	}

	if o.EnableOmegaPythonRuntime {
		targetDir := path.Join(targetDeployDir, "omega_python_plugins")
		if !utils.IsDir(targetDir) {
			if url, hasK := getDeployResourceURL(omegaPythonPluginExamples); hasK {
				deployer := utils.SimpleDeployer{
					CacheFilePath:      path.Join(cacheDir, "Omega标准Python插件例子.zip"),
					TargetDeployDir:    targetDeployDir,
					SourceFileURL:      url,
					SourceFileMD5ByURL: url + ".hash",
				}
				if err := deployer.Deploy(); err != nil {
					pterm.Error.Println("自动部署失败: " + err.Error())
				}
				pterm.Success.Printfln("已经自动将Omega标准Python插件例子下载到 %v 中，你也可以将其他使用 Omega 标准 Python插件放入其中", targetDir)
			} else {
				pterm.Warning.Println("无法在你的设备中自动部署Omega标准Python插件例子")
			}
		}

	}

	if o.EnableDotCSSimulator {
		targetDir := path.Join(targetDeployDir, "dotcs_plugins")
		if !utils.IsDir(targetDir) {
			if url, hasK := getDeployResourceURL(dotcsPluginExamples); hasK {
				deployer := utils.SimpleDeployer{
					CacheFilePath:      path.Join(cacheDir, "DotCS社区版插件例子.zip"),
					TargetDeployDir:    targetDeployDir,
					SourceFileURL:      url,
					SourceFileMD5ByURL: url + ".hash",
				}
				if err := deployer.Deploy(); err != nil {
					pterm.Error.Println("自动部署失败: " + err.Error())
				}
				pterm.Success.Printfln("已经自动将DotCS社区版插件例子下载到 %v 中，你也可以将其他DotCS社区版插件放入其中", targetDir)
			} else {
				pterm.Warning.Println("无法在你的设备中自动部署DotCS社区版插件例子")
			}
		}
	}
	if o.EnablePureDotCSEnv {
		targetDir := path.Join(targetDeployDir, "dotcs")
		if url, hasK := getDeployResourceURL(pureDotcsEnv); hasK {
			deployer := utils.SimpleDeployer{
				CacheFilePath:      path.Join(cacheDir, "原生DotCS环境.zip"),
				TargetDeployDir:    targetDir,
				SourceFileURL:      url,
				SourceFileMD5ByURL: url + ".hash",
			}
			if err := deployer.Deploy(); err != nil {
				pterm.Error.Println("自动部署失败: " + err.Error())
			}
			pterm.Success.Printfln("已经自动DotCS社区版环境部署到 %v 中，其应该与DotCS具有相同的结构", targetDir)
		} else {
			pterm.Warning.Println("无法在你的设备中自动部署DotCS社区版环境")
		}
	}
}

func (o *OmegaSide) deployPythonRuntime() {
	pterm.Warning.Printfln("开始自动部署 Python 插件开发和运行环境，根据你的网速，这可能要很久，但是放心，这仅会发生一次")
	if utils.IsFile(path.Join(o.getCacheDir(), "Python解释器.zip")) {
		os.Rename(path.Join(o.getCacheDir(), "Python解释器.zip"), path.Join(o.getCacheDir(), "Python解释器.tar.gz"))
	}
	if url, hasK := getDeployResourceURL(pythonExecEnv); hasK {
		deployer := utils.SimpleDeployer{
			CacheFilePath:      path.Join(o.getCacheDir(), "Python解释器.tar.gz"),
			TargetDeployDir:    o.getWorkingDir(),
			SourceFileURL:      url,
			SourceFileMD5ByURL: url + ".hash",
		}
		if err := deployer.Deploy(); err != nil {
			pterm.Error.Println("自动部署失败: " + err.Error())
		}
		pterm.Success.Printfln("已经自动准备Python解释器")
		o.pythonPath = "interpreters/python/bin/python"
		if runtime.GOOS == "windows" {
			o.pythonPath = "interpreters/python/bin/python.exe"
		}
		os.Chmod(path.Join(o.getWorkingDir(), o.pythonPath), 0755)
	} else {
		pterm.Warning.Println("无法在你的设备中自动准备Python解释器")
	}
}
