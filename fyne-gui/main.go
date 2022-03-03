package main

import (
	"io/fs"
	"phoenixbuilder/fastbuilder/args"
	bridge_write_path "phoenixbuilder/fastbuilder/bdump/path"
	bridge_read_path "phoenixbuilder/fastbuilder/builder/path"
	"phoenixbuilder_fyne_gui/gui/assets"
	"phoenixbuilder_fyne_gui/gui/global"
	"phoenixbuilder_fyne_gui/gui/profiles"
	my_theme "phoenixbuilder_fyne_gui/gui/theme"
	"phoenixbuilder_fyne_gui/platform_helper"

	"fyne.io/fyne/v2/storage"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

var topWindow fyne.Window
var appTheme *my_theme.MyTheme

func main() {
	args.ParseArgs()
	//args.SetShouldDisableHashCheck()
	bridge_write_path.CreateFile = func(p string) (bridge_write_path.FileWriter, error) {
		uri, err := storage.ParseURI(p)
		if err != nil {
			return nil, &fs.PathError{
				Op:   "ParseURI",
				Path: p,
				Err:  err,
			}
		}
		file, err := storage.Writer(uri)
		if err != nil {
			return nil, &fs.PathError{
				Op:   "OpenWriter",
				Path: uri.String(),
				Err:  err,
			}
		}
		return file, nil
	}
	bridge_read_path.ReadFile = func(p string) (bridge_read_path.FileReader, error) {
		uri, err := storage.ParseURI(p)
		if err != nil {
			return nil, &fs.PathError{
				Op:   "ParseURI",
				Path: p,
				Err:  err,
			}
		}
		file, err := storage.Reader(uri)
		if err != nil {
			return nil, &fs.PathError{
				Op:   "ParseURI",
				Path: p,
				Err:  err,
			}
		}
		return file, nil
	}

	app := app.NewWithID("pro.fastbuilder.app")
	appStorage := app.Storage()
	//appStorage.Create("config.yaml")

	platform_helper.DoNetworkRequest()

	appTheme = my_theme.NewTheme()
	setThemeChineseFont(appTheme)
	appTheme.SetLight()
	app.Settings().SetTheme(appTheme)

	topWindow = app.NewWindow("PhoenixBuilder")
	icon := canvas.NewImageFromResource(assets.ResourceIconPng)
	icon.FillMode = canvas.ImageFillContain
	app.SetIcon(icon.Resource)
	topWindow.SetMaster()

	majorContent := container.NewMax()

	getContent := func() fyne.CanvasObject {
		if len(majorContent.Objects) != 0 {
			return majorContent.Objects[0]
		} else {
			return nil
		}
	}
	setContent := func(v fyne.CanvasObject) {
		majorContent.Objects = []fyne.CanvasObject{v}
		majorContent.Refresh()
	}

	global.MakeBannerAndSettings(args.GetFBVersion(),app,topWindow,setContent,getContent)

	//global.MakeThemeToggleBtn(app, appTheme)
	//global.MakeInformPopButton(topWindow)
	// global.MakeDebugButton(app, setContent, getContent)
	//global.MakeReadMePopupButton(topWindow)
	//global.MakeBanner(args.GetFBVersion())

	//vsplit := container.NewVSplit(debugContent, majorContent)
	//vsplit.Offset = 0.05
	content := container.NewBorder(global.Banner, nil, nil, nil, majorContent)
	topWindow.SetContent(content)

	//onPanicFn := func(err error) {
	//	dialog.ShowError(fmt.Errorf("发生了严重错误，程序即将退出：\n\n%v", err), topWindow)
	//	// os.Exit(-1)
	//}

	//stat, err := os.Stat(dataFolder)
	//if !(err == nil && stat.IsDir()) {
	//	err = os.Mkdir(dataFolder, 0755)
	//	if err != nil {
	//		onPanicFn(fmt.Errorf("权限错误，无法创建必要的数据文件夹 %v (%v)", dataFolder, err))
	//	}
	//}

	profilesObject := profiles.New(appStorage)
	setContent(profilesObject.GetContent(setContent, getContent, topWindow, app))

	topWindow.Resize(fyne.NewSize(480, 640))
	topWindow.ShowAndRun()
}

func setThemeChineseFont(t *my_theme.MyTheme) {
	appTheme.Regular = assets.ResourceRegularFont
	appTheme.Italic = assets.ResourceRegularFont
	appTheme.Monospace = assets.ResourceRegularFont
	appTheme.Bold = assets.ResourceBoldFont
	appTheme.BoldItalic = assets.ResourceBoldFont
}
