package utils

import (
	"crypto/md5"
	"fmt"
)

func GetBinaryHash(fileData []byte) string {
	cvt := func(in [16]byte) []byte {
		return in[:16]
	}
	hashedBytes := cvt(md5.Sum(fileData))
	return fmt.Sprintf("%x\n", hashedBytes)
}
