package built_in

import (
	_ "embed"
	"strings"
)

var builtInScripts string

//go:embed aes.js
var jsAES []byte
//go:embed md5.js
var jsMD5 []byte
//go:embed rc4.js
var jsRC4 []byte
//go:embed sha256.js
var jsSHA256 []byte
//go:embed tripledes.js
var jsTripleDes []byte
//go:embed hmac-md5.js
var jsHMAC_MD5 []byte
//go:embed hmac-sha256.js
var jsHMAC_SHA256 []byte

func GetbuiltIn()string{
	if builtInScripts==""{
		builtInScripts=strings.Join([]string{
			string(jsAES),
			string(jsMD5),
			string(jsRC4),
			string(jsSHA256),
			string(jsTripleDes),
			string(jsHMAC_MD5),
			string(jsHMAC_SHA256),
		},"\n")
	}
	return builtInScripts
}