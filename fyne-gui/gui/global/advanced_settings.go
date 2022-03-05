package global

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"phoenixbuilder/fastbuilder/args"
	"os"
)


func MakeAdvancedSettingsPage(app fyne.App,topWindow fyne.Window, setContent func(v fyne.CanvasObject), getContent func() fyne.CanvasObject) fyne.CanvasObject {
	debugModeCheck:=widget.NewCheck("", args.Set_args_isDebugMode)
	debugModeCheck.Checked=args.DebugMode()
	debugModeCheckLine:=container.NewBorder(nil,nil,widget.NewLabel("调试模式"),debugModeCheck)
	authserverInput:=widget.NewEntry()
	authserverInput.Text=args.AuthServer()
	authserverInput.OnSubmitted=args.Do_replace_authserver
	disableHashCheck_Check:=widget.NewCheck("", args.Set_disableHashCheck)
	disableHashCheck_Check.Checked=args.ShouldDisableHashCheck()
	muteWorldChatCheck:=widget.NewCheck("", args.Set_muteWorldChat)
	muteWorldChatCheck.Checked=args.ShouldMuteWorldChat()
	noPyRpcCheck:=widget.NewCheck("", func(val bool) {})
	noPyRpcCheck.Checked=args.NoPyRpc()
	noPyRpcCheck.OnChanged=func(val bool) {
		if(!val) {
			args.Set_noPyRpc(false)
			return
		}
		dialog.ShowConfirm("警告","禁用PyRpc包将导致程序无法在服务器中进行任何交互！",func(v bool){
			if(!v){
				noPyRpcCheck.Checked=false
				return
			}
			args.Set_noPyRpc(val)
		},topWindow)
	}
	
	return container.NewVScroll(container.NewVBox(
		widget.NewLabel("所有设置将在程序退出后重置。"),
		debugModeCheckLine,widget.NewSeparator(),
		container.NewBorder(nil,nil,widget.NewLabel("验证服务器"),nil,authserverInput),
		widget.NewSeparator(),
		container.NewBorder(nil,nil,widget.NewLabel("禁用版本验证"),disableHashCheck_Check),
		widget.NewSeparator(),
		container.NewBorder(nil,nil,widget.NewLabel("不监听世界聊天"),muteWorldChatCheck),
		widget.NewSeparator(),
		container.NewBorder(nil,nil,widget.NewLabel("禁用PyRpc包"),noPyRpcCheck),
		widget.NewSeparator(),
		widget.NewButton("os.Exit(0)",func(){os.Exit(0)}),
		widget.NewSeparator(),
	))
}