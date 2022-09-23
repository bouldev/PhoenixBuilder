//go:build fyne_gui
// +build fyne_gui

package session

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"phoenixbuilder/fastbuilder/commands_generator"
	"phoenixbuilder/fastbuilder/configuration"
	"phoenixbuilder/fastbuilder/core"

	"phoenixbuilder/bridge/bridge_fmt"
	"phoenixbuilder/fastbuilder/args"
	fbauth "phoenixbuilder/fastbuilder/cv4/auth"
	"phoenixbuilder/fastbuilder/environment"
	I18n "phoenixbuilder/fastbuilder/i18n"
	"phoenixbuilder_fyne_gui/platform_helper"
)

type SessionConfig struct {
	Lang          string `yaml:"lang" json:"lang"`
	FBUserName    string `yaml:"fb_username" json:"fb_username"`
	FBPassword    string `yaml:"fb_password" json:"fb_password"`
	FBToken       string `yaml:"fb_token" json:"fb_token"`
	ServerCode    string `yaml:"server_code" json:"server_code"`
	ServerPasswd  string `yaml:"server_passwd" json:"server_passwd"`
	RespondUser   string `yaml:"respond_user" json:"respond_user"`
	MuteWorldChat bool   `yaml:"mute_world_chat" json:"mute_world_chat"`
	devMode       bool
	// when "iamDeveloper" is true, the following fields are used,
	// otherwise, the fields are ignored (restore to default)
	NoPyRPC    bool   `yaml:"no_py_rpc" json:"no_py_rpc"`
	FBVersion  string `yaml:"fb_version" json:"fb_version"`
	FBHash     string `yaml:"fb_hash" json:"fb_hash"`
	FBCodeName string `yaml:"fb_codename" json:"fb_codename"`
}

func (config *SessionConfig) IsDeveloper() bool {
	return config.devMode
}

func NewConfig() *SessionConfig {
	return &SessionConfig{
		Lang:          "zh_CN", // "en_US"
		FBUserName:    "",
		FBPassword:    "",
		FBToken:       "",
		ServerCode:    "",
		ServerPasswd:  "",
		RespondUser:   "",
		devMode:       false,
		MuteWorldChat: false,
		NoPyRPC:       true,
		FBVersion:     args.GetFBVersion(),
		FBHash:        "gui~" + args.GetFBPlainVersion(),
		FBCodeName:    DefaultFBCodeName,
	}
}

type Session struct {
	// can use this to terminate the session
	stopChan chan struct{}

	// can use this to send command
	cmdChan          chan string
	closeFns         []func()
	worldChatChannel chan []string
	env              *environment.PBEnvironment
	botRuntimeID     string
	Config           *SessionConfig
	// set/ set end callback
	CmdSetCbFn    func(X, Y, Z int)
	CmdSetEndCbFn func(X, Y, Z int)
}

type FBPlainToken struct {
	EncryptToken bool   `json:"encrypt_token"`
	Username     string `json:"username"`
	Password     string `json:"password"`
}

var isStart bool

func init() {
	I18n.Init()
	isStart = false
}

func NewSession(config *SessionConfig) *Session {
	// it's weird that we need to do this, because actually we can only hold one session
	// but maybe in the future we can support multiple sessions
	if isStart {
		return nil
	}

	config.devMode = false

	if !config.devMode {
		defaultConfig := NewConfig()
		config.NoPyRPC = defaultConfig.NoPyRPC
		config.FBVersion = defaultConfig.FBVersion
		config.FBHash = defaultConfig.FBHash
		config.FBCodeName = defaultConfig.FBCodeName
	}

	session := &Session{
		stopChan:      make(chan struct{}),
		cmdChan:       make(chan string),
		closeFns:      make([]func(), 0),
		Config:        config,
		CmdSetCbFn:    func(X, Y, Z int) {},
		CmdSetEndCbFn: func(X, Y, Z int) {},
	}
	// configuration.MonkeyPathFileReader = make(map[string]fyne.URIReadCloser)
	// configuration.MonkeyPathFileWriter = make(map[string]fyne.URIWriteCloser)
	I18n.SelectedLanguage = config.Lang
	I18n.UpdateLanguage()
	return session
}

// func (s *Session) NewMonkeyPathReader(path string, fp fyne.URIReadCloser) {
// 	configuration.MonkeyPathFileReader[path] = fp
// }

// func (s *Session) NewMonkeyPathWriter(path string, fp fyne.URIWriteCloser) {
// 	configuration.MonkeyPathFileWriter[path] = fp
// }

func (s *Session) GetEnvironment() *environment.PBEnvironment {
	return s.env
}

func (s *Session) Start() (terminateChan chan string, startErr error) {
	// we need to make sure no multiple session is running
	if isStart {
		return nil, fmt.Errorf("Session is already started")
	}

	// before we start, we need to make sure that the session is valid
	// if not, we need to return an error

	err := s.beforeStart()
	if err != nil {
		return nil, err
	}

	// after we start, we need to return a channel that we can use to
	// notify reciver of this chan that the session is terminated
	// and the reason for termination

	isStart = true
	// when the session is terminated, we need to notify the caller
	configuration.UserToken = s.Config.FBToken
	c := s.afterStart()
	return c, nil
}

func (s *Session) afterStart() chan string {
	c := make(chan string)
	platform_helper.RunBackground()
	go s.routine(c)
	return c
}

func (s *Session) beforeStart() (err error) {
	configuration.SessionInitID += 1
	core.PassFatal = true
	// in this function, we need to make sure that the session is valid
	// first, we need to connect to the fb auth server and get the token
	// then, we try connecting to netease mc server

	// but first, we need to deal with the panic hidden in the code
	defer func() {
		r := recover()
		if r != nil {
			err = fmt.Errorf("Session Start Fail, because a panic occurred: \n%v", r)
		}
	}()

	// check credentials
	if (s.Config.FBUserName == "" || s.Config.FBPassword == "") && s.Config.FBToken == "" {
		return fmt.Errorf("no credientials provided")
	}

	// check server configuration
	if s.Config.ServerCode == "" {
		return fmt.Errorf("no server code provided")
	}

	var env *environment.PBEnvironment
	if s.Config.devMode {
		env = core.InitDebugEnvironment()
	}else{
		env = core.InitRealEnvironment(s.Config.FBToken, s.Config.ServerCode, s.Config.ServerPasswd)
	}
	authClient := fbauth.CreateClient(env)
	env.FBAuthClient = authClient
	if s.Config.FBToken == "" {
		// we need to get a token
		tokenReq := &FBPlainToken{
			EncryptToken: true,
			Username:     s.Config.FBUserName,
			Password:     s.Config.FBPassword,
		}
		tokenReqStr, err := json.Marshal(tokenReq)
		if err != nil {
			return fmt.Errorf("cannot marshal token request to json: \n%v", err)
		}
		token := authClient.GetToken("", string(tokenReqStr))
		if token == "" {
			return fmt.Errorf("cannot get token: \n" + I18n.T(I18n.FBUC_LoginFailed))
		}
		s.Config.FBToken = token
		env.LoginInfo.Token = token
	}
	core.InitClient(env)
	s.closeFns = append(s.closeFns, func() {
		core.DestroyClient(env)
	})
	bridge_fmt.Println(I18n.T(I18n.ConnectionEstablished))
	return nil
}

func (s *Session) routine(c chan string) {
	terminateReason := "Session terminated by user"
	defer func() {
		// we don't want the whole program to exit when there is a panic
		// hidden in the code
		r := recover()
		if r != nil {
			terminateReason = fmt.Sprintf("Session terminated\n because a panic occurred in routine: \n%v", r)
		} else {
			platform_helper.StopBackground()
		}
		s.close()
		c <- terminateReason
	}()

	go func() {
		defer func() {
			// we don't want the whole program to exit when there is a panic
			// hidden in the code
			r := recover()
			if r != nil {
				terminateReason = fmt.Sprintf("Session terminated\n because a panic occurred in Process Function: \n%v", r)
			}
			s.close()
			c <- terminateReason
		}()
		core.EnterReadlineThread(s.env, s.stopChan)
	}()
	
	// A loop that reads packets from the connection until it is closed.
	core.EnterWorkerThread(s.env, s.stopChan)
}

func (s *Session) GetPos() (x, y, z int) {
	return configuration.GlobalFullConfig(s.env).Main().Position.X, configuration.GlobalFullConfig(s.env).Main().Position.Y, configuration.GlobalFullConfig(s.env).Main().Position.Z
}

func (s *Session) GetEndPos() (x, y, z int) {
	return configuration.GlobalFullConfig(s.env).Main().End.X, configuration.GlobalFullConfig(s.env).Main().End.Y, configuration.GlobalFullConfig(s.env).Main().End.Z
}

func (s *Session) sendCommand(commands string, UUID uuid.UUID) error {
	return s.env.CommandSender.SendCommand(commands, UUID)
}

func (s *Session) tellraw(message string) error {
	commands_generator.AdditionalChatCb(message)
	return s.env.CommandSender.Output(message)
}

func (s *Session) close() {
	isStart = false
	for _, fn := range s.closeFns {
		fn()
	}
	// let GC do the work
	s.env = nil
}

func (s *Session) Execute(cmd string) {
	s.cmdChan <- cmd
}

func (s *Session) Stop() {
	// close the stopChan to nofitify the routine to stop session
	close(s.stopChan)
}
