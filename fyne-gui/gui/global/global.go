package global

import (
	"fmt"
	"net/url"
	"phoenixbuilder_fyne_gui/gui/assets"
	my_theme "phoenixbuilder_fyne_gui/gui/theme"
	"phoenixbuilder/fastbuilder/args"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var ThemeToggleBtn *ThemeToggler
var InformBtn *widget.Button
var ReadmeBtn *widget.Button
var Banner *fyne.Container
var DebugBtn *widget.Button

type ThemeToggler struct {
	app      fyne.App
	appTheme *my_theme.MyTheme
	Btn      *widget.Button
}

func (tt *ThemeToggler) DataChanged() {
	newIcon := theme.RadioButtonCheckedIcon()
	iv, _ := tt.appTheme.IsLight.Get()
	if iv {
		newIcon = theme.RadioButtonIcon()
	}
	tt.Btn.Icon = newIcon
}

func MakeThemeToggleBtn(app fyne.App, appTheme *my_theme.MyTheme) *ThemeToggler {
	if ThemeToggleBtn != nil {
		return ThemeToggleBtn
	}
	t := &ThemeToggler{appTheme: appTheme, app: app}
	initIcon := theme.RadioButtonCheckedIcon()
	iv, _ := appTheme.IsLight.Get()
	if iv {
		initIcon = theme.RadioButtonIcon()
	}
	toggleBtn := &widget.Button{
		DisableableWidget: widget.DisableableWidget{},
		Text:              "",
		Icon:              initIcon,
		Importance:        widget.LowImportance,
		Alignment:         0,
		IconPlacement:     0,
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
	appTheme.IsLight.AddListener(t)
	t.Btn = toggleBtn
	ThemeToggleBtn = t
	return ThemeToggleBtn
}

func MakeReadMePopupButton(win fyne.Window) *widget.Button{
	if ReadmeBtn != nil {
		return InformBtn
	}
	ReadmeBtn = &widget.Button{
		DisableableWidget: widget.DisableableWidget{},
		Text:              "",
		Icon:              theme.QuestionIcon(),
		Importance:        widget.LowImportance,
		Alignment:         0,
		IconPlacement:     0,
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
	if InformBtn != nil {
		return InformBtn
	}
	InformBtn = &widget.Button{
		DisableableWidget: widget.DisableableWidget{},
		Text:              "",
		Icon:              theme.InfoIcon(),
		Importance:        widget.LowImportance,
		Alignment:         0,
		IconPlacement:     0,
		OnTapped: func() {
			dialog.NewInformation("说明", "项目地址: https://github.com/LNSSPsd/PhoenixBuilder\n贡献者: Ruphane, CAIMEO, CMA2401PT\n\n版本: "+args.GetFBVersion()+"\nCommit hash: "+args.GetFBCommitHash(), win).Show()
		},
	}
	return InformBtn
}

func MakeDebugButton(app fyne.App, setContent func(v fyne.CanvasObject), getContent func() fyne.CanvasObject) *widget.Button {
	if DebugBtn != nil {
		return DebugBtn
	}
	debugOutputStr := ""
	debugOutputStrBinding := binding.BindString(&debugOutputStr)
	attachString := func(s string) {
		oldStr, _ := debugOutputStrBinding.Get()
		debugOutputStrBinding.Set(oldStr + s + "\n")
	}
	debugContent := makeDebugContent(app, setContent, getContent, attachString)
	DebugBtn = &widget.Button{
		Text:          "",
		Icon:          theme.WarningIcon(),
		Importance:    widget.LowImportance,
		Alignment:     0,
		IconPlacement: 0,
		OnTapped: func() {
			origContent := getContent()
			closeBtn := &widget.Button{
				Text:       "",
				Icon:       theme.CancelIcon(),
				Importance: widget.MediumImportance,
				OnTapped: func() {
					setContent(origContent)
				},
			}
			setContent(container.NewBorder(debugContent, closeBtn, nil, nil, widget.NewEntryWithData(debugOutputStrBinding)))
		},
	}
	return DebugBtn
}

func makeDebugContent(app fyne.App, setContent func(v fyne.CanvasObject), getContent func() fyne.CanvasObject, attachLine func(string)) fyne.CanvasObject {
	content := container.NewVBox(
		widget.NewLabelWithStyle("Debug", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		container.New(layout.NewGridLayout(3),

			widget.NewButton("Dark", func() {
				app.Settings().Theme().(*my_theme.MyTheme).SetDark()
			}),
			widget.NewButton("Light", func() {
				app.Settings().Theme().(*my_theme.MyTheme).SetLight()
			}),
			widget.NewButton("Chinese", func() {
				//onError := func(info error) {
				//	dialog.ShowError(info, topWindow)
				//	time.Sleep(5 * time.Second)
				//}
				//
				//res, err := utils.LoadFromAssets("Consolas_with_Yahei_Regular.ttf", "Consolas_with_Yahei_Regular.ttf")
				//if err != nil {
				//	onError(err)
				//	return
				//}
				appTheme := app.Settings().Theme().(*my_theme.MyTheme)
				appTheme.Regular = assets.ResourceRegularFont
				appTheme.Italic = assets.ResourceRegularFont
				appTheme.Monospace = assets.ResourceRegularFont
				//res, err = utils.LoadFromAssets("Consolas_with_Yahei_Bold.ttf", "Consolas_with_Yahei_Bold.ttf")
				//if err != nil {
				//	onError(err)
				//	return
				//}
				appTheme.Bold = assets.ResourceBoldFont
				appTheme.BoldItalic = assets.ResourceBoldFont
				//chineseTheme.SetFontsFromAssets("Consolas_with_Yahei_Regular.ttf", "", onError)
				app.Settings().SetTheme(appTheme)
			}),
			widget.NewButton("File", func() {
				dialog.NewFileOpen(func(closer fyne.URIReadCloser, err error) {
					if err != nil {
						attachLine(err.Error())
					} else {
						attachLine("Cannot Open " + closer.URI().String())
					}
				}, app.NewWindow("void")).Show()
			}),
			widget.NewButton("Root", func() {
				attachLine(app.Storage().RootURI().String())
			}),
			widget.NewButton("ListRoot", func() {
				attachLine(fmt.Sprintf("%v", app.Storage().List()))
				//appStorage.List()
			}),
			widget.NewButton("Remove", func() {
				err := app.Storage().Remove("config.yaml")
				if err != nil {
					attachLine("Cannot Remove " + fmt.Sprintf("%v\n%v", app.Storage().List(), err))
				}
			}),
			widget.NewButton("DoSave", func() {
				_, err := app.Storage().Save("config.test.yaml")
				if err != nil {
					attachLine("Cannot Save" + fmt.Sprintf("%v\n%v", app.Storage().List(), err))
				}
			}),
			widget.NewButton("DoCreate", func() {
				_, err := app.Storage().Save("config.test.yaml")
				if err != nil {
					attachLine("Cannot Save" + fmt.Sprintf("%v\n%v", app.Storage().List(), err))
				}
			}),
			widget.NewButton("File&os.Open", func() {
				dialog.NewFileOpen(func(closer fyne.URIReadCloser, err error) {
					if err != nil {
						attachLine(err.Error())
					} else {
						attachLine("Selected(uri) " + closer.URI().String())
						attachLine("Selected(ext) " + closer.URI().Extension())
						p := closer.URI().Path()
						//p = closer.URI().Path()
						cp := p
						//cp = strings.TrimPrefix(cp, "content://")
						//_, err := os.Open(cp)
						closer.Close()
						if err != nil {
							//fyne.Storage.Open()
							attachLine(fmt.Errorf("os.Open error\n%v\n%v", cp, err).Error())
						}
					}
				}, app.NewWindow("void")).Show()
			}),
		),
	)
	return content
}

func MakeBanner(build string) *fyne.Container {
	if Banner != nil {
		return Banner
	}
	// TODO: Move those buttons to an individual page and leave only 1 button
	//       here.
	allBtns:=[]fyne.CanvasObject{ThemeToggleBtn.Btn,ReadmeBtn,InformBtn}
	if DebugBtn!=nil{
		allBtns=append(allBtns, DebugBtn)
	}
	Banner = container.NewBorder(nil, &widget.Separator{},
		widget.NewLabel("PhoenixBuilder"+build),
		container.NewGridWithColumns(len(allBtns),allBtns...),=
		widget.NewLabel(""),
	)
	return Banner
}
