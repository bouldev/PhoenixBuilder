.PHONY: all clean current ios-executable ios-lib macos android-executable-v7 android-executable-64 windows-executable
TARGETS:=build current
ifeq ($(shell uname | grep "Darwin" > /dev/null ; echo $${?}),0)
ifeq ($(shell uname -m | grep -E "iPhone|iPad|iPod" > /dev/null ; echo $${?}),0)
IOS_STRIP=/usr/bin/strip
LDID=/usr/bin/ldid
else
IOS_STRIP=$(shell xcrun --sdk iphoneos -f strip)
LDID=ldid2
TARGETS:=${TARGETS} macos
endif
else
IOS_STRIP=true
LDID=$${THEOS}/toolchain/linux/iphone/bin/ldid
endif
ifneq (${THEOS},)
	TARGETS:=${TARGETS} ios-executable ios-lib
endif
ifneq ($(wildcard ${HOME}/android-ndk-r20b),)
	TARGETS:=${TARGETS} android-executable-v7 android-executable-64
endif
ifneq ($(wildcard /usr/bin/i686-w64-mingw32-gcc),)
	TARGETS:=${TARGETS} windows-executable
endif
all: ${TARGETS} build/hashes.json
#all: build current ios-executable ios-lib macos android-executable-v7 android-executable-64 windows-executable
current: build/phoenixbuilder
ios-executable: build/phoenixbuilder-ios-executable
ios-lib: build/phoenixbuilder-ios-lib.a
macos: build/phoenixbuilder-macos
android-executable-v7: build/phoenixbuilder-android-executable-armv7
android-executable-64: build/phoenixbuilder-android-executable-arm64
windows-executable: build/phoenixbuilder-windows-executable.exe
build:
	mkdir build
build/phoenixbuilder: build
	CGO_ENABLED=1 go build -trimpath -ldflags "-s -w" -o build/phoenixbuilder
build/phoenixbuilder-ios-executable: build
	CC=`pwd`/archs/ios.sh CGO_ENABLED=1 GOOS=ios GOARCH=arm64 go build -trimpath -ldflags "-s -w" -o build/phoenixbuilder-ios-executable
	${IOS_STRIP} build/phoenixbuilder-ios-executable
	${LDID} -S build/phoenixbuilder-ios-executable
build/phoenixbuilder-ios-lib.a: build
	CC=`pwd`/archs/ios.sh CGO_ENABLED=1 GOOS=ios GOARCH=arm64 go build -buildmode=c-archive -trimpath -ldflags "-s -w" -o build/phoenixbuilder-ios-static.a
build/phoenixbuilder-macos-x86_64: build
	CC=`pwd`/archs/macos.sh CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o build/phoenixbuilder-macos-x86_64
build/phoenixbuilder-macos-arm64: build
	CC=`pwd`/archs/macos.sh CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags "-s -w" -o build/phoenixbuilder-macos-arm64
build/phoenixbuilder-macos: build build/phoenixbuilder-macos-x86_64 build/phoenixbuilder-macos-arm64
	lipo -create build/phoenixbuilder-macos-x86_64 build/phoenixbuilder-macos-arm64 -output build/phoenixbuilder-macos
build/phoenixbuilder-android-executable-armv7: build ${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/armv7a-linux-androideabi16-clang
	CC=${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/armv7a-linux-androideabi16-clang GOOS=android GOARCH=arm CGO_ENABLED=1 go build -trimpath -ldflags "-s -w" -o build/phoenixbuilder-android-executable-armv7
build/phoenixbuilder-android-executable-arm64: build ${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/aarch64-linux-android21-clang
	CC=${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/aarch64-linux-android21-clang GOOS=android GOARCH=arm64 CGO_ENABLED=1 go build -trimpath -ldflags "-s -w" -o build/phoenixbuilder-android-executable-arm64
build/phoenixbuilder-windows-executable.exe: build /usr/bin/i686-w64-mingw32-gcc
	CC=/usr/bin/i686-w64-mingw32-gcc GOOS=windows GOARCH=386 CGO_ENABLED=1 go build -trimpath -ldflags "-s -w" -o build/phoenixbuilder-windows-executable.exe
build/hashes.json: build genhash.js
	node genhash.js
clean:
	rm -f build/phoenixbuilder*