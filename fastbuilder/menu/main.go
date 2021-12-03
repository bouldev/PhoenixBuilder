package menu

import (
	"time"
	"fmt"
	"phoenixbuilder/minecraft"
	"phoenixbuilder/fastbuilder/command"
	fbtask "phoenixbuilder/fastbuilder/task"
	"phoenixbuilder/minecraft/protocol/packet"
	"github.com/google/uuid"
	"phoenixbuilder/fastbuilder/configuration"
	"phoenixbuilder/fastbuilder/function"
	"io/ioutil"
	"strings"
	"runtime"
)

var OPRuntimeId uint64 = 0
var OPRuntimeIdReceivedChannel chan bool = nil
var LastOPPitch float32 = 0
var LastOPSneak bool = false
var LastOPMouSneaked bool = false
var SelectableCommands = []string {
	"getPos",
	"getEndPos",
	"ACME",
	"bdump",
	"SCH-",
	"PL0T",
	"EXPORT",
	"Quit Program",
	"Cancel",
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

func OpenSubMenu(sel int, conn *minecraft.Conn) bool {
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
			function.Process(conn,strings.Join(levelArr, " "))
			return true
		}
		curSelArr=[]string{"< Back",fmt.Sprintf("Exclude Commands = %v",excludeCommands),fmt.Sprintf("Invalidate Commands = %v",invalidateCommands),fmt.Sprintf("Strict Mode = %v",strictMode),"OK >"}
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
			function.Process(conn,strings.Join(levelArr, " "))
			return true
		}
		curSelArr=[]string{"< Back",fmt.Sprintf("Exclude Commands = %v",excludeCommands),"OK >"}
		return false
	}
	if(len(curpath)!=0) {
		if(!strings.HasSuffix(curSelArr[sel],"/")) {
			flpath:=strings.Join(append(curpath,curSelArr[sel]),"")
			levelArr=append(levelArr,fmt.Sprintf("\"%s\"",flpath))
			curpath=[]string{}
			if(levelArr[0]!="bdump"&&levelArr[0]!="export") {
				function.Process(conn, strings.Join(levelArr, " "))
				return true
			}
			if(levelArr[0]=="bdump") {
				isInSelectionType=1
				curSelArr=[]string{"< Back",fmt.Sprintf("Exclude Commands = %v",excludeCommands),fmt.Sprintf("Invalidate Commands = %v",invalidateCommands),fmt.Sprintf("Strict Mode = %v",strictMode),"OK >"}
			}else if(levelArr[0]=="export") {
				isInSelectionType=2
				curSelArr=[]string{"< Back",fmt.Sprintf("Exclude Commands = %v",excludeCommands),"OK >"}
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
			function.Process(conn, "get")
			return true
		}else if(sel==1) {
			function.Process(conn, "get end")
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
			function.Process(conn, "fbexit")
			return true
		}else if(sel==8) {
			return true
		}
	}
	return true
}

func OpenMenu(conn *minecraft.Conn) {
	go func() {
		curSelArr = SelectableCommands
		curpath=[]string{}
		levelArr=[]string{}
		isInSelectionType=0
		if(OPRuntimeId==0) {
			player:=configuration.RespondUser
			lineUUID,_:=uuid.NewUUID()
			lineChan:=make(chan *packet.CommandOutput) //　蓮(れん)ちゃん (
			command.UUIDMap.Store(lineUUID.String(),lineChan)
			command.SendWSCommand(fmt.Sprintf("execute %s ~ ~ ~ tp %s ~400 ~ ~",player,conn.IdentityData().DisplayName),lineUUID,conn)
			<-lineChan
			close(lineChan)
			OPRuntimeIdReceivedChannel=make(chan bool)
			dispUUID,_:=uuid.NewUUID()
			command.SendWSCommand(fmt.Sprintf("execute %s ~ ~ ~ tp %s ~ ~2 ~",player,conn.IdentityData().DisplayName),dispUUID,conn)
			<-OPRuntimeIdReceivedChannel
			close(OPRuntimeIdReceivedChannel)
			OPRuntimeIdReceivedChannel=nil
		}
		curSel:=0
		for {
			player:=configuration.RespondUser
			dispUUID,_:=uuid.NewUUID()
			command.SendWSCommand(fmt.Sprintf("execute %s ~ ~ ~ tp %s ~ ~2 ~",player,conn.IdentityData().DisplayName),dispUUID,conn)
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
				cSessionArr=append(cSessionArr,fmt.Sprintf("Current path: %s",strings.Join(curpath,"")))
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
			fbtask.ExtraDisplayStrings=append([]string{fmt.Sprintf("Pitch: %v",LastOPPitch)},cSessionArr...)
			fbtask.ActivateTaskStatus<-true
			if(LastOPSneak&&!LastOPMouSneaked) {
				//LastOPSneak=false
				LastOPMouSneaked=true
				quitmenu:=OpenSubMenu(curSel, conn)
				if(quitmenu) {
					fbtask.ExtraDisplayStrings=[]string{}
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