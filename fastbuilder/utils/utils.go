package utils

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"fmt"
	"net/http"
	"strconv"
	"regexp"
	"golang.org/x/term"
	"strings"
	I18n "phoenixbuilder/fastbuilder/i18n"
	"path/filepath"
	"io/ioutil"
	"bufio"
	"syscall"
)

func SliceAtoi(sa []string) ([]int, error) {
	si := make([]int, 0, len(sa))
	for _, a := range sa {
		i, err := strconv.Atoi(a)
		if err != nil {
			return si, err
		}
		si = append(si, i)
	}
	return si, nil
}

func GetHash(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func GetMD5(i string) string {
	sum:=md5.Sum([]byte(i))
	return hex.EncodeToString(sum[:])
}

func CheckUpdate(currentVersion string) (bool, string) {
	libre_regexp:=regexp.MustCompile("^v?((\\d+)\\.(\\d+)\\.(\\d+))-libre$")
	current_version_reg:=libre_regexp.FindAllStringSubmatch(currentVersion, -1)
	if len(current_version_reg)==0||len(current_version_reg[0])!=5 {
		return false, ""
	}
	// ^ !libre_regexp.MatchString(currentVersion)
	current_major_version, _:=strconv.Atoi(current_version_reg[0][2])
	current_minor_version, _:=strconv.Atoi(current_version_reg[0][3])
	current_patch_version, _:=strconv.Atoi(current_version_reg[0][4])
	resp, err:=http.Get("https://api.github.com/repos/LNSSPsd/PhoenixBuilder/releases")
	if(err!=nil) {
		fmt.Printf("Failed to check update!\nPlease check your network status.\n")
		return false, ""
	}
	content, err:=io.ReadAll(resp.Body)
	if(err!=nil) {
		fmt.Printf("Failed to check update!\nPlease check your network status.\n")
		return false, ""
	}
	var json_structure []interface{}
	err=json.Unmarshal(content, &json_structure)
	if err!=nil {
		fmt.Printf("Failed to check update due to invalid response received from GitHub.\n")
		return false, ""
	}
	for _, _ver := range json_structure {
		ver:=_ver.(map[string]interface{})
		if ver["draft"].(bool) || !ver["prerelease"].(bool) {
			continue
		}
		regexp_res:=libre_regexp.FindAllStringSubmatch(ver["tag_name"].(string), -1)
		if len(regexp_res)==0||len(regexp_res[0])!=5 {
			continue
		}
		latest_major_version, _:=strconv.Atoi(regexp_res[0][2])
		latest_minor_version, _:=strconv.Atoi(regexp_res[0][3])
		latest_patch_version, _:=strconv.Atoi(regexp_res[0][4])
		if(current_major_version<latest_major_version) {
			return true, regexp_res[0][1]
		}else if(current_major_version==latest_major_version) {
			if(current_minor_version<latest_minor_version) {
				return true, regexp_res[0][1]
			}else if(current_minor_version==latest_minor_version&&current_patch_version<latest_patch_version) {
				return true, regexp_res[0][1]
			}
		}
		break
	}
	return false, ""
}

func LoadTokenPath() string {
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

func ReadToken(path string) (string, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func GetUserInput(prompt string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	input, err := reader.ReadString('\n')
	return strings.TrimSpace(input), err
}

func GetUserPasswordInput(prompt string) (string, error) {
	fmt.Print(prompt)
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Printf("\n")
	return strings.TrimSpace(string(bytePassword)), err
}

func GetRentalServerCode() (string, string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(I18n.T(I18n.Enter_Rental_Server_Code))
	code, err := reader.ReadString('\n')
	if err != nil {
		return "", "", err
	}
	fmt.Print(I18n.T(I18n.Enter_Rental_Server_Password))
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Printf("\n")
	return strings.TrimRight(code, "\r\n"), strings.TrimSpace(string(bytePassword)), err
}

func WriteFBToken(token string, tokenPath string) {
	if fp, err := os.Create(tokenPath); err != nil {
		fmt.Println(I18n.T(I18n.FBUC_Token_ErrOnCreate), err)
		fmt.Println(I18n.T(I18n.ErrorIgnored))
	} else {
		_, err = fp.WriteString(token)
		if err != nil {
			fmt.Println(I18n.T(I18n.FBUC_Token_ErrOnSave), err)
			fmt.Println(I18n.T(I18n.ErrorIgnored))
		}
		fp.Close()
	}
}

func ReadUserInfo(userName, userPassword, userToken, serverCode, serverPassword string) (string, string, string, string, string, error) {
	var err error
	// read token or get user input
	I18n.Init()
	if userName == "" && userPassword == "" && userToken == "" {
		userToken, err = ReadToken(LoadTokenPath())
		if err != nil || userToken == "" {
			for userName == "" {
				userName, err = GetUserInput("请输入 FB 用户名或者 Token:")
				if strings.HasPrefix(userName, "w9/") {
					userToken = userName
					userName = ""
					break
				}
				if err != nil {
					return userName, userPassword, userToken, serverCode, serverPassword, err
				}
			}
			if userToken == "" {
				for userPassword == "" {
					userPassword, err = GetUserPasswordInput(I18n.T(I18n.EnterPasswordForFBUC))
					if err != nil {
						return userName, userPassword, userToken, serverCode, serverPassword, err
					}
				}
			}
		}
	}

	// read server code and password
	if serverCode == "" {
		serverCode, serverPassword, err = GetRentalServerCode()
		if err != nil {
			return userName, userPassword, userToken, serverCode, serverPassword, err
		}
	}
	return userName, userPassword, userToken, serverCode, serverPassword, nil
}