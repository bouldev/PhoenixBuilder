package convert_wrapper

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"strings"
)

func StrToHexStringLower(src string) string {
	return hex.EncodeToString([]byte(src))
}

func HexStrToStr(src string) (string, error) {
	dst, err := hex.DecodeString(src)
	return string(dst), err
}

func HexStrToBytes(src string) ([]byte, error) {
	return hex.DecodeString(src)
}

func StrToHexStrUpper(src string) string {
	return strings.ToUpper(hex.EncodeToString([]byte(src)))
}

func StrToBase64EncodeStr(src string) string {
	return base64.StdEncoding.EncodeToString([]byte(src))
}

func StrToMD5HashStr(src string) string {
	h := md5.New()
	h.Write([]byte(src))
	return hex.EncodeToString(h.Sum(nil))
}
