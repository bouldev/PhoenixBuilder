package utils

import (
	"bufio"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	I18n "phoenixbuilder/fastbuilder/i18n"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/term"
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
	sum := md5.Sum([]byte(i))
	return hex.EncodeToString(sum[:])
}

func CheckUpdate(currentVersion string) (bool, string) {
	libre_regexp := regexp.MustCompile(`^v?((\d+)\.(\d+)\.(\d+))-libre$`)
	current_version_reg := libre_regexp.FindAllStringSubmatch(currentVersion, -1)
	if len(current_version_reg) == 0 || len(current_version_reg[0]) != 5 {
		return false, ""
	}
	// ^ !libre_regexp.MatchString(currentVersion)
	current_major_version, _ := strconv.Atoi(current_version_reg[0][2])
	current_minor_version, _ := strconv.Atoi(current_version_reg[0][3])
	current_patch_version, _ := strconv.Atoi(current_version_reg[0][4])
	resp, err := http.Get("https://api.github.com/repos/LNSSPsd/PhoenixBuilder/releases")
	if err != nil {
		fmt.Printf("Failed to check update!\nPlease check your network status.\n")
		return false, ""
	}
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Failed to check update!\nPlease check your network status.\n")
		return false, ""
	}
	var json_structure []interface{}
	err = json.Unmarshal(content, &json_structure)
	if err != nil {
		fmt.Printf("Failed to check update due to invalid response received from GitHub.\n")
		return false, ""
	}
	for _, _ver := range json_structure {
		ver := _ver.(map[string]interface{})
		if ver["draft"].(bool) || !ver["prerelease"].(bool) {
			continue
		}
		regexp_res := libre_regexp.FindAllStringSubmatch(ver["tag_name"].(string), -1)
		if len(regexp_res) == 0 || len(regexp_res[0]) != 5 {
			continue
		}
		latest_major_version, _ := strconv.Atoi(regexp_res[0][2])
		latest_minor_version, _ := strconv.Atoi(regexp_res[0][3])
		latest_patch_version, _ := strconv.Atoi(regexp_res[0][4])
		if current_major_version < latest_major_version {
			return true, regexp_res[0][1]
		} else if current_major_version == latest_major_version {
			if current_minor_version < latest_minor_version {
				return true, regexp_res[0][1]
			} else if current_minor_version == latest_minor_version && current_patch_version < latest_patch_version {
				return true, regexp_res[0][1]
			}
		}
		break
	}
	return false, ""
}

func GetRentalServerCode() (string, string, error) {
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

func GetUsernameInput() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf(I18n.T(I18n.Enter_FBUC_Username))
	fbusername, err := reader.ReadString('\n')
	return fbusername, err
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
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
