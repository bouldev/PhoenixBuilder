package global

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	my_theme "phoenixbuilder_fyne_gui/gui/theme"
	"phoenixbuilder_fyne_gui/gui/assets"
)
var DebugBtn *widget.Button
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