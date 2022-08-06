package main

import (
	"os"
	"path"
	"phoenixbuilder/omega/utils"
	"strings"

	"github.com/pterm/pterm"
)

func GenZip(srcDir string, zipFile string, discardFn func(filePath string, info os.FileInfo) (discard bool)) {
	os.MkdirAll(path.Dir(zipFile), 0755)
	fp, err := os.OpenFile(zipFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		panic(err)
	}
	wrappedDiscard := func(filePath string, info os.FileInfo) (discard bool) {
		if !discardFn(filePath, info) {
			pterm.Info.Println("PutIn   " + filePath)
			return false
		} else {
			pterm.Warning.Println("Discard " + filePath)
			return true
		}
	}
	utils.Zip(srcDir, fp, wrappedDiscard)
	fp.Close()
	if hashStr, err := utils.GetFileHash(zipFile); err != nil {
		panic(err)
	} else {
		fp, err := os.OpenFile(zipFile+".hash", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
		if err != nil {
			panic(err)
		}
		fp.WriteString(hashStr)
		pterm.Success.Println(zipFile, ": ", hashStr)
		fp.Close()
	}
}
func GenHash(zipFile string) {
	if hashStr, err := utils.GetFileHash(zipFile); err != nil {
		panic(err)
	} else {
		fp, err := os.OpenFile(zipFile+".hash", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
		if err != nil {
			panic(err)
		}
		fp.WriteString(hashStr)
		pterm.Success.Println(zipFile, ": ", hashStr)
		fp.Close()
	}
}

// rsync -avP --delete ./zip_out/* FBOmega:/var/www/omega-storage/omega_compoments_deploy_binary/
// zip -q -r zip_out/linux_amd64.python.zip ../plantform_specific_python/linux_amd64
func main() {
	var outDir = "zip_out"
	var srcDir = "../side"
	// 每次运行前必须部署的文件
	GenZip(srcDir, path.Join(outDir, "basic_structure_and_runtime_libs.zip"), func(filePath string, info os.FileInfo) (discard bool) {
		if strings.Contains(filePath, ".DS_Store") {
			return true
		} else if strings.Contains(filePath, "omega_python_plugins") {
			return true
		} else if strings.Contains(filePath, "dotcs_plugins") {
			return true
		} else if strings.HasSuffix(filePath, "NOTE") {
			return false
		} else if strings.HasSuffix(filePath, "python_plugin_starter.py") || strings.HasSuffix(filePath, "dotcs_emulator.py") || strings.HasSuffix(filePath, "只写了一部分的开发文档.md") {
			return false
		} else if strings.Contains(filePath, "omega_side") {
			return false
		}
		return true
	})
	// Omega 标准 Python 插件的示例文件 （在 omega_python_plugins 文件夹消失时部署）
	GenZip(srcDir, path.Join(outDir, "omega_python_plugins.zip"), func(filePath string, info os.FileInfo) (discard bool) {
		if strings.Contains(filePath, ".DS_Store") {
			return true
		} else if strings.Contains(filePath, "omega_python_plugins") {
			return false
		}
		return true
	})
	// DotCS 插件示例文件 （在 dotcs_plugins 文件夹消失时部署）
	GenZip(srcDir, path.Join(outDir, "dotcs_plugins.zip"), func(filePath string, info os.FileInfo) (discard bool) {
		if strings.Contains(filePath, ".DS_Store") {
			return true
		} else if strings.Contains(filePath, "dotcs_plugins") {
			return false
		}
		return true
	})
	// PlantformSpecific := "../plantform_specific"
	// python 运行环境 conda create python=3.10 -p path--no-default-packages
	// Linux_amd64 python 运行环境
	GenHash(path.Join(outDir, "linux_amd64.python.tar.gz"))
	// MacOS_amd64 python 运行环境
	GenHash(path.Join(outDir, "macos_amd64.python.tar.gz"))
	// // MacOS_arm64 python 运行环境
	GenHash(path.Join(outDir, "macos_arm64.python.tar.gz"))
	// // Windows_amd64 python 运行环境
	GenHash(path.Join(outDir, "windows_amd64.python.tar.gz"))
}
