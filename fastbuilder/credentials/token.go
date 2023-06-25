package credentials

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	fbauth "phoenixbuilder/fastbuilder/cv4/auth"
	"phoenixbuilder/fastbuilder/environment"
	I18n "phoenixbuilder/fastbuilder/i18n"
)

func ProcessTokenDefault(env *environment.PBEnvironment) bool {
	token := env.LoginInfo.Token
	client := fbauth.CreateClient(env)
	env.FBAuthClient = client
	if token[0] == '{' {
		token, err_msg := client.GetToken("", token)
		if token == "" {
			fmt.Printf("%s\n", err_msg)
			fmt.Println(I18n.T(I18n.FBUC_LoginFailed))
			return false
		}
		tokenPath := LoadTokenPath()
		if fi, err := os.Create(tokenPath); err != nil {
			fmt.Println(I18n.T(I18n.FBUC_Token_ErrOnCreate), err)
			fmt.Println(I18n.T(I18n.ErrorIgnored))
		} else {
			env.LoginInfo.Token = token
			_, err = fi.WriteString(token)
			if err != nil {
				fmt.Println(I18n.T(I18n.FBUC_Token_ErrOnSave), err)
				fmt.Println(I18n.T(I18n.ErrorIgnored))
			}
			fi.Close()
			fi = nil
		}
	}
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
