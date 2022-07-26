//go:build linux && !android
// +build linux,!android

package embed_binary

import (
	_ "embed"
)

//go:embed cqhttp-linux.brotli
var embedding_cqhttp []byte
var PLANTFORM = Linux_x86_64
