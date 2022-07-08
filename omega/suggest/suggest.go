package suggest

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/pterm/pterm"
)

func getSuggest(err string) string {
	fmt.Println(err)
	if strings.Contains(err, ".kick") {
		return "机器人被踢出去了"
	}
	if strings.Contains(err, "loggedinOtherLocation") {
		return "该错误一般由于在别处已经用了同一个fb账号登陆了同一个租赁服，请关闭那个fb或者omega"
	}
	if strings.Contains(err, "server outdated") {
		return "机器人版本较新，但是租赁服版本较低，请升级租赁服"
	}
	if strings.Contains(err, "->42") {
		return "该错误一般由于机器人无法进入租赁服导致，请检查租赁服是否关闭或崩溃"
	}
	if strings.Contains(err, "api.fastbuilder.pro") {
		return "该错误一般由于机器人连接fb服务器，请重试或者更换网络"
	}
	if m, _ := regexp.MatchString(`配置`, err); m {
		return "该错误一般由于Omega配置文件错误导致，请按报错信息检查配置文件"
	}
	return ""
}

func GetOmegaErrorSuggest(err string) string {
	suggest := getSuggest(err)
	if suggest == "" {
		return ""
	} else {
		return pterm.Warning.Sprintln("来自 Omega 系统的建议: " + suggest)
	}
}
