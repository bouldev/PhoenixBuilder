.PHONY: all clean current ios-executable ios-app macos-executable windows-executable android-lib-arm64 android-app windows-executable-x86 windows-executable-x86_64 macos-app
TARGETS:=build/ current
ifeq ($(shell uname | grep "Darwin" > /dev/null ; echo $${?}),0)
ifeq ($(shell uname -m | grep -E "iPhone|iPad|iPod" > /dev/null ; echo $${?}),0)
IOS_STRIP=/usr/bin/strip
LIPO=/usr/bin/lipo
LDID=/usr/bin/ldid
TARGETS:=${TARGETS} ios-executable ios-app
else
IOS_STRIP=$(shell xcrun --sdk iphoneos -f strip)
IOS_OBJCOPY=$(shell xcrun --sdk iphoneos -f objcopy)
LDID=ldid2
LIPO=/usr/bin/lipo
TARGETS:=${TARGETS} macos-executable macos-app ios-executable ios-app
endif
else
IOS_STRIP=true
LDID=$${THEOS}/toolchain/linux/iphone/bin/ldid
LIPO=$${THEOS}/toolchain/linux/iphone/bin/lipo
endif
ifneq (${THEOS},)
	TARGETS:=${TARGETS} ios-executable macos-executable ios-app macos-app
endif
ifneq ($(wildcard ${HOME}/android-ndk-r20b),)
	TARGETS:=${TARGETS} android-lib-arm64 android-app
endif
ifneq ($(wildcard /usr/bin/i686-w64-mingw32-gcc),)
	TARGETS:=${TARGETS} windows-executable-x86
endif
ifneq ($(wildcard /usr/bin/x86_64-w64-mingw32-gcc),)
	TARGETS:=${TARGETS} windows-executable-x86_64
endif
#ifneq ($(wildcard /usr/bin/aarch64-linux-gnu-gcc),)
#	TARGETS:=${TARGETS} current-arm64-executable
#endif

VERSION=$(shell cat version)

SRCS_GO := $(foreach dir, $(shell find . -type d), $(wildcard $(dir)/*.go $(dir)/*.c))

CGO_DEF := "-DFB_VERSION=\"$(VERSION)\" -DFB_COMMIT=\"$(shell git log -1 --format=format:"%h")\" -DFB_COMMIT_LONG=\"$(shell git log -1 --format=format:"%H")\""

all: ${TARGETS}
current: build/phoenixbuilder
ios-executable: build/phoenixbuilder-ios-executable
ios-app: build/phoenixbuilder-ios-app.ipa
macos-executable: build/phoenixbuilder-macos
macos-app: build/phoenixbuilder-macos-app.app.zip
android-app: build/phoenixbuilder-android-app.apk
android-lib-arm64: build/phoenixbuilder-android-arm64-lib.so
windows-executable: windows-executable-x86 windows-executable-x86_64
windows-executable-x86: build/phoenixbuilder-windows-executable-x86.exe
windows-executable-x86_64: build/phoenixbuilder-windows-executable-x86_64.exe

build/:
	mkdir build
build/phoenixbuilder: build/ ${SRCS_GO}
	CGO_CFLAGS=${CGO_DEF} CGO_ENABLED=1  go build -tags fyne_gui -trimpath -ldflags "-s -w" -o build/phoenixbuilder
#build/phoenixbuilder-aarch64: build/ ${SRCS_GO}
#	CGO_CFLAGS=${CGO_DEF} CC=/usr/bin/aarch64-linux-gnu-gcc CGO_ENABLED=1 GOARCH=arm64 go build -tags fyne_gui -trimpath -ldflags "-s -w" -o build/phoenixbuilder-aarch64
build/phoenixbuilder-ios-executable: build/ ${SRCS_GO}
	CGO_CFLAGS=${CGO_DEF} CGO_LDFLAGS="-Wl,-undefined,dynamic_lookup -Wl,-platform_version,2,11.0,14.5" CC=`pwd`/archs/ios.sh CGO_ENABLED=1 GOOS=ios GOARCH=arm64 go build -tags "fyne_gui ios" -trimpath -ldflags "-s -w -extldflags \"-Wl,-platform_version,2,11.0,14.5 platform_helper/assets/isOSVersionAtLeast_provider.m\"" -o build/phoenixbuilder-ios-executable
	${LDID} -Sios-ent.xml build/phoenixbuilder-ios-executable
	${IOS_STRIP} build/phoenixbuilder-ios-executable
build/phoenixbuilder-ios-app.ipa: build/ build/phoenixbuilder-ios-executable
	rm -rf build/ipa
	mkdir -p build/ipa/Payload/phoenixbuilder.app
	sed "s/@VERSION@/${VERSION}/g; s/@COMMIT_COUNT@/`git rev-list --count --all`/g" unbundled_assets/Info.xml > build/ipa/Payload/phoenixbuilder.app/Info.xml
	plistutil -i build/ipa/Payload/phoenixbuilder.app/Info.xml -o build/ipa/Payload/phoenixbuilder.app/Info.plist
	ln build/phoenixbuilder-ios-executable build/ipa/Payload/phoenixbuilder.app/
	ln unbundled_assets/Assets.car build/ipa/Payload/phoenixbuilder.app/
	ln unbundled_assets/empty.mp3 build/ipa/Payload/phoenixbuilder.app/
	rm -f build/phoenixbuilder-ios-app.ipa
	cd build/ipa&&zip -r ../phoenixbuilder-ios-app.ipa Payload&&cd ../..
	rm -rf build/ipa
build/phoenixbuilder-macos-x86_64: build/ ${SRCS_GO}
	CGO_CFLAGS=${CGO_DEF} CC=`pwd`/archs/macos.sh CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -tags fyne_gui -trimpath -ldflags "-s -w" -o build/phoenixbuilder-macos-x86_64
build/phoenixbuilder-macos-arm64: build/ ${SRCS_GO}
	CGO_CFLAGS=${CGO_DEF} CC=`pwd`/archs/macos.sh CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -tags fyne_gui -trimpath -ldflags "-s -w" -o build/phoenixbuilder-macos-arm64
build/phoenixbuilder-macos: build/ build/phoenixbuilder-macos-x86_64 build/phoenixbuilder-macos-arm64 ${SRCS_GO}
	${LIPO} -create build/phoenixbuilder-macos-x86_64 build/phoenixbuilder-macos-arm64 -output build/phoenixbuilder-macos
build/phoenixbuilder-macos-app.app.zip: build/ build/phoenixbuilder-macos
	rm -rf build/phoenixbuilder-macos-app.app
	rm -f build/phoenixbuilder-macos-app.app.zip
	mkdir -p build/phoenixbuilder-macos-app.app
	mkdir -p build/phoenixbuilder-macos-app.app/Resources
	mkdir -p build/phoenixbuilder-macos-app.app/MacOS
	ln build/phoenixbuilder-macos build/phoenixbuilder-macos-app.app/MacOS/phoenixbuilder-macos
	ln unbundled_assets/icon.icns build/phoenixbuilder-macos-app.app/Resources/icon.icns
	sed "s/@VERSION@/${VERSION}/g; s/@COMMIT_COUNT@/`git rev-list --count --all`/g" unbundled_assets/macos_Info.plist>build/phoenixbuilder-macos-app.app/Info.plist
	cd build/&&zip -r phoenixbuilder-macos-app.app.zip phoenixbuilder-macos-app.app&&rm -rf phoenixbuilder-macos-app.app&&cd ..
build/phoenixbuilder-android-arm64-lib.so: build/ ${SRCS_GO}
	CGO_CFLAGS=${CGO_DEF} CC=${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/aarch64-linux-android21-clang GOOS=android GOARCH=arm64 CGO_ENABLED=1 go build -tags "fyne_gui android" -trimpath -buildmode=c-shared -ldflags "-s -w" -o build/phoenixbuilder-android-arm64-lib.so
build/phoenixbuilder-android-app.apk: build/ build/phoenixbuilder-android-arm64-lib.so android_wrapper/android_pack
	./android_wrapper/android_pack ${VERSION} `git rev-list --count --all`
build/phoenixbuilder-windows-executable-x86.exe: build/ /usr/bin/i686-w64-mingw32-gcc ${SRCS_GO}
	CGO_CFLAGS=${CGO_DEF} CC=/usr/bin/i686-w64-mingw32-gcc GOOS=windows GOARCH=386 CGO_ENABLED=1 go build -trimpath -tags "fyne_gui windows" -ldflags "-s -w" -o build/phoenixbuilder-windows-executable-x86.exe
build/phoenixbuilder-windows-executable-x86_64.exe: build/ /usr/bin/x86_64-w64-mingw32-gcc ${SRCS_GO}
	CGO_CFLAGS=${CGO_DEF} CC=/usr/bin/x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build -trimpath -tags "fyne_gui windows" -ldflags "-s -w" -o build/phoenixbuilder-windows-executable-x86_64.exe

clean:
	rm -f build/phoenixbuilder*
