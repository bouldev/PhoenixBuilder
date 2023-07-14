package fastbuilder

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"phoenixbuilder/fastbuilder/args"
	"phoenixbuilder/fastbuilder/credentials"
	I18n "phoenixbuilder/fastbuilder/i18n"
	"strings"
	"syscall"

	"golang.org/x/term"
)

func loadFBTokenOrAskFBCredential() (token string, username string, password string) {
	if !args.SpecifiedToken() {
		token = credentials.LoadTokenPath()
		if _, err := os.Stat(token); os.IsNotExist(err) {
			fbusername, err := credentials.GetInputUserName()
			if err != nil {
				panic(err)
			}
			fbuntrim := fmt.Sprintf("%s", strings.TrimSuffix(fbusername, "\n"))
			fbun := strings.TrimRight(fbuntrim, "\r\n")
			fmt.Printf(I18n.T(I18n.EnterPasswordForFBUC))
			fbpassword, err := term.ReadPassword(int(syscall.Stdin))
			fmt.Printf("\n")
			token=""
			username=fbun
			psw_sum:=sha256.Sum256([]byte(fbpassword))
			password=hex.EncodeToString(psw_sum[:])
		} else {
			token, err = credentials.ReadToken(token)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	} else {
		token = args.CustomTokenContent
	}
	return
}
