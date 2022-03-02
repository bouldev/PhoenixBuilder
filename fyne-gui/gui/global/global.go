package global

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)


var Banner *fyne.Container




//func MakeBanner(build string) *fyne.Container {
//	if Banner != nil {
//		return Banner
//	}
//	// TODO: Move those buttons to an individual page and leave only 1 button
//	//       here.
//	allBtns:=[]fyne.CanvasObject{ThemeToggleBtn.Btn,ReadmeBtn,InformBtn}
//	if DebugBtn!=nil{
//		allBtns=append(allBtns, DebugBtn)
//	}
//	Banner = container.NewBorder(nil, &widget.Separator{},
//		widget.NewLabel("PhoenixBuilder "+build),
//		container.NewGridWithColumns(len(allBtns),allBtns...),
//		widget.NewLabel(""),
//	)
//	return Banner
//}


func MakeBannerAndSettings(build string,app fyne.App,topWindow fyne.Window, setContent func(v fyne.CanvasObject), getContent func() fyne.CanvasObject)*fyne.Container{
	if Banner != nil {
		return Banner
	}
	var SettingsPage fyne.CanvasObject
	SettingsPage=MakeSettingsPage(app,topWindow,setContent,getContent)
	SettingsBtn:=&widget.Button{
		Icon:              theme.SettingsIcon(),
		Importance:        widget.LowImportance,
	}
	SettingsBtn.OnTapped=func() {
		SettingsBtn.Hide()
		origContent:=getContent()
		setContent(container.NewBorder(nil,
			&widget.Button{
				Text:              "关闭",
				OnTapped: func() {
					setContent(origContent)
					SettingsBtn.Show()
				},
			},nil,nil,SettingsPage))
	}
	Banner = container.NewBorder(nil, &widget.Separator{},
		widget.NewLabel("PhoenixBuilder "+build),
		SettingsBtn,
		widget.NewLabel(""),
	)
	return Banner
}
