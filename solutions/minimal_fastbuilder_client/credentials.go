package main

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

func WriteToken(token string, tokenPath string) {
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
