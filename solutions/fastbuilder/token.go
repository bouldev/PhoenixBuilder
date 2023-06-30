package fastbuilder

import (
	"encoding/json"
	"fmt"
	"os"
	"phoenixbuilder/fastbuilder/args"
	"phoenixbuilder/fastbuilder/credentials"
	I18n "phoenixbuilder/fastbuilder/i18n"
	"strings"
	"syscall"

	"golang.org/x/term"
)

func loadFBTokenOrAskFBCredential() (token string) {
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
			tokenstruct := &map[string]interface{}{
				"encrypt_token": true,
				"username":      fbun,
				"password":      string(fbpassword),
			}
			bytes_token, err := json.Marshal(tokenstruct)
			if err != nil {
				fmt.Println(I18n.T(I18n.FBUC_Token_ErrOnGen))
				fmt.Println(err)
				return
			}
			token = string(bytes_token)
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
