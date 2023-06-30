package utils

import (
	"bytes"
	"fmt"

	"golang.org/x/text/encoding/simplifiedchinese"
)

// Better to just drop GBK support

func IsGBK(data []byte) (isGBK bool, failAtPos int) {
	length := len(data)
	var i int = 0
	for i < length {
		if data[i] <= 0x7f {
			i++
			continue
		} else {
			if data[i] >= 0x81 &&
				data[i] <= 0xfe &&
				data[i+1] >= 0x40 &&
				data[i+1] <= 0xfe &&
				data[i+1] != 0xf7 {
				i += 2
				continue
			} else {
				return false, i
			}
		}
	}
	return true, 0
}

func preNUm(data byte) int {
	var mask byte = 0x80
	var num int = 0
	for i := 0; i < 8; i++ {
		if (data & mask) == mask {
			num++
			mask = mask >> 1
		} else {
			break
		}
	}
	return num
}
func IsUtf8(data []byte) (isUtf8 bool, failAtPos int) {
	i := 0
	for i < len(data) {
		if (data[i] & 0x80) == 0x00 {
			// 0XXX_XXXX
			i++
			continue
		} else if num := preNUm(data[i]); num > 2 {
			// 110X_XXXX 10XX_XXXX
			// 1110_XXXX 10XX_XXXX 10XX_XXXX
			// 1111_0XXX 10XX_XXXX 10XX_XXXX 10XX_XXXX
			// 1111_10XX 10XX_XXXX 10XX_XXXX 10XX_XXXX 10XX_XXXX
			// 1111_110X 10XX_XXXX 10XX_XXXX 10XX_XXXX 10XX_XXXX 10XX_XXXX
			i++
			for j := 0; j < num-1; j++ {
				if (data[i] & 0xc0) != 0x80 {
					return false, i
				}
				i++
			}
		} else {
			fmt.Println(i)
			return false, i
		}
	}
	return true, 0
}

func GBKTextToUTF8Text(gbkText []byte) (utf8Text []byte, err error) {
	return simplifiedchinese.GBK.NewDecoder().Bytes(gbkText)
}

type Encoding int

const (
	UNKNOWN = Encoding(iota)
	UTF8
	GBK
)

func GetStrCoding(data []byte) Encoding {
	fallBackEncoding := UNKNOWN
	fallBackDecodeLen := 0
	if isUtf8, failAt := IsUtf8(data); isUtf8 {
		return UTF8
	} else {
		if failAt > fallBackDecodeLen {
			fallBackDecodeLen = failAt
			fallBackEncoding = UTF8
		}
	}
	if isGBK, failAt := IsGBK(data); isGBK {
		return GBK
	} else {
		if failAt > fallBackDecodeLen {
			fallBackDecodeLen = failAt
			fallBackEncoding = GBK
		}
	}
	if fallBackDecodeLen > 10 {
		return fallBackEncoding
	} else {
		return UNKNOWN
	}
}

func AutoConvertTextToUtf8(data []byte) (utf8Data []byte, err error) {
	data = bytes.Trim(data, "\xef\xbb\xbf")
	switch GetStrCoding(data) {
	case UTF8:
		return data, nil
	case GBK:
		return GBKTextToUTF8Text(data)
	}
	return data, fmt.Errorf("unknown encoding")
}
