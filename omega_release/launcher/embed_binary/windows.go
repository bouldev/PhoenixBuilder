//go:build windows
// +build windows

package embed_binary

import (
	_ "embed"
)

//go:embed cqhttp-windows.brotli
var embedding_cqhttp []byte
var PLANTFORM = WINDOWS_x86_64
