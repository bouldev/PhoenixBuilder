package compress_wrapper

import (
	"bytes"
	"io/ioutil"

	"github.com/andybalholm/brotli"
)

func DecompressBrotli(compressedData []byte) ([]byte, error) {
	return ioutil.ReadAll(brotli.NewReader(bytes.NewReader(compressedData)))
}
