package script_engine

import (
	"crypto/sha256"
	"encoding/base64"
	"phoenixbuilder/fastbuilder/configuration"
)

func GetStringSha(data string) string {
	h := sha256.New()
	h.Write([]byte(configuration.UserToken))
	return base64.RawStdEncoding.EncodeToString(h.Sum(nil))
}
