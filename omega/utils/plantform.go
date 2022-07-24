package utils

import "runtime"

type PREPARED_PLATFORM_MARK int

const (
	// for some specific platform, it could be useful if we provide some specific data/information, so let's us mark them first
	PLATFORM_ALL          = PREPARED_PLATFORM_MARK(-1)
	PLATFORM_NOT_PROVIDED = PREPARED_PLATFORM_MARK(iota)
	PLATFORM_LINUX_AMD64
	PLATFORM_LINUX_ARM64
	PLATFORM_MACOS_AMD64
	PLATFORM_MACOS_ARM64
	PLATFORM_ANDROID_ARM64
	PLATFORM_WINDOWS_AMD64
	PLATFORM_WINDOWS_X86
)

var PLATFORM_MARK_FOR_PREPARED = PLATFORM_NOT_PROVIDED
var PLATFORM_NAME_STR = runtime.GOOS + "_" + runtime.GOARCH

func init() {
	platformNameMapping := map[string]PREPARED_PLATFORM_MARK{
		"linux_amd64":   PLATFORM_LINUX_AMD64,
		"linux_arm64":   PLATFORM_LINUX_ARM64,
		"darwin_amd64":  PLATFORM_MACOS_AMD64,
		"darwin_arm64":  PLATFORM_MACOS_ARM64,
		"android_arm64": PLATFORM_ANDROID_ARM64,
		"windows_amd64": PLATFORM_WINDOWS_AMD64,
		"windows_386":   PLATFORM_WINDOWS_X86,
	}
	if mark, hasK := platformNameMapping[PLATFORM_NAME_STR]; hasK {
		PLATFORM_MARK_FOR_PREPARED = mark
	}
}
