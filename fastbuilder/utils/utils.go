package utils

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"fmt"
	"net/http"
	"strconv"
	"strings"
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
	resp, err:=http.Get("https://storage.fastbuilder.pro/version")
	if(err!=nil) {
		fmt.Printf("Failed to check update!\nPlease check your network status.\n")
		return false, ""
	}
	content, err:=io.ReadAll(resp.Body)
	if(err!=nil) {
		fmt.Printf("Failed to check update!\nPlease check your network status.\n")
		return false, ""
	}
	return C.compareVersion(C.CString(string(content)),C.CString(currentVersion))!=0, strings.Replace(string(content),"\n","",1)
}