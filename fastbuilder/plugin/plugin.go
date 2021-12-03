package plugin

import (
	"fmt"
	"path/filepath"
	"os"
	"io/ioutil"
	"phoenixbuilder/minecraft"
	"plugin"
)

func StartPluginSystem(conn *minecraft.Conn) {
	plugins:=loadConfigPath()
	files, _ := ioutil.ReadDir(plugins)
	for _, file := range files {
		path:=fmt.Sprintf("%s/%s",plugins,file.Name())
		if filepath.Ext(path)!=".so" {
			continue
		}
		go func() {
			RunPlugin(conn,path)
		} ()
	}
}

func RunPlugin(conn *minecraft.Conn,path string) {
	plugin, err := plugin.Open(path)
	if err != nil {
		fmt.Printf("Failed to load plugin: %s\n%v\n",path,err)
		return
	}
	mainfunc, err := plugin.Lookup("plugin_main")
	if err != nil {
		fmt.Printf("Failed to find entry point for plugin %s.",path)
		return
	}
	name:=mainfunc.(func()string)()
	fmt.Printf("Plugin %s(%s) loaded!",name,path)
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