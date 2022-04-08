package built_in

import (
	_ "embed"
	"strings"
)

var builtInScripts string

//go:embed aes.js
var jsAES string
//go:embed md5.js
var jsMD5 string
//go:embed rc4.js
var jsRC4 string
//go:embed sha256.js
var jsSHA256 string
//go:embed tripledes.js
var jsTripleDes string
//go:embed hmac-md5.js
var jsHMAC_MD5 string
//go:embed hmac-sha256.js
var jsHMAC_SHA256 string
//go:embed pbutils.js
var jsUtils string

func GetbuiltIn()string{
	if builtInScripts==""{
		builtInScripts=strings.Join([]string{
			jsAES,
			jsMD5,
			jsRC4,
			jsSHA256,
			jsTripleDes,
			jsHMAC_MD5,
			jsHMAC_SHA256,
			jsUtils,
		},"\n")
	}
	return builtInScripts
}