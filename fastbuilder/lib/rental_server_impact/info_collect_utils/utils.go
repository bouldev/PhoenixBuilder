package info_collect_utils

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	I18n "phoenixbuilder/fastbuilder/i18n"
	"strings"
	"syscall"

	"golang.org/x/term"
)

func LoadTokenPath() string {
	homedir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(I18n.T(I18n.Warning_UserHomeDir))
		homedir = "."
	}
	fbconfigdir := filepath.Join(homedir, ".config", "fastbuilder")
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

func GetUserInput(prompt string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	input, err := reader.ReadString('\n')
	return strings.TrimSpace(input), err
}

func GetUserPasswordInput(prompt string) (string, error) {
	fmt.Print(prompt)
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Printf("\n")
	return strings.TrimSpace(string(bytePassword)), err
}

func GetRentalServerCode() (string, string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(I18n.T(I18n.Enter_Rental_Server_Code))
	code, err := reader.ReadString('\n')
	if err != nil {
		return "", "", err
	}
	fmt.Print(I18n.T(I18n.Enter_Rental_Server_Password))
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Printf("\n")
	return strings.TrimRight(code, "\r\n"), strings.TrimSpace(string(bytePassword)), err
}

func WriteFBToken(token string, tokenPath string) {
	if fp, err := os.Create(tokenPath); err != nil {
		fmt.Println(I18n.T(I18n.FBUC_Token_ErrOnCreate), err)
		fmt.Println(I18n.T(I18n.ErrorIgnored))
	} else {
		_, err = fp.WriteString(token)
		if err != nil {
			fmt.Println(I18n.T(I18n.FBUC_Token_ErrOnSave), err)
			fmt.Println(I18n.T(I18n.ErrorIgnored))
		}
		fp.Close()
	}
}

func ReadUserInfo(userName, userPassword, userToken, serverCode, serverPassword string) (string, string, string, string, string, error) {
	var err error
	// read token or get user input
	I18n.Init()
	if userName == "" && userPassword == "" && userToken == "" {
		userToken, err = ReadToken(LoadTokenPath())
		if err != nil || userToken == "" {
			for userName == "" {
				userName, err = GetUserInput("请输入 FB 用户名或者 Token:")
				if strings.HasPrefix(userName, "w9/") {
					userToken = userName
					userName = ""
					break
				}
				if err != nil {
					return userName, userPassword, userToken, serverCode, serverPassword, err
				}
			}
			if userToken == "" {
				for userPassword == "" {
					userPassword, err = GetUserPasswordInput(I18n.T(I18n.EnterPasswordForFBUC))
					if err != nil {
						return userName, userPassword, userToken, serverCode, serverPassword, err
					}
				}
			}
		}
	}

	// read server code and password
	if serverCode == "" {
		serverCode, serverPassword, err = GetRentalServerCode()
		if err != nil {
			return userName, userPassword, userToken, serverCode, serverPassword, err
		}
	}
	return userName, userPassword, userToken, serverCode, serverPassword, nil
}
