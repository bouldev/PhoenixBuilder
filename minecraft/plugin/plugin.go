package plugin

import (
	"fmt"
	"path/filepath"
	"os"
	"io/ioutil"
	"github.com/robertkrimen/otto"
	"phoenixbuilder/minecraft"
	//"phoenixbuilder/minecraft/command"
)

func StartPluginSystem(conn *minecraft.Conn) {
	plugins:=loadConfigPath()
	files, _ := ioutil.ReadDir(plugins)
	for _, file := range files {
		path:=fmt.Sprintf("%s/%s",plugins,file.Name())
		if filepath.Ext(path)!=".js" {
			continue
		}
		go func() {
			RunPlugin(conn,path)
		} ()
	}
}

func DefineGlobalObjects(conn *minecraft.Conn, vm *otto.Otto) {
	functionObject:=otto.Object{}
	functionObject.Set("RegularFunction",60)
	functionObject.Set("SimpleFunction",61)
	functionObject.Set("registerFunction",func (call otto.FunctionCall) otto.Value {
		if !call.Argument(0).IsNumber() {
			return call.Otto.MakeTypeError("func.registerFunction: Argument 0 must be func.RegularFunction or func.SimpleFunction")
		}
		val,_:=call.Argument(0).ToInteger()
		if val!=60 && val!=61 {
			return call.Otto.MakeTypeError("Invalid function type")
		}
		return otto.Value{}
	})
}

func RunPlugin(conn *minecraft.Conn,path string) {
	vm:=otto.New()
	script_content, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Printf("Failed to read script %s :\n%v\n",path,err)
		return
	}
	_, err=vm.Compile(path,script_content)
	if err != nil {
		fmt.Printf("Failed to compile script %s: \n%v\n",path,err)
		return
	}
}

func loadConfigPath() string {
	homedir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("[PLUGIN] WARNING - Failed to obtain the user's home directory. made homedir=\".\";")
		homedir="."
	}
	fbconfigdir := filepath.Join(homedir, ".config/fastbuilder/plugins")
	os.MkdirAll(fbconfigdir, 0755)
	return fbconfigdir
}