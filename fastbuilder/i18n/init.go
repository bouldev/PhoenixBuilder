//go:build !fyne_gui
// +build !fyne_gui

package I18n

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
