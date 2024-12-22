package core

/*
 * This file is part of PhoenixBuilder.

 * PhoenixBuilder is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License.

 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.

 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.

 * Copyright (C) 2021-2025 Bouldev
 */

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"phoenixbuilder/fastbuilder/args"
	I18n "phoenixbuilder/fastbuilder/i18n"
	"phoenixbuilder/fastbuilder/utils"
	"strings"
	"syscall"

	"golang.org/x/term"
)

func loadTokenOrAskForCredential() (token string, username string, password string) {
	if !args.SpecifiedToken() {
		token = utils.LoadTokenPath()
		if _, err := os.Stat(token); os.IsNotExist(err) {
			fbusername, err := utils.GetUsernameInput()
			if err != nil {
				panic(err)
			}
			fbuntrim := fmt.Sprintf("%s", strings.TrimSuffix(fbusername, "\n"))
			fbun := strings.TrimRight(fbuntrim, "\r\n")
			fmt.Printf(I18n.T(I18n.EnterPasswordForFBUC))
			fbpassword, err := term.ReadPassword(int(syscall.Stdin))
			fmt.Printf("\n")
			token = ""
			username = fbun
			psw_sum := sha256.Sum256([]byte(fbpassword))
			password = hex.EncodeToString(psw_sum[:])
		} else {
			token, err = utils.ReadToken(token)
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
