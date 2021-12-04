package I18n

import (
	"path/filepath"
	"os"
	"fmt"
)

const (
	LanguageEnglish = iota
	LanguageSimplifiedChinese
)

var SelectedLanguage = LanguageSimplifiedChinese

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
)

var I18nDict map[uint16]string

func ShouldDisplaySpecial() bool {
	_, has:=I18nDict[Special_Startup]
	return has
}

func HasTranslationFor(transtype uint16) bool {
	_, has:=I18nDict[transtype]
	return has
}

func Init() {
	if(SelectedLanguage==LanguageEnglish) {
		I18nDict=I18nDict_en
	}else if(SelectedLanguage==LanguageSimplifiedChinese) {
		I18nDict=I18nDict_cn
	}else{
		panic("Invalid language setting")
	}
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