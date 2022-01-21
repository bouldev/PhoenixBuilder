package I18n

import (
	"path/filepath"
	"os"
	"fmt"
	"bufio"
	"strings"
	"strconv"
	"io/ioutil"
)

const (
	LanguageEnglish = "en_US"
	LanguageSimplifiedChinese = "zh_CN"
)

var SelectedLanguage = LanguageEnglish

const (
	Special_Startup = iota
	Copyright_Notice_Headline
	Copyright_Notice_Line_1
	Copyright_Notice_Line_2
	Copyright_Notice_Line_3
	Copyright_Notice_Line_4 // Not used
	Crashed_Tip
	Crashed_StackDump_And_Error
	Crashed_OS_Windows
	EnterPasswordForFBUC
	FBUC_LoginFailed
	ServerCodeTrans
	ConnectionEstablished
	InvalidPosition
	PositionGot
	PositionGot_End
	Enter_FBUC_Username
	Enter_Rental_Server_Code
	Enter_Rental_Server_Password
	NotAnACMEFile
	UnsupportedACMEVersion
	ACME_FailedToSeek
	ACME_FailedToGetCommand
	ACME_StructureErrorNotice
	ACME_UnknownCommand
	SysError_NoTranslation // Do not add a translation for it!
	SysError_HasTranslation
	SysError_EACCES
	SysError_EBUSY
	SysError_EINVAL
	SysError_EISDIR
	SysError_ENOENT
	SysError_ETXTBSY
	BDump_EarlyEOFRightWhenOpening
	BDump_NotBDX_Invheader
	InvalidFileError
	BDump_SignedVerifying
	BDump_VerificationFailedFor
	ERRORStr
	IgnoredStr
	BDump_FileSigned
	BDump_FileNotSigned
	BDump_NotBDX_Invinnerheader
	BDump_FailedToReadAuthorInfo
	FileCorruptedError
	BDump_Author
	CommandNotFound
	Sch_FailedToResolve
	SimpleParser_Too_few_args
	SimpleParser_Invalid_decider
	SimpleParser_Int_ParsingFailed
	SimpleParser_InvEnum
	QuitCorrectly
	PositionSet
	PositionSet_End
	DelaySetUnavailableUnderNoneMode
	CurrentDefaultDelayMode
	DelaySet
	DelayModeSet
	DelayModeSet_DelayAuto
	DelayModeSet_ThresholdAuto
	DelayThreshold_OnlyDiscrete
	DelayThreshold_Set
	CurrentTasks
	TaskStateLine
	TaskTotalCount
	TaskNotFoundMessage
	TaskPausedNotice
	TaskResumedNotice
	TaskStoppedNotice
	Task_SetDelay_Unavailable
	Task_DelaySet
	TaskTTeIuKoto
	TaskTypeSwitchedTo
	TaskDisplayModeSet
	TaskCreated
	Menu_GetPos
	Menu_GetEndPos
	Menu_Quit
	Menu_Cancel
	Menu_ExcludeCommandsOption
	Menu_InvalidateCommandsOption
	Menu_StrictModeOption
	Menu_BackButton
	Menu_CurrentPath
	Parsing_UnterminatedQuotedString
	Parsing_UnterminatedEscape
	Get_Warning
	LanguageName
	TaskTypeUnknown
	TaskTypeRunning
	TaskTypePaused
	TaskTypeDied
	TaskTypeCalculating
	TaskTypeSpecialTaskBreaking
	TaskFailedToParseCommand
	Task_D_NothingGenerated
	Task_Summary_1
	Task_Summary_2
	Task_Summary_3
	Logout_Done
	FailedToRemoveToken
	SelectLanguageOnConsole
	LanguageUpdated
	Auth_ServerNotFound // 104
	Auth_FailedToRequestEntry // 105
	Auth_InvalidHelperUsername // 106
	Auth_BackendError //107
	Auth_UnauthorizedRentalServerNumber //108
	Auth_HelperNotCreated //109
	Auth_InvalidUser //110
	Auth_InvalidToken //111
	Auth_UserCombined //112
	Auth_InvalidFBVersion //113
	Notify_TurnOnCmdFeedBack
	Notify_NeedOp
)

var LangDict map[string]map[uint16]string = map[string]map[uint16]string {
	LanguageEnglish: I18nDict_en,
	LanguageSimplifiedChinese: I18nDict_cn,
}

var I18nDict map[uint16]string

func ShouldDisplaySpecial() bool {
	_, has:=I18nDict[Special_Startup]
	return has
}

func HasTranslationFor(transtype uint16) bool {
	_, has:=I18nDict[transtype]
	return has
}

func SelectLanguage() {
	config:=loadConfigPath()
	curLangDict:=make(map[uint16]string)
	{
		i:=1
		for lang:=range LangDict {
			curLangDict[uint16(i)]=lang
			fmt.Printf("[%d] %s\n",i,LangDict[lang][LanguageName])
			i++
		}
	}
	reader:=bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("(ID): ") // No \n
		inp, _ := reader.ReadString('\n')
		inpl:=strings.TrimRight(inp,"\r\n")
		parsedInt, err := strconv.Atoi(inpl)
		if(err!=nil) {
			continue
		}
		if(parsedInt<=0||parsedInt>len(curLangDict)) {
			continue
		}
		SelectedLanguage=curLangDict[uint16(parsedInt)]
		break
	}
	if file,err:=os.Create(config);err!=nil {
		fmt.Println("Error creating language config file: %v",err)
		fmt.Println("Error ignored.")
	}else{
		_, err = file.WriteString(SelectedLanguage)
		if(err!=nil) {
			fmt.Println("Error saving language config: %v",err)
			fmt.Println("Error ignored.")
		}
		file.Close()
	}
}

func UpdateLanguage() {
	langdict, aru := LangDict[SelectedLanguage]
	if(!aru) {
		panic("Updating to a language that doesn't exist")
		return
	}
	I18nDict=langdict
	fmt.Printf("%s\n",T(LanguageUpdated))
}

func Init() {
	config:=loadConfigPath()
	if _, err:=os.Stat(config); os.IsNotExist(err) {
		SelectLanguage()
	}else{
		content, err:=ioutil.ReadFile(config)
		if (err != nil) {
			panic("Language config file isn't accessible")
			return
		}
		langCode:=string(content)
		SelectedLanguage=langCode
	}
	langdict, aru := LangDict[SelectedLanguage]
	if(!aru) {
		fmt.Printf("Ordered language doesn't exist.\nPlease reselect one:\n")
		SelectLanguage()
		langdict, aru=LangDict[SelectedLanguage]
		if !aru {
			panic("Language still unexists after reselection")
			return
		}
	}
	I18nDict=langdict
}

func T(code uint16) string {
	r, has := I18nDict[code]
	if !has {
		r,has=I18nDict_en[code]
		if !has {
			return "???"
		}
	}
	return r
}

func loadConfigPath() string {
	homedir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("[PLUGIN] WARNING - Failed to obtain the user's home directory. made homedir=\".\";\n")
		homedir="."
	}
	fbconfigdir := filepath.Join(homedir, ".config/fastbuilder")
	os.MkdirAll(fbconfigdir, 0755)
	file:=filepath.Join(fbconfigdir,"language")
	return file
}