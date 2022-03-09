// +build with_v8

package script_kickstarter

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"phoenixbuilder/fastbuilder/script"
	"phoenixbuilder/wayland_v8/host"
	v8 "rogchap.com/v8go"
	"strings"
	"time"
)

func LoadScript(scriptPath string, hb script.HostBridge) (func(),error) {
	iso := v8.NewIsolate()
	global := v8.NewObjectTemplate(iso)
	scriptPath = strings.TrimSpace(scriptPath)
	if scriptPath == "" {
		return nil,fmt.Errorf("Empty script path!")
	}
	fmt.Printf("Loading script: %s\n", scriptPath)
	fmt.Printf("JS engine vesion: %v\n",host.JSVERSION)
	var script string
	var scriptName string
	urlPath, err := url.ParseRequestURI(scriptPath)
	if err==nil{
		scriptName=urlPath.Path
		fmt.Printf("It seems to be a url, try loading it...\n")
		result,err:=obtainPageContent(scriptPath,30*time.Second)
		if err!=nil{
			return nil,err
		}
		script=string(result)
	}else{
		_, scriptName = path.Split(scriptPath)
		file, err := os.OpenFile(scriptPath, os.O_RDONLY, 0755)
		if err != nil {
			return nil,err
		}
		scriptData, err := ioutil.ReadAll(file)
		if err != nil {
			return nil,err
		}
		script=string(scriptData)
	}
	identifyStr:= ""//script.GetStringSha(script)
	stopFunc:=host.InitHostFns(iso,global,hb,scriptName,identifyStr,scriptPath)
	ctx := v8.NewContext(iso, global)
	host.CtxFunctionInject(ctx)
	go func() {
		finalVal, err := ctx.RunScript(script, scriptPath)
		if err != nil {
			fmt.Printf("Script %s ran into a runtime error: %v\n",scriptPath,err.Error())
		}
		fmt.Printf("Script %s completed: %v\n",scriptPath,finalVal)
	}()
	return stopFunc,nil
}

func obtainPageContent(pageUrl string, timeout time.Duration) ([]byte, error) {
	client := &http.Client{Timeout: timeout}
	resp, err := client.Get(pageUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var buffer [512]byte
	result := bytes.NewBuffer(nil)
	for {
		n, err := resp.Body.Read(buffer[0:])
		result.Write(buffer[0:n])
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
	}
	return result.Bytes(), nil
}