package config

import (
	//"golang.design/x/clipboard"
	"phoenixbuilder_fyne_gui/dedicate/fyne/session"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type SessionConfigWithName struct {
	Name   string
	Config *session.SessionConfig
}

type GUI struct {
	setContent   func(v fyne.CanvasObject)
	getContent   func() fyne.CanvasObject
	content      fyne.CanvasObject
	config       *SessionConfigWithName
	onEditDone   func()
	masterWindow fyne.Window
}

func New(config *SessionConfigWithName, onEditDone func(config *SessionConfigWithName)) *GUI {
	gui := &GUI{
		config: config,
		onEditDone: func() {
			onEditDone(config)
		},
	}
	return gui
}

func (g *GUI) makeForm(config *SessionConfigWithName) fyne.CanvasObject {
	noName := strings.HasPrefix(config.Name, "未命名配置")
	suppressName := false
	bindingConfigName := binding.BindString(&config.Name)
	configNameEntry := widget.NewEntryWithData(bindingConfigName)
	configNameEntry.OnChanged = func(s string) {
		config.Name = s
		if !suppressName {
			noName = strings.HasPrefix(config.Name, "未命名配置")
		} else {
			suppressName = false
		}
	}
	fbUserNameEntry := widget.NewEntryWithData(binding.BindString(&config.Config.FBUserName))
	fbUserNameEntry.OnChanged = func(s string) {
		config.Config.FBUserName = s
		if noName {
			bindingConfigName.Set(config.Config.FBUserName + "@" + config.Config.ServerCode)
			suppressName = true
		}
	}

	fbPasswordEntry := widget.NewEntryWithData(binding.BindString(&config.Config.FBPassword))
	fbPasswordEntry.OnChanged = func(s string) {
		config.Config.FBPassword = s
	}

	fbTokenLabel := widget.NewEntryWithData(binding.BindString(&config.Config.FBToken))
	fbTokenLabel.SetPlaceHolder("在第一次登录时自动计算")
	//fbTokenLabel.Disable()

	fbTokenLabel.Wrapping = fyne.TextTruncate
	fbPasswordEntry.OnSubmitted = func(s string) {
		config.Config.FBToken = ""
		fbTokenLabel.SetPlaceHolder("Token将重新计算")
	}
	fbUserNameEntry.OnSubmitted = func(s string) {
		config.Config.FBToken = ""
		fbTokenLabel.SetPlaceHolder("Token将重新计算")
	}
	serverCodeEntry := widget.NewEntryWithData(binding.BindString(&config.Config.ServerCode))
	serverCodeEntry.OnChanged = func(s string) {
		config.Config.ServerCode = s
		if noName {
			bindingConfigName.Set(config.Config.FBUserName + "@" + config.Config.ServerCode)
			suppressName = true
		}
	}
	serverPasswdEntry := widget.NewEntryWithData(binding.BindString(&config.Config.ServerPasswd))
	serverPasswdEntry.PlaceHolder = "(没有密码就留空)"
	languageSelector := widget.NewRadioGroup([]string{"中文", "English"}, func(lang string) {
		if lang == "中文" {
			config.Config.Lang = "zh_CN"
		} else if lang == "English" {
			config.Config.Lang = "en_US"
		}
	})
	switch config.Config.Lang {
	case "zh_CN":
		languageSelector.SetSelected("中文")
		break
	case "en_US":
		languageSelector.SetSelected("English")
		break
	}
	languageSelector.Horizontal = true
	operatorEntry := widget.NewEntryWithData(binding.BindString(&config.Config.RespondUser))
	operatorEntry.PlaceHolder = "(留空时将自动从FB服务器获取)"

	worldChatEnable := widget.NewCheck("启用", func(b bool) { config.Config.MuteWorldChat = !b })
	worldChatEnable.Checked = !config.Config.MuteWorldChat

	var developerOptions fyne.CanvasObject
	if !config.Config.IsDeveloper() {
		developerOptions = widget.NewLabel("请在源码中启用")
	} else {
		developerOptions = container.NewVBox(
			widget.NewCheckWithData("NoPyRPC", binding.BindBool(&config.Config.NoPyRPC)),
			widget.NewCheckWithData("NBTConstructorEnabled", binding.BindBool(&config.Config.NBTConstructorEnabled)),
			container.NewGridWithColumns(2, widget.NewLabel("FBVersion"), widget.NewEntryWithData(binding.BindString(&config.Config.FBVersion))),
			container.NewGridWithColumns(2, widget.NewLabel("FBHash"), widget.NewEntryWithData(binding.BindString(&config.Config.FBHash))),
			container.NewGridWithColumns(2, widget.NewLabel("FBCodeName"), widget.NewEntryWithData(binding.BindString(&config.Config.FBCodeName))),
		)
	}

	majorContent := widget.NewAccordion(
		&widget.AccordionItem{
			Title: "FastBuilder 账户",
			Detail: container.NewVBox(
				container.NewGridWithColumns(2, widget.NewLabel("账号:"), fbUserNameEntry),
				container.NewGridWithColumns(2, widget.NewLabel("密码:"), fbPasswordEntry),
				container.NewGridWithColumns(2,
					container.NewHBox(
						widget.NewLabel("Token(不用填)"),
						&widget.Button{
							Text: "",
							Icon: theme.ContentCopyIcon(),
							OnTapped: func() {
								//glfw.SetClipboardString(fbTokenLabel.Text)
								//fyne.Clipboard()
								//clipboard.Write(clipboard.FmtText, []byte(fbTokenLabel.Text))
								g.masterWindow.Clipboard().SetContent(fbTokenLabel.Text)
							},
							IconPlacement: widget.ButtonIconLeadingText,
							Importance:    widget.LowImportance,
						},
					), fbTokenLabel),
			),
			Open: true,
		},
		&widget.AccordionItem{
			Title: "网易租赁服",
			Detail: container.NewVBox(
				container.NewGridWithColumns(2, widget.NewLabel("租赁服号:"), serverCodeEntry),
				container.NewGridWithColumns(2, widget.NewLabel("租赁服密码:"), serverPasswdEntry),
			),
			Open: true,
		},
		&widget.AccordionItem{
			Title: "其他选项",
			Detail: container.NewVBox(
				container.NewGridWithColumns(2, widget.NewLabel("语言:"), languageSelector),
				container.NewGridWithColumns(2, widget.NewLabel("操作员:"), operatorEntry),
				container.NewGridWithColumns(2, widget.NewLabel("世界聊天:"), worldChatEnable),
			),
			Open: false,
		},
		&widget.AccordionItem{
			Title:  "开发者选项",
			Detail: developerOptions,
			Open:   false,
		},
	)
	majorContent.MultiOpen = true
	return container.NewVBox(
		container.NewGridWithColumns(2, widget.NewLabel("配置名称:"), configNameEntry),
		majorContent,
	)
}

func (g *GUI) GetContent(masterWindow fyne.Window, setContent func(v fyne.CanvasObject), getContent func() fyne.CanvasObject) fyne.CanvasObject {
	g.setContent = setContent
	g.getContent = getContent
	g.masterWindow = masterWindow
	fallbackContent := getContent()
	doneBtns := container.NewVBox(&widget.Button{
		Text: "取消",
		OnTapped: func() {
			setContent(fallbackContent)
		},
		Icon:          theme.CancelIcon(),
		IconPlacement: widget.ButtonIconLeadingText,
	}, &widget.Button{
		Text: "完成",
		OnTapped: func() {
			g.onEditDone()
			setContent(fallbackContent)
		},
		Icon:          theme.ConfirmIcon(),
		IconPlacement: widget.ButtonIconLeadingText,
	})
	g.content = container.NewBorder(
		nil, doneBtns, nil, nil,
		g.makeForm(g.config),
	)
	return g.content
}
