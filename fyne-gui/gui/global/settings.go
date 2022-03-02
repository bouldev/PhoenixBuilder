package global

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"net/url"
	"phoenixbuilder/fastbuilder/args"
	my_theme "phoenixbuilder_fyne_gui/gui/theme"
)

type ThemeToggler struct {
	app      fyne.App
	appTheme *my_theme.MyTheme
	Btn      *widget.Button
}

func (tt *ThemeToggler) DataChanged() {
	iv, _ := tt.appTheme.IsLight.Get()
	if iv {
		tt.Btn.Icon=theme.RadioButtonIcon()
		tt.Btn.Text="切换为深色主题"
	}else{
		tt.Btn.Icon=theme.RadioButtonCheckedIcon()
		tt.Btn.Text="切换为浅色主题"
	}
}

func MakeThemeToggleBtn(app fyne.App, appTheme *my_theme.MyTheme) *ThemeToggler {
	t := &ThemeToggler{appTheme: appTheme, app: app}
	toggleBtn := &widget.Button{
		DisableableWidget: widget.DisableableWidget{},
		Text:              "",
		Icon:              nil,
		Importance:        widget.LowImportance,
		Alignment:         widget.ButtonAlignLeading,
		IconPlacement:     widget.ButtonIconLeadingText,
		OnTapped: func() {
			iv, _ := t.appTheme.IsLight.Get()
			if iv {
				t.appTheme.SetDark()
				app.Settings().SetTheme(t.appTheme)
			} else {
				t.appTheme.SetLight()
				app.Settings().SetTheme(t.appTheme)
			}
		},
	}
	iv, _ := appTheme.IsLight.Get()
	if iv {
		toggleBtn.Icon=theme.RadioButtonIcon()
		toggleBtn.Text="切换为深色主题"
	}else{
		toggleBtn.Icon=theme.RadioButtonCheckedIcon()
		toggleBtn.Text="切换为浅色主题"
	}
	appTheme.IsLight.AddListener(t)
	t.Btn = toggleBtn
	return t
}

func MakeReadMePopupButton(win fyne.Window) *widget.Button{
	ReadmeBtn := &widget.Button{
		DisableableWidget: widget.DisableableWidget{},
		Text:              "教程和帮助",
		Icon:              theme.QuestionIcon(),
		Importance:        widget.LowImportance,
		Alignment:         widget.ButtonAlignLeading,
		IconPlacement:     widget.ButtonIconLeadingText,
		OnTapped: func() {
			uclink:=&widget.Hyperlink{
				Text:       "用户中心",
				URL:        &url.URL{Path: "http://uc.fastbuilder.pro/"},
				TextStyle:  fyne.TextStyle{Bold: true},
			}
			uclink.SetURLFromString("http://uc.fastbuilder.pro/")
			downloadLink:=&widget.Hyperlink{
				Text:       "软件下载/更新",
				URL:        &url.URL{Path: "https://storage.fastbuilder.pro/epsilon/"},
				TextStyle:  fyne.TextStyle{Bold: true},
			}
			downloadLink.SetURLFromString("https://storage.fastbuilder.pro/epsilon/")
			readmeLink:=&widget.Hyperlink{
				Text:       "FB使用教程",
				URL:        &url.URL{Path: "https://fastbuilder.pro/phoenix.cn.html"},
				TextStyle:  fyne.TextStyle{Bold: true},
			}
			readmeLink.SetURLFromString("https://fastbuilder.pro/phoenix.cn.html")
			nbtLink:=&widget.Hyperlink{
				Text:       "NBT教程",
				URL:        &url.URL{Path: "https://fastbuilder.pro/nbt.html"},
				TextStyle:  fyne.TextStyle{Bold: true},
			}
			nbtLink.SetURLFromString("https://fastbuilder.pro/nbt.html")
			dialog.ShowCustom("帮助链接","知道了",container.NewVBox(
				uclink,
				readmeLink,
				downloadLink,
				nbtLink,
			), win)
		},
	}
	return ReadmeBtn
}

func MakeInformPopButton(win fyne.Window) *widget.Button {
	InformBtn := &widget.Button{
		DisableableWidget: widget.DisableableWidget{},
		Text:              "关于",
		Icon:              theme.InfoIcon(),
		Importance:        widget.LowImportance,
		Alignment:         widget.ButtonAlignLeading,
		IconPlacement:     widget.ButtonIconLeadingText,
		OnTapped: func() {
			dialog.NewInformation("说明", "项目地址: https://github.com/LNSSPsd/PhoenixBuilder\n贡献者: Ruphane, CAIMEO, CMA2401PT\n\n版本: "+args.GetFBVersion()+"\nCommit hash: "+args.GetFBCommitHash(), win).Show()
		},
	}
	return InformBtn
}

func MakeElementScaleBtn(app fyne.App,topWindow fyne.Window, setContent func(v fyne.CanvasObject), getContent func() fyne.CanvasObject) fyne.CanvasObject {
	// 75 100 125 150 175 200 default 100
	scaleSlider:=widget.NewSlider(75,200)
	scaleSlider.Step=25
	scaleSlider.Value= float64(app.Settings().Theme().(*my_theme.MyTheme).SizeScale)*100
	scaleSlider.OnChanged= func(f float64) {
		t:=app.Settings().Theme().(*my_theme.MyTheme)
		t.SizeScale=float32(f/100)
		app.Settings().SetTheme(t)
	}
	sampleTitle:=widget.NewMultiLineEntry()
	sampleTitle.Text="Task 1: Async\nProgress 50/100 (50%)"
	SettingsPage:=container.NewBorder(
		container.NewBorder(nil,nil,widget.NewLabel("界面元素缩放"),nil, scaleSlider),
		nil,nil,nil,

		widget.NewCard("预览","调整直到找到一个合适的尺寸",
			container.NewVBox(
				sampleTitle,
				container.NewBorder(nil, nil,
					widget.NewButton("X", func() {}),
					widget.NewButton(">", func() {}),
					widget.NewEntry()),
				container.NewGridWithColumns(3,
					widget.NewButton("结束会话", func() {}),
					widget.NewButton("任务及配置", func() {}),
					widget.NewButton("可用命令", func() {}),
				),
			)),
		)
	SettingsBtn:=&widget.Button{
		Text: 	"界面和缩放",
		Icon:              theme.ZoomInIcon(),
		Importance:        widget.LowImportance,
		Alignment:         widget.ButtonAlignLeading,
		IconPlacement:     widget.ButtonIconLeadingText,
		OnTapped: func() {
			origContent:=getContent()
			setContent(container.NewBorder(nil,
				&widget.Button{
					Text:              "好的",
					Icon:              theme.ConfirmIcon(),
					Importance: widget.HighImportance,
					OnTapped: func() {
						setContent(origContent)
					},
				},nil,nil,SettingsPage))
		},
	}
	return SettingsBtn
}

func MakeSettingsPage(app fyne.App,topWindow fyne.Window, setContent func(v fyne.CanvasObject), getContent func() fyne.CanvasObject) fyne.CanvasObject{
	ThemeToggler:=MakeThemeToggleBtn(app,app.Settings().Theme().(*my_theme.MyTheme))
	ScaleBtn:=MakeElementScaleBtn(app,topWindow,setContent, getContent )
	ReadMeBtn:=MakeReadMePopupButton(topWindow)
	InformBtn:=MakeInformPopButton(topWindow)
	return container.NewVScroll(container.NewVBox(
		ThemeToggler.Btn,widget.NewSeparator(),
		ScaleBtn,widget.NewSeparator(),
		ReadMeBtn,widget.NewSeparator(),
		InformBtn,widget.NewSeparator(),
	))
}
