package credentials

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"phoenixbuilder/fastbuilder/environment"
	I18n "phoenixbuilder/fastbuilder/i18n"
	fbauth "phoenixbuilder/fastbuilder/pv4"
)

func ProcessTokenDefault(env *environment.PBEnvironment) bool {
	client := fbauth.CreateClient(env.ClientOptions)
	env.FBAuthClient = client
	return true
}

func LoadTokenPath() string {
	homedir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(I18n.T(I18n.Warning_UserHomeDir))
		homedir = "."
	}
	fbconfigdir := filepath.Join(homedir, ".config/fastbuilder")
	os.MkdirAll(fbconfigdir, 0700)
	token := filepath.Join(fbconfigdir, "fbtoken")
	return token
}

func ReadToken(path string) (string, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
