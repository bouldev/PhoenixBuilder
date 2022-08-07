SRCS_GO := $(foreach dir, $(shell find . -type d), $(wildcard $(dir)/*.go $(dir)/*.c))

MACOS_AMD64_CC:=/usr/bin/clang 
MACOS_ARM64_CC:=/usr/bin/clang
WINDOWS_AMD64_CC:=/opt/homebrew/bin/x86_64-w64-mingw32-gcc
LINUX_AMD64_CC:=/opt/homebrew/bin/x86_64-unknown-linux-gnu-gcc

ANDROID_NDK_HOME:=$(shell brew --prefix)/share/android-ndk
ANDROID_ARM64_CC:=${ANDROID_NDK_HOME}/toolchains/llvm/prebuilt/darwin-x86_64/bin/aarch64-linux-android21-clang

GO_CGO_FLAGS_COMMON :=CGO_ENABLED=1
GO_BUILD_FLAGS_COMMON :=-trimpath -ldflags "-s -w"

RELEASE_DIR:=build
OUTPUT_DIR:=${RELEASE_DIR}
EXEC_PREFIX:=launcher-
MCSM_TAG=-tags=mcsm

TYPE:=linux
${TYPE}_EXEC:=${OUTPUT_DIR}/${EXEC_PREFIX}${TYPE}
${${TYPE}_EXEC}_CC:=${LINUX_AMD64_CC}
${${TYPE}_EXEC}_TRIPLE:=GOOS=linux GOARCH=amd64
${TYPE}: ${${TYPE}_EXEC}
EXECS:=${EXECS} ${${TYPE}_EXEC}

TYPE:=linux-mcsm
${TYPE}_EXEC:=${OUTPUT_DIR}/${EXEC_PREFIX}${TYPE}
${${TYPE}_EXEC}_CC:=${LINUX_AMD64_CC}
${${TYPE}_EXEC}_TRIPLE:=GOOS=linux GOARCH=amd64
${${TYPE}_EXEC}_MCSM_TAG:=${MCSM_TAG}
${TYPE}: ${${TYPE}_EXEC}
EXECS:=${EXECS} ${${TYPE}_EXEC}

TYPE:=windows
${TYPE}_EXEC:=${OUTPUT_DIR}/${EXEC_PREFIX}${TYPE}.exe
${TYPE}: ${${TYPE}_EXEC}
${${TYPE}_EXEC}_CC:=${WINDOWS_AMD64_CC}
${${TYPE}_EXEC}_TRIPLE:=GOOS=windows GOARCH=amd64
EXECS:=${EXECS} ${${TYPE}_EXEC}

TYPE:=macos
${TYPE}_EXEC:=${OUTPUT_DIR}/${EXEC_PREFIX}${TYPE}
${TYPE}: ${${TYPE}_EXEC}
${${TYPE}_EXEC}_CC:=${MACOS_AMD64_CC}
${${TYPE}_EXEC}_TRIPLE:=GOOS=darwin GOARCH=amd64
EXECS:=${EXECS} ${${TYPE}_EXEC}

TYPE:=android
${TYPE}_EXEC:=${OUTPUT_DIR}/${EXEC_PREFIX}${TYPE}
${TYPE}: ${${TYPE}_EXEC}
${${TYPE}_EXEC}_CC:=${ANDROID_ARM64_CC}
${${TYPE}_EXEC}_TRIPLE:=GOOS=android GOARCH=arm64
EXECS:=${EXECS} ${${TYPE}_EXEC}

${OUTPUT_DIR}:
	@echo make output dir $@
	@mkdir -p $@

clean:
	rm -f ${OUTPUT_DIR}/*

.PHONY: ${EXECS}

all: ${EXECS}

${EXECS}: ${OUTPUT_DIR}/${EXEC_PREFIX}%: ${OUTPUT_DIR} ${SRCS_GO}
	@${GO_CGO_FLAGS_COMMON} ${$@_TRIPLE} CC=${$@_CC}  go build ${GO_BUILD_FLAGS_COMMON} ${$@_MCSM_TAG} -o $@
	@echo "\033[32mbuild $@ Done \033[0m"
	

