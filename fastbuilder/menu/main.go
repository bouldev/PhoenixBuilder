package menu

import (
	"time"
	"fmt"
	"phoenixbuilder/minecraft"
	fbtask "phoenixbuilder/fastbuilder/task"
	"phoenixbuilder/minecraft/protocol/packet"
	"github.com/google/uuid"
	"io/ioutil"
	"strings"
	"runtime"
	"phoenixbuilder/fastbuilder/i18n"
	"phoenixbuilder/fastbuilder/environment"
)

var OPRuntimeId uint64 = 0
var OPRuntimeIdReceivedChannel chan bool = nil
var LastOPPitch float32 = 0
var LastOPSneak bool = false
var LastOPMouSneaked bool = false
var SelectableCommands = []string {
	I18n.T(I18n.Menu_GetPos),
	I18n.T(I18n.Menu_GetEndPos),
	"ACME",
	"bdump",
	"SCH-",
	"PL0T",
	"EXPORT",
	I18n.T(I18n.Menu_Quit),
	I18n.T(I18n.Menu_Cancel),
}
var curSelArr []string = SelectableCommands
var levelArr []string
var curpath []string
var windowsDrives []string=[]string{
"A:/","B:/","C:/","D:/",
"E:/","F:/","G:/","H:/",
"I:/","J:/","K:/","L:/",
"M:/","N:/","O:/","P:/",
"Q:/","R:/","S:/","T:/",
"U:/","V:/","W:/","X:/",
"Y:/","Z:/",
}
var isInSelectionType byte=0
var excludeCommands bool=false
var invalidateCommands bool=false
var strictMode bool=false


func LS() {
	outArr:=[]string{"../"}
	if curpath[0]==":" {
		outArr=append(outArr,windowsDrives...)
		return
	}
	cpath:=strings.Join(curpath, "")
	files, err := ioutil.ReadDir(cpath)
	if err != nil {
		curSelArr=outArr
		return
	}
	for _, fl := range files {
		if(fl.Name()[0]=='.') {
			continue
		}
		if(fl.IsDir()) {
			outArr=append(outArr,fmt.Sprintf("%s/",fl.Name()))
		}else{
			outArr=append(outArr,fmt.Sprintf("%s",fl.Name()))
		}
	}
	curSelArr=outArr
	return
}

func SetRootForCurPath() bool {
	if runtime.GOOS == "windows" {
		curpath=[]string { ":" }
		return true
	}
	_, err:=ioutil.ReadDir("/")
	if(err==nil) {
		curpath=[]string { "/" }
		return true
	}
	_, err=ioutil.ReadDir("/storage/emulated/0/")
	if(err==nil) {
		curpath=[]string { "/storage/emulated/0/" }
		return true
	}
	_, err=ioutil.ReadDir("/home/")
	if(err==nil) {
		curpath=[]string { "/home/" }
		return true
	}
	return false
}

func OpenSubMenu(sel int, env *environment.PBEnvironment) bool {
	fh := env.FunctionHolder
	if(isInSelectionType==1){
		if(sel==0) {
			isInSelectionType=0
			levelArr=[]string{}
			curSelArr = SelectableCommands
			return false
		}else if(sel==1) {
			excludeCommands=!excludeCommands
		}else if(sel==2) {
			invalidateCommands=!invalidateCommands
		}else if(sel==3) {
			strictMode=!strictMode
		}else if(sel==4) {
			if excludeCommands {
				levelArr=append(levelArr,"--excludecommands")
			}
			if invalidateCommands {
				levelArr=append(levelArr,"--invalidatecommands")
			}
			if strictMode {
				levelArr=append(levelArr,"-S")
			}
			fh.Process(strings.Join(levelArr, " "))
			return true
		}
		curSelArr=[]string{"< Back",fmt.Sprintf("%s = %v",I18n.T(I18n.Menu_ExcludeCommandsOption),excludeCommands),fmt.Sprintf("%s = %v",I18n.T(I18n.Menu_InvalidateCommandsOption),invalidateCommands),fmt.Sprintf("%s = %v",I18n.T(I18n.Menu_StrictModeOption),strictMode),"OK >"}
		return false
	}else if(isInSelectionType==2){
		if(sel==0) {
			isInSelectionType=0
			levelArr=[]string{}
			curSelArr = SelectableCommands
			return false
		}else if(sel==1) {
			excludeCommands=!excludeCommands
		}else if(sel==2) {
			if excludeCommands {
				levelArr=append(levelArr,"--excludecommands")
			}
			fh.Process(strings.Join(levelArr, " "))
			return true
		}
		curSelArr=[]string{I18n.T(I18n.Menu_BackButton),fmt.Sprintf("%s = %v",I18n.T(I18n.Menu_ExcludeCommandsOption),excludeCommands),"OK >"}
		return false
	}
	if(len(curpath)!=0) {
		if(!strings.HasSuffix(curSelArr[sel],"/")) {
			flpath:=strings.Join(append(curpath,curSelArr[sel]),"")
			levelArr=append(levelArr,fmt.Sprintf("\"%s\"",flpath))
			curpath=[]string{}
			if(levelArr[0]!="bdump"&&levelArr[0]!="export") {
				fh.Process(strings.Join(levelArr, " "))
				return true
			}
			if(levelArr[0]=="bdump") {
				isInSelectionType=1
				curSelArr=[]string{I18n.T(I18n.Menu_BackButton),fmt.Sprintf("%s = %v",I18n.T(I18n.Menu_ExcludeCommandsOption),excludeCommands),fmt.Sprintf("%s = %v",I18n.T(I18n.Menu_InvalidateCommandsOption),invalidateCommands),fmt.Sprintf("%s = %v",I18n.T(I18n.Menu_StrictModeOption),strictMode),"OK >"}
			}else if(levelArr[0]=="export") {
				isInSelectionType=2
				curSelArr=[]string{I18n.T(I18n.Menu_BackButton),fmt.Sprintf("%s = %v",I18n.T(I18n.Menu_ExcludeCommandsOption),excludeCommands),"OK >"}
			}
			return false
		}
		if(curSelArr[sel]=="../") {
			curpath=curpath[:len(curpath)-1]
			if len(curpath)==0 {
				levelArr=[]string{}
				curSelArr = SelectableCommands
			}else{
				LS()
			}
			return false
		}else{
			curpath=append(curpath,curSelArr[sel])
			LS()
			return false
		}
	}
	if(len(levelArr)==0) {
		if(sel==0) {
			fh.Process("get")
			return true
		}else if(sel==1) {
			fh.Process("get end")
			return true
		}else if(sel==2) {
			levelArr=[]string { "acme", "-p" }
			if (!SetRootForCurPath()) {
				return true
			}
			LS()
			return false
		}else if(sel==3) {
			levelArr=[]string { "bdump", "-p" }
			if (!SetRootForCurPath()) {
				return true
			}
			LS()
			return false
		}else if(sel==4) {
			levelArr=[]string { "schem", "-p" }
			if (!SetRootForCurPath()) {
				return true
			}
			LS()
			return false
		}else if(sel==5) {
			levelArr=[]string { "plot", "-p" }
			if (!SetRootForCurPath()) {
				return true
			}
			LS()
			return false
		}else if(sel==6) {
			levelArr=[]string { "export", "-p" }
			if (!SetRootForCurPath()) {
				return true
			}
			LS()
			return false
		}else if(sel==7) {
			fh.Process("fbexit")
			return true
		}else if(sel==8) {
			return true
		}
	}
	return true
}

func OpenMenu(env *environment.PBEnvironment) {
	conn:=env.Connection.(*minecraft.Conn)
	go func() {
		curSelArr = SelectableCommands
		curpath=[]string{}
		levelArr=[]string{}
		isInSelectionType=0
		if(OPRuntimeId==0) {
			player:=env.RespondUser
			lineUUID,_:=uuid.NewUUID()
			lineChan:=make(chan *packet.CommandOutput)
			(*env.CommandSender.GetUUIDMap()).Store(lineUUID.String(),lineChan)
			env.CommandSender.SendWSCommand(fmt.Sprintf("execute %s ~ ~ ~ tp %s ~400 ~ ~",player,conn.IdentityData().DisplayName),lineUUID)
			<-lineChan
			close(lineChan)
			OPRuntimeIdReceivedChannel=make(chan bool)
			dispUUID,_:=uuid.NewUUID()
			env.CommandSender.SendWSCommand(fmt.Sprintf("execute %s ~ ~ ~ tp %s ~ ~2 ~",player,conn.IdentityData().DisplayName),dispUUID)
			<-OPRuntimeIdReceivedChannel
			close(OPRuntimeIdReceivedChannel)
			OPRuntimeIdReceivedChannel=nil
		}
		curSel:=0
		for {
			player:=env.RespondUser
			dispUUID,_:=uuid.NewUUID()
			env.CommandSender.SendWSCommand(fmt.Sprintf("execute %s ~ ~ ~ tp %s ~ ~2 ~",player,conn.IdentityData().DisplayName),dispUUID)
			if(LastOPPitch<(-60)) {
				curSel--
				if(curSel<0) {
					curSel=len(curSelArr)-1
				}
			}else if(LastOPPitch>60) {
				curSel++
				if(curSel>=len(curSelArr)) {
					curSel=0
				}
			}
			var cSessionArr []string
			if len(curpath)!=0 {
				cSessionArr=append(cSessionArr,fmt.Sprintf("%s: %s",I18n.T(I18n.Menu_CurrentPath),strings.Join(curpath,"")))
			}
			for i:=curSel-3;i<len(curSelArr);i++ {
				if i<0 {
					continue
				}
				if len(cSessionArr)==0 && i!=0 {
					cSessionArr=append(cSessionArr,"^ (...) ^")
				}
				if i==curSel {
					cSessionArr=append(cSessionArr,fmt.Sprintf("§b>%s<§r", curSelArr[i]))
				}else{
					cSessionArr=append(cSessionArr,curSelArr[i])
				}
				if len(cSessionArr)==8 && i+1<len(curSelArr) {
					cSessionArr=append(cSessionArr,"_ (...) _")
					break
				}
			}
			//cSessionArr[curSel]=fmt.Sprintf("§b>%s<§r", curSelArr[curSel])
			taskholder:=env.TaskHolder.(*fbtask.TaskHolder)
			taskholder.ExtraDisplayStrings=append([]string{fmt.Sprintf("Pitch: %v",LastOPPitch)},cSessionArr...)
			env.ActivateTaskStatus<-true
			if(LastOPSneak&&!LastOPMouSneaked) {
				//LastOPSneak=false
				LastOPMouSneaked=true
				quitmenu:=OpenSubMenu(curSel, env)
				if(quitmenu) {
					taskholder.ExtraDisplayStrings=[]string{}
					return
				}
				curSel=0
				continue
			}
			//fbtask.ExtraDisplayStrings=[]string{fmt.Sprintf("Pitch: %v, Sneak: %v",LastOPPitch,LastOPSneak)}
			time.Sleep(time.Duration(600)*time.Millisecond)
		}
	} ()
}