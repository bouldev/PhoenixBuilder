SRCS_GO := $(foreach dir, $(shell find . -type d), $(wildcard $(dir)/*.go $(dir)/*.c))

MACOS_AMD64_CC:=/usr/bin/clang 
MACOS_ARM64_CC:=/usr/bin/clang
WINDOWS_AMD64_CC:=/opt/homebrew/bin/x86_64-w64-mingw32-gcc
LINUX_AMD64_CC:=/opt/homebrew/bin/x86_64-unknown-linux-gnu-gcc

ANDROID_NDK_HOME:=$(shell brew --prefix)/share/android-ndk
ANDROID_ARM64_CC:=${ANDROID_NDK_HOME}/toolchains/llvm/prebuilt/darwin-x86_64/bin/aarch64-linux-android21-clang

LINUX_AMD64_EXEC:=build/launcher-linux
MACOS_AMD64_EXEC:=build/launcher-macos
WINDOWS_AMD64_EXEC:=build/launcher-windows.exe
ANDROID_ARM64_EXEC:=build/launcher-android
LINUX_MCSM_AMD64_EXEC:=build/launcher-linux-mcsm

GO_CGO_FLAGS_COMMON :=CGO_CFLAGS=${CGO_DEF} CGO_ENABLED=1
GO_BUILD_FLAGS_COMMON :=-trimpath -ldflags "-s -w"

all: ${LINUX_AMD64_EXEC} ${MACOS_AMD64_EXEC} ${WINDOWS_AMD64_EXEC} ${ANDROID_ARM64_EXEC} ${LINUX_MCSM_AMD64_EXEC}

build/:
	mkdir build

.PHONY: ${LINUX_AMD64_EXEC}
${LINUX_AMD64_EXEC}: ${LINUX_AMD64_CC} build/ ${SRCS_GO}
	@echo build $@
	@${GO_CGO_FLAGS_COMMON} GOOS=linux GOARCH=amd64 CC=$< go build ${GO_BUILD_FLAGS_COMMON} -o $@

.PHONY: ${LINUX_MCSM_AMD64_EXEC}
${LINUX_MCSM_AMD64_EXEC}: ${LINUX_AMD64_CC} build/ ${SRCS_GO}
	@echo build $@
	@${GO_CGO_FLAGS_COMMON} GOOS=linux GOARCH=amd64 CC=$< go build ${GO_BUILD_FLAGS_COMMON} -tags=mcsm -o $@

.PHONY: ${WINDOWS_AMD64_EXEC}
${WINDOWS_AMD64_EXEC}: ${WINDOWS_AMD64_CC} build/ ${SRCS_GO}
	@echo build $@
	@${GO_CGO_FLAGS_COMMON} GOOS=windows GOARCH=amd64 CC=$< go build ${GO_BUILD_FLAGS_COMMON} -o $@

.PHONY: ${MACOS_AMD64_EXEC}
${MACOS_AMD64_EXEC}: ${MACOS_AMD64_CC} build/ ${SRCS_GO}
	@echo build $@
	@${GO_CGO_FLAGS_COMMON} GOOS=darwin GOARCH=amd64 CC=$< go build ${GO_BUILD_FLAGS_COMMON} -o $@

.PHONY: ${MACOS_ARM64_EXEC}
${MACOS_ARM64_EXEC}: ${MACOS_ARM64_CC} build/ ${SRCS_GO}
	@echo build $@
	@${GO_CGO_FLAGS_COMMON} GOOS=darwin GOARCH=arm64 CC=$< go build ${GO_BUILD_FLAGS_COMMON} -o $@

.PHONY: ${ANDROID_ARM64_EXEC}
${ANDROID_ARM64_EXEC}: ${ANDROID_ARM64_CC} build/ ${SRCS_GO}
	@echo build $@
	@${GO_CGO_FLAGS_COMMON} GOOS=android GOARCH=arm64 CC=$< go build ${GO_BUILD_FLAGS_COMMON} -o $@

clean:
	rm -rf build/*
