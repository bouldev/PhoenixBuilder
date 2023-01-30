package Happy2018new_depends

import (
	"bytes"
	"io"

	"github.com/andybalholm/brotli"
)

func Compress(input []byte) []byte {
	var ans bytes.Buffer
	compressor := brotli.NewWriter(&ans)
	compressor.Write(input)
	compressor.Close()
	return ans.Bytes()
}

func Decompress(input []byte) []byte {
	var output bytes.Buffer
	var in bytes.Buffer
	in.Write(input)
	reader := brotli.NewReader(&in)
	io.Copy(&output, reader)
	return output.Bytes()
}
