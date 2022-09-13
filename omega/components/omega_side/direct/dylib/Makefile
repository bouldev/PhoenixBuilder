MACOS_AMD64_CC:=/usr/bin/clang 
MACOS_ARM64_CC:=/usr/bin/clang
WINDOWS_AMD64_CC:=/opt/homebrew/bin/x86_64-w64-mingw32-gcc
WINDOWS_X86_CC:=/opt/homebrew/bin/i686-w64-mingw32-gcc
LINUX_AMD64_CC:=/opt/homebrew/bin/x86_64-unknown-linux-gnu-gcc
LINUX_ARM64_CC:=/opt/homebrew/bin/aarch64-unknown-linux-gnu-gcc
ANDROID_NDK_HOME:=$(shell brew --prefix)/share/android-ndk
ANDROID_ARM64_CC:=${ANDROID_NDK_HOME}/toolchains/llvm/prebuilt/darwin-x86_64/bin/aarch64-linux-android21-clang

SRCS_GO := $(foreach dir, $(shell find .. -type d), $(wildcard $(dir)/*.go $(dir)/*.c))
CGO_DEF := ""
GO_CGO_FLAGS_COMMON :=CGO_CFLAGS=${CGO_DEF} CGO_ENABLED=1
GO_BUILD_FLAGS_COMMON := -tags fbconn -buildmode=c-shared -trimpath -ldflags "-s -w"

RELEASE_DIR:=libs
OUTPUT_DIR:=${RELEASE_DIR}
EXEC_PREFIX:=fbconn_

TYPE:=linux_amd64
${TYPE}_LIB:=${OUTPUT_DIR}/${EXEC_PREFIX}${TYPE}.so
${${TYPE}_LIB}_CC:=${LINUX_AMD64_CC}
${${TYPE}_LIB}_TRIPLE:=GOOS=linux GOARCH=amd64
${TYPE}: ${${TYPE}_LIB}
LIBS:=${LIBS} ${${TYPE}_LIB}

TYPE:=linux_arm64
${TYPE}_LIB:=${OUTPUT_DIR}/${EXEC_PREFIX}${TYPE}.so
${${TYPE}_LIB}_CC:=${LINUX_ARM64_CC}
${${TYPE}_LIB}_TRIPLE:=GOOS=linux GOARCH=arm64
${TYPE}: ${${TYPE}_LIB}
LIBS:=${LIBS} ${${TYPE}_LIB}

TYPE:=windows_x86
${TYPE}_LIB:=${OUTPUT_DIR}/${EXEC_PREFIX}${TYPE}.dll
${${TYPE}_LIB}_CC:=${WINDOWS_X86_CC}
${${TYPE}_LIB}_TRIPLE:=GOOS=windows GOARCH=386
${TYPE}: ${${TYPE}_LIB}
LIBS:=${LIBS} ${${TYPE}_LIB}

TYPE:=windows_amd64
${TYPE}_LIB:=${OUTPUT_DIR}/${EXEC_PREFIX}${TYPE}.dll
${${TYPE}_LIB}_CC:=${WINDOWS_AMD64_CC}
${${TYPE}_LIB}_TRIPLE:=GOOS=windows GOARCH=amd64
${TYPE}: ${${TYPE}_LIB}
LIBS:=${LIBS} ${${TYPE}_LIB}

TYPE:=macos_amd64
${TYPE}_LIB:=${OUTPUT_DIR}/${EXEC_PREFIX}${TYPE}.dylib
${${TYPE}_LIB}_CC:=${MACOS_AMD64_CC}
${${TYPE}_LIB}_TRIPLE:=GOOS=darwin GOARCH=amd64
${TYPE}: ${${TYPE}_LIB}
LIBS:=${LIBS} ${${TYPE}_LIB}

TYPE:=macos_arm64
${TYPE}_LIB:=${OUTPUT_DIR}/${EXEC_PREFIX}${TYPE}.dylib
${${TYPE}_LIB}_CC:=${MACOS_ARM64_CC}
${${TYPE}_LIB}_TRIPLE:=GOOS=darwin GOARCH=arm64
${TYPE}: ${${TYPE}_LIB}
LIBS:=${LIBS} ${${TYPE}_LIB}

TYPE:=android_arm64
${TYPE}_LIB:=${OUTPUT_DIR}/${EXEC_PREFIX}${TYPE}
${${TYPE}_LIB}_CC:=${ANDROID_ARM64_CC}
${${TYPE}_LIB}_TRIPLE:=GOOS=android GOARCH=arm64
${TYPE}: ${${TYPE}_LIB}
LIBS:=${LIBS} ${${TYPE}_LIB}

${OUTPUT_DIR}:
	@echo make output dir $@
	@mkdir -p $@
	

.PHONY: ${LIBS}
${LIBS}: ${OUTPUT_DIR}/${EXEC_PREFIX}%: ${OUTPUT_DIR} ${SRCS_GO}
	${GO_CGO_FLAGS_COMMON} ${$@_TRIPLE} CC=${$@_CC}  go build ${GO_BUILD_FLAGS_COMMON} -o $@ main.go
	@echo "\033[32mbuild $@ Done \033[0m\t"

all: ${LIBS}

# upload: 
# 	mv ./binary/fastbuilder-windows.exe.hash ./binary/fastbuilder-windows.hash 
# 	scp -r ./binary/* FBOmega:/var/www/omega/binary/
# 	omega_release/file_uploader/main -d ./binary -s binary/ -c omega_release/file_uploader/config.json
# 	mv ./binary/fastbuilder-windows.hash ./binary/fastbuilder-windows.exe.hash

clean:
	rm -f ${OUTPUT_DIR}/*