package global

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"phoenixbuilder/fastbuilder/args"
	"encoding/json"
	"os"
	"fmt"
)


func MakeAdvancedSettingsPage(app fyne.App,topWindow fyne.Window, setContent func(v fyne.CanvasObject), getContent func() fyne.CanvasObject) fyne.CanvasObject {
	currentConfig:=args.ParsedArgs
	parsedConfigBS, err:=json.MarshalIndent(currentConfig, "", "\t")
	if(err!=nil) {
		panic(err)
	}
	parsedConfig:=string(parsedConfigBS)
	configurationJSONEntry:=widget.NewMultiLineEntry()
	configurationJSONEntry.Text=parsedConfig
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
		//args.Set_noPyRpc(val)
		return
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
	configurationJSONSubmitButton:=widget.NewButton("PARSE", func() {
		buf:=[]string{}
		err:=json.Unmarshal([]byte(configurationJSONEntry.Text),&buf)
		if err != nil {
			dialog.ShowInformation("错误", fmt.Sprintf("未能粘贴 JSON: %v",err),topWindow)
			return
		}
		args.ParseCustomArgs(buf)
		debugModeCheck.Checked=args.DebugMode()
		authserverInput.Text=args.AuthServer()
		disableHashCheck_Check.Checked=args.ShouldDisableHashCheck()
		muteWorldChatCheck.Checked=args.ShouldMuteWorldChat()
		noPyRpcCheck.Checked=args.NoPyRpc()
		debugModeCheck.Refresh()
		authserverInput.Refresh()
		disableHashCheck_Check.Refresh()
		muteWorldChatCheck.Refresh()
		noPyRpcCheck.Refresh()
	})
	
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
		container.NewBorder(widget.NewLabel("Flags (JSON)"), container.NewVBox(
			configurationJSONEntry,
			configurationJSONSubmitButton,
		), nil, nil), 
		widget.NewButton("os.Exit(0)",func(){os.Exit(0)}),
		widget.NewSeparator(),
	))
}