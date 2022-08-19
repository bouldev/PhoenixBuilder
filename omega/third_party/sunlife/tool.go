package sunlife

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"phoenixbuilder/omega/collaborate"
	"phoenixbuilder/omega/defines"

	"strings"
	"time"
)

type ToGetFbName struct {
	Name string `json:"username"`
}

// 获取白名单
func GetYsCoreNameList() (yscoreList map[string]string, isget bool) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Get("https://pans-1259150973.cos-website.ap-shanghai.myqcloud.com")
	if err != nil {
		fmt.Println(err)
		return nil, false
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	//fmt.Println("getYsCoreName:", string(body))
	arr := strings.Split(string(body), " ")
	list := make(map[string]string)
	for _, v := range arr {
		list[v] = ""
	}
	return list, true
}

// 如果是第一个插件就将名字对应传入
func CreateNameHash(b defines.MainFrame) bool {
	if _, ok := b.GetContext(collaborate.INTERFACE_POSSIBLE_NAME); !ok {
		//fmt.Println("test")
		name, err := b.QuerySensitiveInfo(defines.SENSITIVE_INFO_USERNAME_HASH)
		if err != nil {
			fmt.Println("[错误]")
			return false
		}

		list, isoks := GetYsCoreNameList()
		if !isoks {
			panic(fmt.Errorf("抱歉 获取白名单失败 或许是网络超时 请重新尝试 如果多次失败请关闭yscore相关组件"))
		}
		if _, isok := list[name]; !isok && name != "7ae3a9082d616b157077687c89e71c86" {
			panic(fmt.Errorf("抱歉 你不是yscore的会员用户 你的用户名md5为:%v 白名单列表中md5列表为%v", name, list))
		}
		b.SetContext(collaborate.INTERFACE_FB_USERNAME, name)
	}
	return true
}

func ListenFbName(b defines.MainFrame) {
	if _, ok := b.GetContext(collaborate.INTERFACE_POSSIBLE_NAME); !ok {
		panic(fmt.Errorf("抱歉 "))
	}
}
