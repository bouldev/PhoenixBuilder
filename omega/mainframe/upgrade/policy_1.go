package upgrade

import (
	"path"
	"path/filepath"
	"phoenixbuilder/omega/utils"

	"github.com/pterm/pterm"
)

var android_store_path = "/sdcard/Download/omega_storage"

func Policy_1() string {
	if abs1, err := filepath.Abs(android_store_path); err != nil {
		return "omega_storage"
	} else {
		if abs2, err := filepath.Abs("omega_storage"); err != nil || abs1 == abs2 {
			return "omega_storage"
		}
	}
	if utils.IsDir(path.Dir(android_store_path)) && !utils.IsDir(android_store_path) && utils.IsDir("omega_storage") {
		pterm.Warning.Println("您似乎在使用一部安卓手机，我们将尝试将 omega 配置文件移动到 omega_storage 下")
		if err := utils.MakeDirP(android_store_path); err != nil {
			pterm.Error.Println("移动失败，无法创建 " + android_store_path)
			return "omega_storage"
		}
		if err := utils.CopyDirectory("omega_storage", android_store_path); err != nil {
			pterm.Error.Println("移动失败，拷贝到 " + android_store_path + " 时出现错误")
			return "omega_storage"
		} else {
			pterm.Success.Println("成功移动到 " + android_store_path + " 了")
			return android_store_path
		}
	}
	return "omega_storage"
}
