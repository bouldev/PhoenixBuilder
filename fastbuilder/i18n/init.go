// +build !fyne_gui

package I18n

func Init() {
	config:=loadConfigPath()
	if _, err:=os.Stat(config); os.IsNotExist(err) {
		SelectLanguage()
	}else{
		content, err:=ioutil.ReadFile(config)
		if (err != nil) {
			panic("Language config file isn't accessible")
			return
		}
		langCode:=string(content)
		SelectedLanguage=langCode
	}
	langdict, aru := LangDict[SelectedLanguage]
	if(!aru) {
		fmt.Printf("Ordered language doesn't exist.\nPlease reselect one:\n")
		SelectLanguage()
		langdict, aru=LangDict[SelectedLanguage]
		if !aru {
			panic("Language still unexists after reselection")
			return
		}
	}
	I18nDict=langdict
}