//go:build !fyne_gui
// +build !fyne_gui

package I18n

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
	"bufio"
	"fmt"
	"os"

	"github.com/pterm/pterm"
)

func Init() {
	config := loadConfigPath()
	if _, err := os.Stat(config); os.IsNotExist(err) {
		SelectLanguage()
	} else {
		content, err := os.ReadFile(config)
		if err != nil {
			fmt.Printf("WARNING: Language config file isn't accessible\n")
			I18nDict = LangDict["en_US"]
			return
		}
		langCode := string(content)
		SelectedLanguage = langCode
	}
	langdict, aru := LangDict[SelectedLanguage]
	is_deprecated := IsDeprecated(SelectedLanguage)
	if is_deprecated {
		pterm.Warning.Printf("The language named `%s` has been removed and will be redirected to `%s`.\nPress Enter to continue.\n", SelectedLanguage, DefaultLanguage)
		langdict = LangDict["en_US"]
		if file, err := os.Create(config); err != nil {
			fmt.Println(T(Lang_Config_ErrOnCreate), err)
			fmt.Println(T(ErrorIgnored))
		} else {
			_, err = file.WriteString(DefaultLanguage)
			if err != nil {
				fmt.Println(T(Lang_Config_ErrOnSave), err)
				fmt.Println(T(ErrorIgnored))
			}
			file.Close()
		}
		bufio.NewReader(os.Stdin).ReadString('\n')
	} else if !aru {
		fmt.Printf("Ordered language doesn't exist.\nPlease reselect one:\n")
		SelectLanguage()
		langdict, aru = LangDict[SelectedLanguage]
		if !aru {
			panic("Language still unexists after reselection")
		}
	}
	I18nDict = langdict
}
