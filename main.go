package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"phoenixbuilder/fastbuilder/core"
	"phoenixbuilder/fastbuilder/args"
	I18n "phoenixbuilder/fastbuilder/i18n"
	script_bridge "phoenixbuilder/fastbuilder/script_engine/bridge"
	"phoenixbuilder/fastbuilder/utils"
	"strings"
	"syscall"

	"github.com/pterm/pterm"
	"golang.org/x/term"

	"phoenixbuilder/fastbuilder/readline"
	_ "phoenixbuilder/io"
	_ "phoenixbuilder/plantform_specific/fix_timer"
)

func main() {
	args.ParseArgs()
	if len(args.PackScripts()) != 0 {
		os.Exit(script_bridge.MakePackage(args.PackScripts(), args.PackScriptsOut()))
	}
	pterm.Error.Prefix = pterm.Prefix{
		Text:  "ERROR",
		Style: pterm.NewStyle(pterm.BgBlack, pterm.FgRed),
	}

	I18n.Init()

	pterm.DefaultBox.Println(pterm.LightCyan("https://github.com/LNSSPsd/PhoenixBuilder"))
	pterm.Println(pterm.Yellow(I18n.T(I18n.Copyright_Notice_Contrib)))
	pterm.Println(pterm.Yellow(I18n.T(I18n.Copyright_Notice_Bouldev)))
	pterm.Println(pterm.Yellow("PhoenixBuilder " + args.GetFBVersion()))

	// iSH.app specific, for foreground ability
	if _, err := os.Stat("/dev/location"); err == nil {
		// Call location service
		pterm.Println(pterm.Yellow(I18n.T(I18n.Notice_iSH_Location_Service)))
		cmd := exec.Command("ash", "-c", "cat /dev/location > /dev/null &")
		err := cmd.Start()
		if err != nil {
			fmt.Println(err)
		}
	}

	if !args.NoReadline() && !args.ShouldEnableOmegaSystem() {
		readline.InitReadline()
	}

	if I18n.ShouldDisplaySpecial() {
		fmt.Printf("%s", I18n.T(I18n.Special_Startup))
	}

	defer core.Fatal()
	if args.DebugMode() {
		init_and_run_debug_client()
		return
	}
	if !args.ShouldDisableHashCheck() {
		fmt.Printf(I18n.T(I18n.Notice_CheckUpdate))
		hasUpdate, latestVersion := utils.CheckUpdate(args.GetFBVersion())
		fmt.Printf(I18n.T(I18n.Notice_OK))
		if hasUpdate {
			fmt.Printf(I18n.T(I18n.Notice_UpdateAvailable), latestVersion)
			fmt.Printf(I18n.T(I18n.Notice_UpdateNotice))
			// To ensure user won't ignore it directly, can be suppressed by command line argument.
			os.Exit(0)
		}
	}

	if !args.SpecifiedToken() {
		token := loadTokenPath()
		if _, err := os.Stat(token); os.IsNotExist(err) {
			fbusername, err := getInputUserName()
			if err != nil {
				panic(err)
			}
			fbuntrim := fmt.Sprintf("%s", strings.TrimSuffix(fbusername, "\n"))
			fbun := strings.TrimRight(fbuntrim, "\r\n")
			fmt.Printf(I18n.T(I18n.EnterPasswordForFBUC))
			fbpassword, err := term.ReadPassword(int(syscall.Stdin))
			fmt.Printf("\n")
			tokenstruct := &map[string]interface{}{
				"encrypt_token": true,
				"username":      fbun,
				"password":      string(fbpassword),
			}
			token, err := json.Marshal(tokenstruct)
			if err != nil {
				fmt.Println(I18n.T(I18n.FBUC_Token_ErrOnGen))
				fmt.Println(err)
				return
			}
			runInteractiveClient(string(token))

		} else {
			token, err := readToken(token)
			if err != nil {
				fmt.Println(err)
				return
			}
			runInteractiveClient(token)
		}
	} else {
		runInteractiveClient(args.CustomTokenContent())
	}
}

func runInteractiveClient(token string) {
	var code, serverPasswd string
	var err error
	if !args.SpecifiedServer() {
		code, serverPasswd, err = getRentalServerCode()
	} else {
		code = args.ServerCode()
		serverPasswd = args.ServerPassword()
	}

	if err != nil {
		fmt.Println(err)
		return
	}
	env:=core.InitRealEnvironment(token, code, serverPasswd)
	ptoken_succ:=core.ProcessTokenDefault(env)
	//init_and_run_client(token, code, serverPasswd)
	if !ptoken_succ {
		panic("Failed to load token")
	}
	core.InitClient(env)
	go core.EnterReadlineThread(env,nil)
	defer core.DestroyClient(env)
	core.EnterWorkerThread(env,nil)
}

func init_and_run_debug_client() {
	env := core.InitDebugEnvironment()
	core.InitClient(env)
	go core.EnterReadlineThread(env,nil)
	defer core.DestroyClient(env)
	core.EnterWorkerThread(env,nil)
}

func getInput() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	inp, err := reader.ReadString('\n')
	inpl := strings.TrimRight(inp, "\r\n")
	return inpl, err
}

func getInputUserName() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	pterm.Printf(I18n.T(I18n.Enter_FBUC_Username))
	fbusername, err := reader.ReadString('\n')
	return fbusername, err
}

func getRentalServerCode() (string, string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf(I18n.T(I18n.Enter_Rental_Server_Code))
	code, err := reader.ReadString('\n')
	if err != nil {
		return "", "", err
	}
	fmt.Printf(I18n.T(I18n.Enter_Rental_Server_Password))
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Printf("\n")
	return strings.TrimRight(code, "\r\n"), string(bytePassword), err
}

func readToken(path string) (string, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func loadTokenPath() string {
	homedir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(I18n.T(I18n.Warning_UserHomeDir))
		homedir = "."
	}
	fbconfigdir := filepath.Join(homedir, ".config/fastbuilder")
	os.MkdirAll(fbconfigdir, 0700)
	token := filepath.Join(fbconfigdir, "fbtoken")
	return token
}
