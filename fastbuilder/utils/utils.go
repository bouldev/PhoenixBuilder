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
)

// int compareVersion(char *latestVersion,char *currentVersion);
import "C"

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
	libre_regexp:=regexp.MustCompile("^(v\\d+\\.\\d+\\.\\d+)-libre$")
	for _, _ver := range json_structure {
		ver:=_ver.(map[string]interface{})
		if ver["draft"].(bool) || !ver["prerelease"].(bool) {
			continue
		}
		regexp_res:=libre_regexp.FindAllStringSubmatch(ver["tag_name"].(string), -1)
		if len(regexp_res)==0||len(regexp_res[0])!=2 {
			continue
		}
		version:=regexp_res[0][1]
		return C.compareVersion(C.CString(version[1:]),C.CString(currentVersion))!=0, version[1:]
	}
	return false, ""
}