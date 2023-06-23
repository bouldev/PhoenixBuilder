package defines

import (
	"phoenixbuilder/fastbuilder/uqHolder"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/mirror"
	"phoenixbuilder/mirror/define"
	"phoenixbuilder/mirror/io/assembler"
	"phoenixbuilder/mirror/io/world"
	"time"

	"github.com/google/uuid"
)

type LineSource interface {
	Read() string
}

type LineDst interface {
	Write(string)
}

type NoSqlDB interface {
	Get(key string) string
	Delete(key string)
	Commit(key string, v string)
	IterAll(cb func(key string, v string) (stop bool))
	IterWithPrefix(cb func(key string, v string) (stop bool), prefix string)
	IterWithRange(cb func(key string, v string) (stop bool), start, end string)
}

type GameChat struct {
	Name               string
	Msg                []string
	Type               byte
	FrameWorkTriggered bool
	FallBack           bool
	RawMsg             string
	RawName            string
	RawParameters      []string
	// Aux                interface{}
}

type MenuEntry struct {
	Triggers     []string
	ArgumentHint string
	FinalTrigger bool
	Usage        string
}

type Cmd struct {
	Conditinal  bool    `json:"有条件"`
	Cmd         string  `json:"指令"`
	SleepBefore float32 `json:"执行前延迟"`
	Sleep       float32 `json:"执行后延迟"`
	Record      string  `json:"结果记录"`
	As          string  `json:"身份"`
	Note        string  `json:"备注"`
}

type CmdsWithName struct {
	Name   string                 `json:"玩家"`
	Cmds   []Cmd                  `json:"指令"`
	Params map[string]interface{} `json:"参数"`
}

type Currency struct {
	CurrencyName   string `json:"货币名"`
	ScoreboardName string `json:"记分板名"`
}

type VerificationRule struct {
	Enable     bool     `json:"启用身份验证"`
	BySelector string   `json:"依据选择器"`
	ByNameList []string `json:"依据名字"`
}

type GameMenuEntry struct {
	MenuEntry
	OptionalOnTriggerFn func(chat *GameChat) (stop bool)
	Verification        *VerificationRule
}

type BackendMenuEntry struct {
	MenuEntry
	OptionalOnTriggerFn func(cmds []string) (stop bool)
}

// CtxProvider 旨在帮助插件发现别的插件主动暴露的接口 GetContext()
// GetUQHolder() 可以获得框架代为维持的信息
type CtxProvider interface {
	// GetAllContext() *map[string]interface{}
	GetContext(key string) (entry interface{}, hasK bool)
	SetContext(key string, entry interface{})
	GetUQHolder() *uqHolder.UQHolder
}

// ConfigProvider 是帮助一个插件获得和修改别的插件的接口
// 如果仅仅需要自己的配置，这是不必要的
type ConfigProvider interface {
	GetOmegaConfig() *OmegaConfig
	GetAllConfigs() []*ComponentConfig
}

// 框架帮忙提供的储存机制，目的在于共享而非沙箱隔离
type StorageAndLogProvider interface {
	GetLogger(topic string) LineDst
	//GetNoSqlDB(topic string) NoSqlDB
	GetRelativeFileName(topic string) string
	GetFileData(topic string) ([]byte, error)
	GetJsonData(topic string, data interface{}) error
	GetStorageRoot() string
	GetWorldsDir(elem ...string) string
	GetOmegaSideDir(elem ...string) string
	GetOmegaCacheDir(elem ...string) string
	GetOmegaNormalCacheDir(elem ...string) string
	WriteFileData(topic string, data []byte) error
	WriteJsonData(topic string, data interface{}) error
	WriteJsonDataWithTMP(topic string, tmpSuffix string, data interface{}) error
}

// 与后端的交互接口
type BackendInteract interface {
	GetBackendDisplay() LineDst
	SetBackendMenuEntry(entry *BackendMenuEntry)
	SetBackendCmdInterceptor(func(cmds []string) (stop bool))
}

type ExtendOperation interface {
}

// 与游戏的交互接口，通过发出点什么来影响游戏
// 建议扩展该接口以提供更丰富的功能
// 另一种扩展方式是定义新插件并暴露接口
type GameControl interface {
	SendMCPacket(packet.Packet)
	SayTo(target string, msg string)
	RawSayTo(target string, msg string)
	ActionBarTo(target string, msg string)
	TitleTo(target string, msg string)
	SubTitleTo(target string, msg string)
	SendBytes([]byte)
	SendCmd(cmd string)
	SendCmdWithUUID(cmd string, ud uuid.UUID, ws bool)
	SendWOCmd(cmd string)
	SendCmdAndInvokeOnResponse(string, func(output *packet.CommandOutput))
	SendCmdAndInvokeOnResponseWithFeedback(string, func(output *packet.CommandOutput))
	GetPlayerKit(name string) PlayerKit
	GetPlayerKitByUUID(ud uuid.UUID) PlayerKit
	SetOnParamMsg(string, func(chat *GameChat) (catch bool)) error
	PlaceCommandBlock(pos define.CubePos, commandBlockName string, commandBlockData int,
		withMove, withAirPrePlace bool, updatePacket *packet.CommandBlockUpdate,
		onDone func(done bool), timeOut time.Duration)
}

type PlayerKit interface {
	Say(msg string)
	RawSay(msg string)
	ActionBar(msg string)
	Title(msg string)
	SubTitle(msg string)
	GetRelatedUQ() *uqHolder.Player

	GetViolatedStorage() map[string]interface{}
	//GetPersistStorage(k string) string
	//CommitPersistStorageChange(k string, v string)

	SetOnParamMsg(func(chat *GameChat) (catch bool)) error
	GetOnParamMsg() func(chat *GameChat) (catch bool)

	HasPermission(key string) bool
	SetPermission(key string, b bool)

	GetPos(selector string) chan *define.CubePos
}

// 与游戏的交互接口，如何捕获和处理游戏的数据包和消息
type GameListener interface {
	GetChunkAssembler() *assembler.Assembler
	SetOnAnyPacketCallBack(func(packet.Packet))
	SetOnAnyPacketBytesCallBack(func([]byte))
	SetOnTypedPacketCallBack(uint32, func(packet.Packet))
	SetGameMenuEntry(entry *GameMenuEntry)
	SetGameChatInterceptor(func(chat *GameChat) (stop bool))
	SetOnLevelChunkCallBack(fn func(cd *mirror.ChunkData))
	AppendOnKnownPlayerExistCallback(cb func(string))
	AppendLoginInfoCallback(cb func(entry protocol.PlayerListEntry))
	AppendLogoutInfoCallback(cb func(entry protocol.PlayerListEntry))
	Throw(chat *GameChat)
	GetTriggerWord() string
	AppendOnBlockUpdateInfoCallBack(cb func(pos define.CubePos, origRTID uint32, currentRTID uint32))
}

// 安全事件发送和处理，比如某插件发现有玩家在恶意修改设置
// 而另一个插件则在 QQ 群里通知这个事件的发生
type SecurityEventIO interface {
	QuerySensitiveInfo(SensitiveInfoType) (string, error)
	RedAlert(info string)
	RegOnAlertHandler(cb func(info string))
}

type BasicBotTask struct {
	Name       string
	ActivateFn func()
}

func (bbt *BasicBotTask) Activate() {
	// fmt.Println("Executing " + bbt.Name)
	bbt.ActivateFn()
}

type BotTask interface {
	Activate()
}

type BasicBotTaskPauseAble struct {
	BasicBotTask
}

func (bbtpa *BasicBotTaskPauseAble) Pause() {
	// fmt.Println("Pasue " + bbtpa.Name)
}

func (bbtpa *BasicBotTaskPauseAble) Resume() {
	// fmt.Println("Resume " + bbtpa.Name)
}

type BotTaskPauseAble interface {
	BotTask
	Pause()
	Resume()
}

type BotTaskScheduler interface {
	// background task will not execute if a normal task or urgent task exist
	CommitBackgroundTask(BotTaskPauseAble) (reject, pending bool)
	// a normal task will pause the executing background task
	CommitNormalTask(BotTaskPauseAble) (pending bool)
	// a urgent task will pause normal task and background task
	CommitUrgentTask(BotTask) (pending bool)
}

type MainFrame interface {
	CtxProvider
	ConfigProvider
	StorageAndLogProvider
	BackendInteract
	SecurityEventIO
	FatalError(err string)
	GetGameControl() GameControl
	GetGameListener() GameListener
	GetBotTaskScheduler() BotTaskScheduler
	GetWorld() *world.World
	GetWorldProvider() mirror.ChunkProvider
	GetScoreboardHolder() *ScoreBoardHolder
	FBEval(cmd string)
	AllowChunkRequestCache()
	NoChunkRequestCache()
}
