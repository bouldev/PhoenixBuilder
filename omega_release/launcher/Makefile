.PHONY: all current 
TARGETS:=build/ current current-v8
PACKAGETARGETS:=
ifeq ($(shell uname | grep "Darwin" > /dev/null ; echo $${?}),0)
	ifeq ($(shell uname -m | grep -E "iPhone|iPad|iPod" > /dev/null ; echo $${?}),0)
		IOS_STRIP=/usr/bin/strip
		LIPO=/usr/bin/lipo
		LDID=/usr/bin/ldid
		TARGETS:=${TARGETS} ios-executable ios-v8-executable ios-lib
	else
		IOS_STRIP=$(shell xcrun --sdk iphoneos -f strip)
		IOS_OBJCOPY=$(shell xcrun --sdk iphoneos -f objcopy)
		LDID=ldid2
		LIPO=/usr/bin/lipo
		TARGETS:=${TARGETS} macos ios-v8-executable ios-executable ios-lib
	endif
		PACKAGETARGETS:=${PACKAGETARGETS} package/ios
else
	IOS_STRIP=true
	LDID=$${THEOS}/toolchain/linux/iphone/bin/ldid
	LIPO=$${THEOS}/toolchain/linux/iphone/bin/lipo
	IOS_OBJCOPY=$${THEOS}/toolchain/linux/iphone/bin/llvm-objcopy
endif 

SRCS_GO := $(foreach dir, $(shell find . -type d), $(wildcard $(dir)/*.go $(dir)/*.c))

all: linux macos windows mcsm-linux mcsm-windows android
linux: build/launcher-linux
macos: build/launcher-macos
windows: build/launcher-windows.exe
android: build/launcher-android
mcsm-linux: build/launcher-linux-mcsm
mcsm-windows: build/launcher-windows-mcsm

package: ${PACKAGETARGETS}
release/:
	mkdir -p release
build/:
	mkdir build
build/launcher-linux: build/ ${SRCS_GO}
	CGO_CFLAGS=${CGO_DEF} CGO_ENABLED=1  go build -trimpath -ldflags "-s -w" -o $@
	cp build/launcher-linux build/Linux版Omega启动器
build/launcher-linux-mcsm:
	CGO_CFLAGS=${CGO_DEF} CGO_ENABLED=1  go build -trimpath -tags="mcsm" -ldflags "-s -w" -o $@
build/launcher-windows-mcsm:
	CGO_CFLAGS=${CGO_DEF} CC=/usr/bin/x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build -trimpath -tags="mcsm" -ldflags "-s -w" -o $@
	mv $@ build/Windows版MCSM特供Omega启动器.exe
build/launcher-macos: build/ ${SRCS_GO}
	CGO_CFLAGS=${CGO_DEF} CC=`pwd`/macos CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o $@
	cp $@ build/MacOS版Omega启动器
build/launcher-windows.exe: build/ /usr/bin/x86_64-w64-mingw32-gcc ${SRCS_GO}
	CGO_CFLAGS=${CGO_DEF} CC=/usr/bin/x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build -trimpath -ldflags "-s -w" -o $@
	mv $@ build/Windows版Omega启动器.exe
build/launcher-android: build/ ${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/aarch64-linux-android21-clang ${SRCS_GO}
	CGO_CFLAGS=${CGO_DEF} CC=${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/aarch64-linux-android21-clang CXX=${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/aarch64-linux-android21-clang++ GOOS=android GOARCH=arm64 CGO_ENABLED=1 go build -trimpath -ldflags "-s -w -extldflags -static-libstdc++" -o $@
clean:
	rm -rf build/*
