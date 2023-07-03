.PHONY: current all current-v8 current-arm64-executable ios-executable ios-v8-executable ish-executable macos macos-v8 android-executable-armv7 android-executable-arm64 android-executable-x86_64 android-executable-x86 windows-executable windows-executable-x86 windows-executable-x86_64 freebsd-executable freebsd-executable-x86 freebsd-executable-x86_64 freebsd-executable-arm64 netbsd-executable netbsd-executable-x86 netbsd-executable-x86_64 netbsd-executable-arm64 netbsd-executable netbsd-executable-x86 netbsd-executable-x86_64 netbsd-executable-arm64 openwrt-mt7620-mipsel_24kc
TARGETS:=build/ current current-no-readline current-v8
PACKAGETARGETS:=
ifeq ($(shell uname | grep "Darwin" > /dev/null ; echo $${?}),0)
ifeq ($(shell uname -m | grep -E "iPhone|iPad|iPod" > /dev/null ; echo $${?}),0)
IOS_STRIP=/usr/bin/strip
LIPO=/usr/bin/lipo
LDID=/usr/bin/ldid
TARGETS:=${TARGETS} ios-executable ios-v8-executable
else
IOS_STRIP=$(shell xcrun --sdk iphoneos -f strip)
IOS_OBJCOPY=$(shell xcrun --sdk iphoneos -f objcopy)
LDID=ldid2
LIPO=/usr/bin/lipo
TARGETS:=${TARGETS} macos ios-v8-executable ios-executable
endif
PACKAGETARGETS:=${PACKAGETARGETS} package/ios
else
IOS_STRIP=true
LDID=$${THEOS}/toolchain/linux/iphone/bin/ldid
LIPO=$${THEOS}/toolchain/linux/iphone/bin/lipo
IOS_OBJCOPY=$${THEOS}/toolchain/linux/iphone/bin/llvm-objcopy
endif

### *-----------------------------------* ###
### | These processes are designed for  | ###
### | GitHub Actions. You should ignore | ###
### | this part if not performing cross | ###
### | -compilations.                    | ###
### *-----------------------------------* ###

ifneq (${THEOS},)
	TARGETS:=${TARGETS} ios-executable ios-v8-executable macos macos-v8
	PACKAGETARGETS:=${PACKAGETARGETS} package/ios
endif
ifneq ($(wildcard ${HOME}/android-ndk-r20b),)
	TARGETS:=${TARGETS} android-v8-executable-arm64 android-executable-armv7 android-executable-arm64 android-executable-x86_64 android-executable-x86
	PACKAGETARGETS:=${PACKAGETARGETS} package/android
endif
ifneq ($(wildcard /usr/bin/i686-w64-mingw32-gcc),)
	TARGETS:=${TARGETS} windows-executable-x86
endif
ifneq ($(wildcard /usr/bin/x86_64-w64-mingw32-gcc),)
	TARGETS:=${TARGETS} windows-executable-x86_64
endif
ifneq ($(wildcard /usr/bin/aarch64-linux-gnu-gcc),)
	TARGETS:=${TARGETS} current-arm64-executable
endif
ifneq ($(wildcard ${HOME}/openwrt-sdk-21.02.2-ramips-mt7620_gcc-8.4.0_musl.Linux-x86_64),)
	TARGETS:=${TARGETS} openwrt-mt7620-mipsel_24kc
endif
ifneq ($(wildcard ${HOME}/openwrt-sdk-22.03.0-rc4-ipq40xx-generic_gcc-11.2.0_musl_eabi.Linux-x86_64),)
	TARGETS:=${TARGETS} openwrt-ipq40xx-generic-armv7
endif
ifneq ($(wildcard ${HOME}/openwrt-sdk-22.03.0-rc4-mediatek-mt7622_gcc-11.2.0_musl.Linux-x86_64),)
	TARGETS:=${TARGETS} openwrt-mt7622-arm64
endif
ifneq ($(wildcard ${HOME}/llvm),)
	TARGETS:=${TARGETS} netbsd-executable freebsd-executable openbsd-executable
	# Do other BSDs later
endif
ifneq ($(wildcard ${HOME}/i686-unknown-linux-musl),)
	TARGETS:=${TARGETS} ish-executable
endif

VERSION=$(shell cat version)

SRCS_GO := $(foreach dir, $(shell find . -type d), $(wildcard $(dir)/*.go $(dir)/*.c))

CGO_DEF := "-DFB_VERSION=\"$(VERSION)\" -DFB_COMMIT=\"$(shell git log -1 --format=format:"%h")\" -DFB_COMMIT_LONG=\"$(shell git log -1 --format=format:"%H")\""

current: build/phoenixbuilder
all: ${TARGETS} build/hashes.json
current-no-readline: build/phoenixbuilder-no-readline
current-debug: build/phoenixbuilder-debug
current-v8: build/phoenixbuilder-v8
current-arm64-executable: build/phoenixbuilder-aarch64
ios-executable: build/phoenixbuilder-ios-executable
ios-v8-executable: build/phoenixbuilder-v8-ios-executable
ish-executable: build/phoenixbuilder-ish-executable
macos: build/phoenixbuilder-macos
macos-v8: build/phoenixbuilder-v8-macos
android-executable-armv7: build/phoenixbuilder-android-static-executable-armv7 build/phoenixbuilder-android-termux-shared-executable-armv7 build/phoenixbuilder-android-shared-executable-armv7
android-executable-arm64: build/phoenixbuilder-android-static-executable-arm64 build/phoenixbuilder-android-termux-shared-executable-arm64 build/phoenixbuilder-android-shared-executable-arm64
android-v8-executable-arm64: build/phoenixbuilder-v8-android-static-executable-arm64 build/phoenixbuilder-v8-android-termux-shared-executable-arm64 build/phoenixbuilder-v8-android-shared-executable-arm64
android-executable-x86_64: build/phoenixbuilder-android-static-executable-x86_64 build/phoenixbuilder-android-termux-shared-executable-x86_64 build/phoenixbuilder-android-shared-executable-x86_64
android-executable-x86: build/phoenixbuilder-android-shared-executable-x86 build/phoenixbuilder-android-termux-shared-executable-x86 build/phoenixbuilder-android-static-executable-x86
windows-executable: windows-executable-x86 windows-executable-x86_64
windows-executable-x86: build/phoenixbuilder-windows-executable-x86.exe
windows-executable-x86_64: build/phoenixbuilder-windows-executable-x86_64.exe
freebsd-executable: freebsd-executable-x86 freebsd-executable-x86_64 freebsd-executable-arm64
freebsd-executable-x86: build/phoenixbuilder-freebsd-executable-x86
freebsd-executable-x86_64: build/phoenixbuilder-freebsd-executable-x86_64
freebsd-executable-arm64: build/phoenixbuilder-freebsd-executable-arm64
#freebsd-executable-armv6: build/phoenixbuilder-freebsd-executable-armv6
#freebsd-executable-armv7: build/phoenixbuilder-freebsd-executable-armv7
# RISC-V targets will be supported in future Go releases (Or use patched versions)
#freebsd-executable-riscv64: build/phoenixbuilder-freebsd-executable-riscv64
netbsd-executable: netbsd-executable-x86 netbsd-executable-x86_64 netbsd-executable-arm64
netbsd-executable-x86: build/phoenixbuilder-netbsd-executable-x86
netbsd-executable-x86_64: build/phoenixbuilder-netbsd-executable-x86_64
#netbsd-executable-armv6: build/phoenixbuilder-netbsd-executable-armv6
#netbsd-executable-armv7: build/phoenixbuilder-netbsd-executable-armv7
netbsd-executable-arm64: build/phoenixbuilder-netbsd-executable-arm64
openbsd-executable: openbsd-executable-x86 openbsd-executable-x86_64
# disable openbsd-executable-arm64 until I figure it out
openbsd-executable-x86: build/phoenixbuilder-openbsd-executable-x86
openbsd-executable-x86_64: build/phoenixbuilder-openbsd-executable-x86_64
#openbsd-executable-mips64: build/phoenixbuilder-openbsd-executable-mips64
#openbsd-executable-armv7: build/phoenixbuilder-openbsd-executable-armv7
openbsd-executable-arm64: build/phoenixbuilder-openbsd-executable-arm64
openwrt-mt7620-mipsel_24kc: build/phoenixbuilder-openwrt-mt7620-mipsel_24kc
openwrt-ipq40xx-generic-armv7: build/phoenixbuilder-openwrt-ipq40xx-generic-armv7
#windows-v8-executable-x86_64: build/phoenixbuilder-v8-windows-executable-x86_64.exe
#windows-shared: build/phoenixbuilder-windows-shared.dll
openwrt-mt7622-arm64: build/phoenixbuilder-openwrt-mt7622-arm64

package: ${PACKAGETARGETS}
release/:
	mkdir -p release
build/:
	mkdir build

#ifeq ($(shell uname | grep -iq 'Linux' && echo 1),1)
#ifeq ($(shell uname -m | grep -iqE "x86_64|amd64" && echo 1),1)
#APPEND_GO_TAGS := use_x86_64_linux_rl
#else
#APPEND_GO_TAGS :=
#endif
#endif

build/phoenixbuilder: build/ ${SRCS_GO}
	cd depends/stub&&make clean&&cd -
	GODEBUG=madvdontneed=1 CGO_CFLAGS=${CGO_DEF} CGO_ENABLED=1 go build -tags "${APPEND_GO_TAGS}" -trimpath -ldflags "-s -w" -o $@
build/phoenixbuilder-no-readline: build/ ${SRCS_GO}
	cd depends/stub&&make clean&&cd -
	GODEBUG=madvdontneed=1 CGO_CFLAGS=${CGO_DEF} CGO_ENABLED=1 go build -tags "no_readline ${APPEND_GO_TAGS}" -trimpath -ldflags "-s -w" -o $@
build/phoenixbuilder-with-symbols: build/ ${SRCS_GO}
	cd depends/stub&&make clean&&cd -
	GODEBUG=madvdontneed=1 CGO_CFLAGS=${CGO_DEF} CGO_ENABLED=1 go build -tags "${APPEND_GO_TAGS}" -trimpath -o $@
build/phoenixbuilder-v8: build/ ${SRCS_GO}
	cd depends/stub&&make clean&&cd -
	CGO_CFLAGS=${CGO_DEF}" -DWITH_V8" CGO_ENABLED=1 go build -tags "with_v8 ${APPEND_GO_TAGS}" -trimpath -ldflags "-s -w" -o build/phoenixbuilder-v8
build/libexternal_functions_provider.so: build/ io/external_functions_provider/provider.c
	gcc -shared io/external_functions_provider/provider.c -o build/libexternal_functions_provider.so
build/phoenixbuilder-aarch64: build/ ${SRCS_GO}
	cd depends/stub&&make clean&&make CC=/usr/bin/aarch64-linux-gnu-gcc&&cd -
	GODEBUG=madvdontneed=1 CGO_CFLAGS=${CGO_DEF} CC=/usr/bin/aarch64-linux-gnu-gcc CGO_ENABLED=1 GOARCH=arm64 go build -tags use_aarch64_linux_rl -trimpath -ldflags "-s -w" -o build/phoenixbuilder-aarch64
build/phoenixbuilder-ios-executable: build/ ${SRCS_GO}
	GODEBUG=madvdontneed=1 CGO_CFLAGS=${CGO_DEF} CC=`pwd`/archs/ios.sh CXX=`pwd`/archs/ios.sh CGO_ENABLED=1 GOOS=ios GOARCH=arm64 go build -trimpath -ldflags "-s -w" -o build/phoenixbuilder-ios-executable
	${IOS_STRIP} build/phoenixbuilder-ios-executable
	${LDID} -Sios-ent.xml build/phoenixbuilder-ios-executable
build/phoenixbuilder-v8-ios-executable: build/ ${SRCS_GO}
	GODEBUG=madvdontneed=1 CGO_CFLAGS=${CGO_DEF}" -DWITH_V8" CC=`pwd`/archs/ios.sh CXX=`pwd`/archs/ios.sh CGO_ENABLED=1 GOOS=ios GOARCH=arm64 go build -tags with_v8 -trimpath -ldflags "-s -w" -o build/phoenixbuilder-v8-ios-executable
	${IOS_STRIP} build/phoenixbuilder-v8-ios-executable
	${LDID} -Sios-ent.xml build/phoenixbuilder-v8-ios-executable
build/libexternal_functions_provider.dylib: build/ io/external_functions_provider/provider.c
	`pwd`/archs/ios.sh io/external_functions_provider/provider.c -shared -o build/libexternal_functions_provider.dylib
build/phoenixbuilder-ish-executable: build/ ${SRCS_GO}
	cd depends/stub&&make clean&&make CC="${HOME}/i686-unknown-linux-musl/bin/i686-unknown-linux-musl-gcc --sysroot=`pwd`/../buildroot/ish -L`pwd`/../buildroot/ish/usr/lib" && cd -
	GODEBUG=madvdontneed=1 CGO_CFLAGS=${CGO_DEF} CC="${HOME}/i686-unknown-linux-musl/bin/i686-unknown-linux-musl-gcc --sysroot=`pwd`/depends/buildroot/ish" CGO_ENABLED=1 GOOS=linux GOARCH=386 go build -tags "ish" -trimpath -ldflags "-s -w" -o build/phoenixbuilder-ish-executable
build/phoenixbuilder-macos-x86_64: build/ ${SRCS_GO}
	GODEBUG=madvdontneed=1 CGO_CFLAGS=${CGO_DEF} CC=`pwd`/archs/macos.sh CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o build/phoenixbuilder-macos-x86_64
build/phoenixbuilder-macos-arm64: build/ ${SRCS_GO}
	GODEBUG=madvdontneed=1 CGO_CFLAGS=${CGO_DEF} CC=`pwd`/archs/macos.sh CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags "-s -w" -o build/phoenixbuilder-macos-arm64
build/phoenixbuilder-macos: build/ build/phoenixbuilder-macos-x86_64 build/phoenixbuilder-macos-arm64 ${SRCS_GO}
	GODEBUG=madvdontneed=1 ${LIPO} -create build/phoenixbuilder-macos-x86_64 build/phoenixbuilder-macos-arm64 -output build/phoenixbuilder-macos
build/phoenixbuilder-v8-macos-x86_64: build/ ${SRCS_GO}
	GODEBUG=madvdontneed=1 CGO_CFLAGS=${CGO_DEF}" -DWITH_V8" CGO_LDFLAGS="-lc++" CC=`pwd`/archs/macos.sh CXX=`pwd`/archs/macos.sh CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -tags with_v8 -trimpath -ldflags "-s -w" -o build/phoenixbuilder-v8-macos-x86_64
build/phoenixbuilder-v8-macos-arm64: build/ ${SRCS_GO}
	GODEBUG=madvdontneed=1 CGO_CFLAGS=${CGO_DEF}" -DWITH_V8" CGO_LDFLAGS="-lc++" CC=`pwd`/archs/macos.sh CXX=`pwd`/archs/macos.sh CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -tags with_v8 -trimpath -ldflags "-s -w" -o build/phoenixbuilder-v8-macos-arm64
build/phoenixbuilder-v8-macos: build/ build/phoenixbuilder-v8-macos-x86_64 build/phoenixbuilder-v8-macos-arm64 ${SRCS_GO}
	GODEBUG=madvdontneed=1 ${LIPO} -create build/phoenixbuilder-v8-macos-x86_64 build/phoenixbuilder-v8-macos-arm64 -output build/phoenixbuilder-v8-macos
build/phoenixbuilder-android-static-executable-armv7: build/ ${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/armv7a-linux-androideabi16-clang ${SRCS_GO}
	GODEBUG=madvdontneed=1 CGO_CFLAGS=${CGO_DEF} CC=${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/armv7a-linux-androideabi16-clang GOOS=android GOARCH=arm CGO_ENABLED=1 go build -trimpath -ldflags "-s -w" -o build/phoenixbuilder-android-static-executable-armv7
build/phoenixbuilder-v8-android-static-executable-arm64: build/ ${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/aarch64-linux-android21-clang ${SRCS_GO}
	GODEBUG=madvdontneed=1 CGO_CFLAGS=${CGO_DEF}" -DWITH_V8" CC=${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/aarch64-linux-android21-clang CXX=${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/aarch64-linux-android21-clang++ GOOS=android GOARCH=arm64 CGO_ENABLED=1 go build -tags with_v8 -trimpath -ldflags "-s -w -extldflags -static-libstdc++" -o build/phoenixbuilder-v8-android-static-executable-arm64
build/phoenixbuilder-android-static-executable-arm64: build/ ${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/aarch64-linux-android21-clang ${SRCS_GO}
	GODEBUG=madvdontneed=1 CGO_CFLAGS=${CGO_DEF} CC=${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/aarch64-linux-android21-clang CXX=${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/aarch64-linux-android21-clang++ GOOS=android GOARCH=arm64 CGO_ENABLED=1 go build -trimpath -ldflags "-s -w -extldflags -static-libstdc++" -o build/phoenixbuilder-android-static-executable-arm64
build/phoenixbuilder-android-static-executable-x86: build/ ${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/i686-linux-android21-clang ${SRCS_GO}
	GODEBUG=madvdontneed=1 CGO_CFLAGS=${CGO_DEF} CC=${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/i686-linux-android21-clang GOOS=android GOARCH=386 CGO_ENABLED=1 go build -trimpath -ldflags "-s -w" -o build/phoenixbuilder-android-static-executable-x86
build/phoenixbuilder-android-static-executable-x86_64: build/ ${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/x86_64-linux-android21-clang ${SRCS_GO}
	GODEBUG=madvdontneed=1 CGO_CFLAGS=${CGO_DEF} CC=${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/x86_64-linux-android21-clang GOOS=android GOARCH=amd64 CGO_ENABLED=1 go build -trimpath -ldflags "-s -w" -o build/phoenixbuilder-android-static-executable-x86_64
build/phoenixbuilder-android-termux-shared-executable-armv7: build/ ${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/armv7a-linux-androideabi16-clang ${SRCS_GO}
	cd depends/stub&&make clean&&make CC=${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/armv7a-linux-androideabi16-clang&&cd -
	GODEBUG=madvdontneed=1 CGO_CFLAGS=${CGO_DEF} CGO_LDFLAGS="-Wl,-rpath,/data/data/com.termux/files/usr/lib" CC=${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/armv7a-linux-androideabi16-clang GOOS=android GOARCH=arm CGO_ENABLED=1 go build -tags android_shared -trimpath -ldflags "-s -w" -o build/phoenixbuilder-android-termux-shared-executable-armv7
build/phoenixbuilder-v8-android-termux-shared-executable-arm64: build/ ${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/aarch64-linux-android21-clang ${SRCS_GO}
	cd depends/stub&&make clean&&make CC=${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/aarch64-linux-android21-clang&&cd -
	GODEBUG=madvdontneed=1 CGO_CFLAGS=${CGO_DEF}" -DWITH_V8" CGO_LDFLAGS="-Wl,-rpath,/data/data/com.termux/files/usr/lib" CC=${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/aarch64-linux-android21-clang CXX=${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/aarch64-linux-android21-clang++ GOOS=android GOARCH=arm64 CGO_ENABLED=1 go build -tags with_v8,android_shared -trimpath -ldflags "-s -w -extldflags -static-libstdc++" -o build/phoenixbuilder-v8-android-termux-shared-executable-arm64
build/phoenixbuilder-android-termux-shared-executable-arm64: build/ ${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/aarch64-linux-android21-clang ${SRCS_GO}
	cd depends/stub&&make clean&&make CC=${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/aarch64-linux-android21-clang&&cd -
	GODEBUG=madvdontneed=1 CGO_CFLAGS=${CGO_DEF} CGO_LDFLAGS="-Wl,-rpath,/data/data/com.termux/files/usr/lib" CC=${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/aarch64-linux-android21-clang CXX=${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/aarch64-linux-android21-clang++ GOOS=android GOARCH=arm64 CGO_ENABLED=1 go build -tags android_shared -trimpath -ldflags "-s -w -extldflags -static-libstdc++" -o build/phoenixbuilder-android-termux-shared-executable-arm64
build/phoenixbuilder-android-termux-shared-executable-x86: build/ ${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/i686-linux-android21-clang ${SRCS_GO}
	cd depends/stub&&make clean&&make CC=${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/i686-linux-android21-clang&&cd -
	GODEBUG=madvdontneed=1 CGO_CFLAGS=${CGO_DEF} CGO_LDFLAGS="-Wl,-rpath,/data/data/com.termux/files/usr/lib" CC=${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/i686-linux-android21-clang GOOS=android GOARCH=386 CGO_ENABLED=1 go build -tags android_shared -trimpath -ldflags "-s -w" -o build/phoenixbuilder-android-termux-shared-executable-x86
build/phoenixbuilder-android-termux-shared-executable-x86_64: build/ ${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/x86_64-linux-android21-clang ${SRCS_GO}
	cd depends/stub&&make clean&&make CC=${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/x86_64-linux-android21-clang&&cd -
	GODEBUG=madvdontneed=1 CGO_CFLAGS=${CGO_DEF} CGO_LDFLAGS="-Wl,-rpath,/data/data/com.termux/files/usr/lib" CC=${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/x86_64-linux-android21-clang GOOS=android GOARCH=amd64 CGO_ENABLED=1 go build -tags android_shared -trimpath -ldflags "-s -w" -o build/phoenixbuilder-android-termux-shared-executable-x86_64
build/phoenixbuilder-android-executable-arm64: build/phoenixbuilder-android-termux-shared-executable-arm64
	ln build/phoenixbuilder-android-termux-shared-executable-arm64 build/phoenixbuilder-android-executable-arm64
build/phoenixbuilder-android-executable-armv7: build/phoenixbuilder-android-termux-shared-executable-armv7
	ln build/phoenixbuilder-android-termux-shared-executable-armv7 build/phoenixbuilder-android-executable-armv7
build/phoenixbuilder-android-executable-x86: build/phoenixbuilder-android-termux-shared-executable-x86
	ln build/phoenixbuilder-android-termux-shared-executable-x86 build/phoenixbuilder-android-executable-x86
build/phoenixbuilder-android-executable-x86_64: build/phoenixbuilder-android-termux-shared-executable-x86_64
	ln build/phoenixbuilder-android-termux-shared-executable-x86_64 build/phoenixbuilder-android-executable-x86_64
build/phoenixbuilder-android-shared-executable-armv7: build/ ${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/armv7a-linux-androideabi16-clang ${SRCS_GO}
	cd depends/stub&&make clean&&make CC=${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/armv7a-linux-androideabi16-clang&&cd -
	GODEBUG=madvdontneed=1 CGO_CFLAGS=${CGO_DEF} CC=${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/armv7a-linux-androideabi16-clang GOOS=android GOARCH=arm CGO_ENABLED=1 go build -tags android_shared -trimpath -ldflags "-s -w" -o build/phoenixbuilder-android-shared-executable-armv7
build/phoenixbuilder-v8-android-shared-executable-arm64: build/ ${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/aarch64-linux-android21-clang ${SRCS_GO}
	cd depends/stub&&make clean&&make CC=${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/aarch64-linux-android21-clang&&cd -
	GODEBUG=madvdontneed=1 CGO_CFLAGS=${CGO_DEF}" -DWITH_V8" CC=${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/aarch64-linux-android21-clang CXX=${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/aarch64-linux-android21-clang++ GOOS=android GOARCH=arm64 CGO_ENABLED=1 go build -tags with_v8,android_shared -trimpath -ldflags "-s -w -extldflags -static-libstdc++" -o build/phoenixbuilder-v8-android-shared-executable-arm64
build/phoenixbuilder-android-shared-executable-arm64: build/ ${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/aarch64-linux-android21-clang ${SRCS_GO}
	cd depends/stub&&make clean&&make CC=${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/aarch64-linux-android21-clang&&cd -
	GODEBUG=madvdontneed=1 CGO_CFLAGS=${CGO_DEF} CC=${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/aarch64-linux-android21-clang CXX=${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/aarch64-linux-android21-clang++ GOOS=android GOARCH=arm64 CGO_ENABLED=1 go build -tags android_shared -trimpath -ldflags "-s -w -extldflags -static-libstdc++" -o build/phoenixbuilder-android-shared-executable-arm64
build/phoenixbuilder-android-shared-executable-x86: build/ ${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/i686-linux-android21-clang ${SRCS_GO}
	cd depends/stub&&make clean&&make CC=${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/i686-linux-android21-clang&&cd -
	GODEBUG=madvdontneed=1 CGO_CFLAGS=${CGO_DEF} CC=${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/i686-linux-android21-clang GOOS=android GOARCH=386 CGO_ENABLED=1 go build -tags android_shared -trimpath -ldflags "-s -w" -o build/phoenixbuilder-android-shared-executable-x86
build/phoenixbuilder-android-shared-executable-x86_64: build/ ${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/x86_64-linux-android21-clang ${SRCS_GO}
	cd depends/stub&&make clean&&make CC=${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/x86_64-linux-android21-clang&&cd -
	GODEBUG=madvdontneed=1 CGO_CFLAGS=${CGO_DEF} CC=${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/x86_64-linux-android21-clang GOOS=android GOARCH=amd64 CGO_ENABLED=1 go build -tags android_shared -trimpath -ldflags "-s -w" -o build/phoenixbuilder-android-shared-executable-x86_64
build/phoenixbuilder-windows-executable-x86.exe: build/ /usr/bin/i686-w64-mingw32-gcc ${SRCS_GO}
	GODEBUG=madvdontneed=1 CGO_CFLAGS=${CGO_DEF} CC=/usr/bin/i686-w64-mingw32-gcc GOOS=windows GOARCH=386 CGO_ENABLED=1 go build -trimpath -ldflags "-s -w" -o build/phoenixbuilder-windows-executable-x86.exe
build/phoenixbuilder-windows-executable-x86_64.exe: build/ /usr/bin/x86_64-w64-mingw32-gcc ${SRCS_GO}
	GODEBUG=madvdontneed=1 CGO_CFLAGS=${CGO_DEF} CC=/usr/bin/x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build -trimpath -ldflags "-s -w" -o build/phoenixbuilder-windows-executable-x86_64.exe
build/phoenixbuilder-freebsd-executable-x86:
	cd depends/stub&&make clean&&make ZLIB_SOVERSION=6 ALT_CLANG="${HOME}/llvm/bin/clang" CC="`pwd`/../buildroot/freebsd/bin/clang -target i686-unknown-freebsd --sysroot=`pwd`/../buildroot/freebsd/i386 -L`pwd`/../buildroot/freebsd/i386/usr/lib -fuse-ld=${HOME}/llvm/bin/ld.lld -Wno-unused-command-line-argument"&&cd -
	GODEBUG=madvdontneed=1 CGO_CFLAGS=${CGO_DEF} CGO_LDFLAGS="-Ldepends/buildroot/freebsd/i386/lib -Ldepends/buildroot/freebsd/i386/usr/lib -Wl,-rpath,/usr/local/lib,-rpath,/lib" ALT_CLANG="${HOME}/llvm/bin/clang" CC="`pwd`/depends/buildroot/freebsd/bin/clang -target i686-unknown-freebsd --sysroot=`pwd`/depends/buildroot/freebsd/i386 -fuse-ld=${HOME}/llvm/bin/ld.lld -Wno-unused-command-line-argument" GOOS=freebsd GOARCH=386 CGO_ENABLED=1 go build -trimpath -ldflags "-s -w" -o build/phoenixbuilder-freebsd-executable-x86
build/phoenixbuilder-freebsd-executable-x86_64:
	cd depends/stub&&make clean&&make ZLIB_SOVERSION=6 ALT_CLANG="${HOME}/llvm/bin/clang" CC="`pwd`/../buildroot/freebsd/bin/clang -target amd64-unknown-freebsd --sysroot=`pwd`/../buildroot/freebsd/amd64 -L`pwd`/../buildroot/freebsd/amd64/usr/lib -fuse-ld=${HOME}/llvm/bin/ld.lld -Wno-unused-command-line-argument"&&cd -
	GODEBUG=madvdontneed=1 CGO_CFLAGS=${CGO_DEF} CGO_LDFLAGS="-Ldepends/buildroot/freebsd/amd64/lib -Ldepends/buildroot/freebsd/amd64/usr/lib -Wl,-rpath,/usr/local/lib,-rpath,/lib" ALT_CLANG="${HOME}/llvm/bin/clang" CC="`pwd`/depends/buildroot/freebsd/bin/clang -target amd64-unknown-freebsd --sysroot=`pwd`/depends/buildroot/freebsd/amd64 -fuse-ld=${HOME}/llvm/bin/ld.lld -Wno-unused-command-line-argument" GOOS=freebsd GOARCH=amd64 CGO_ENABLED=1 go build -trimpath -ldflags "-s -w" -o build/phoenixbuilder-freebsd-executable-x86_64
build/phoenixbuilder-freebsd-executable-arm64:
	cd depends/stub&&make clean&&make ZLIB_SOVERSION=6 ALT_CLANG="${HOME}/llvm/bin/clang" CC="`pwd`/../buildroot/freebsd/bin/clang -target aarch64-unknown-freebsd --sysroot=`pwd`/../buildroot/freebsd/arm64 -L`pwd`/../buildroot/freebsd/arm64/usr/lib -fuse-ld=${HOME}/llvm/bin/ld.lld -Wno-unused-command-line-argument"&&cd -
	GODEBUG=madvdontneed=1 CGO_CFLAGS=${CGO_DEF} CGO_LDFLAGS="-Ldepends/buildroot/freebsd/arm64/lib -Ldepends/buildroot/freebsd/arm64/usr/lib -Wl,-rpath,/usr/local/lib,-rpath,/lib" ALT_CLANG="${HOME}/llvm/bin/clang" CC="`pwd`/depends/buildroot/freebsd/bin/clang -target aarch64-unknown-freebsd --sysroot=`pwd`/depends/buildroot/freebsd/arm64 -fuse-ld=${HOME}/llvm/bin/ld.lld -Wno-unused-command-line-argument" GOOS=freebsd GOARCH=arm64 CGO_ENABLED=1 go build -trimpath -ldflags "-s -w" -o build/phoenixbuilder-freebsd-executable-arm64
build/phoenixbuilder-netbsd-executable-x86:
	cd depends/stub&&make clean&&make ALT_CLANG="${HOME}/llvm/bin/clang" CC="`pwd`/../buildroot/netbsd/bin/clang -target i386--netbsd --sysroot=`pwd`/../buildroot/netbsd/i386 -L`pwd`/../buildroot/netbsd/i386/usr/lib -fuse-ld=${HOME}/llvm/bin/ld.lld -Wno-unused-command-line-argument"&&cd -
	GODEBUG=madvdontneed=1 CGO_CFLAGS=${CGO_DEF} CGO_LDFLAGS="-Ldepends/buildroot/netbsd/i386/lib -Ldepends/buildroot/netbsd/i386/usr/lib -Wl,-rpath,/usr/pkg/lib" ALT_CLANG="${HOME}/llvm/bin/clang" CC="`pwd`/depends/buildroot/netbsd/bin/clang -target i386--netbsd --sysroot=`pwd`/depends/buildroot/netbsd/i386 -fuse-ld=${HOME}/llvm/bin/ld.lld -Wno-unused-command-line-argument" GOOS=netbsd GOARCH=386 CGO_ENABLED=1 go build -trimpath -ldflags "-s -w" -o build/phoenixbuilder-netbsd-executable-x86
build/phoenixbuilder-netbsd-executable-x86_64:
	cd depends/stub&&make clean&&make ALT_CLANG="${HOME}/llvm/bin/clang" CC="`pwd`/../buildroot/netbsd/bin/clang -target amd64--netbsd --sysroot=`pwd`/../buildroot/netbsd/amd64 -L`pwd`/../buildroot/netbsd/amd64/usr/lib -fuse-ld=${HOME}/llvm/bin/ld.lld -Wno-unused-command-line-argument"&&cd -
	GODEBUG=madvdontneed=1 CGO_CFLAGS=${CGO_DEF} CGO_LDFLAGS="-Ldepends/buildroot/netbsd/amd64/lib -Ldepends/buildroot/netbsd/amd64/usr/lib -Wl,-rpath,/usr/pkg/lib" ALT_CLANG="${HOME}/llvm/bin/clang" CC="`pwd`/depends/buildroot/netbsd/bin/clang -target amd64--netbsd --sysroot=`pwd`/depends/buildroot/netbsd/amd64 -fuse-ld=${HOME}/llvm/bin/ld.lld -Wno-unused-command-line-argument" CGO_ENABLED=1 GOOS=netbsd GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o build/phoenixbuilder-netbsd-executable-x86_64
# ld.lld: error: unknown emulation: armelf_nbsd
# We need alternative ld for arm
#build/phoenixbuilder-netbsd-executable-armv6:
#	GODEBUG=madvdontneed=1 CGO_CFLAGS=${CGO_DEF} CGO_LDFLAGS="-Ldepends/buildroot/netbsd/armv6/lib -Ldepends/buildroot/netbsd/armv6/usr/lib -Wl,-rpath,/usr/pkg/lib" CC="${HOME}/llvm/bin/clang -target armv6--netbsd --sysroot=`pwd`/depends/buildroot/netbsd/armv6 -fuse-ld=${HOME}/llvm/bin/ld.lld -Wno-unused-command-line-argument" CGO_ENABLED=1 GOOS=netbsd GOARCH=arm GOARM=6 go build -trimpath -ldflags "-s -w" -o build/phoenixbuilder-netbsd-executable-armv6
#build/phoenixbuilder-netbsd-executable-armv7:
#	GODEBUG=madvdontneed=1 CGO_CFLAGS=${CGO_DEF} CGO_LDFLAGS="-Ldepends/buildroot/netbsd/armv7/lib -Ldepends/buildroot/netbsd/armv7/usr/lib -Wl,-rpath,/usr/pkg/lib" CC="${HOME}/llvm/bin/clang -target armv7--netbsd --sysroot=`pwd`/depends/buildroot/netbsd/armv7 -fuse-ld=${HOME}/llvm/bin/ld.lld -Wno-unused-command-line-argument" CGO_ENABLED=1 GOOS=netbsd GOARCH=arm GOARM=7 go build -trimpath -ldflags "-s -w" -o build/phoenixbuilder-netbsd-executable-armv7
build/phoenixbuilder-netbsd-executable-arm64:
	cd depends/stub&&make clean&&make ALT_CLANG="${HOME}/llvm/bin/clang" CC="`pwd`/../buildroot/netbsd/bin/clang -target aarch64--netbsd --sysroot=`pwd`/../buildroot/netbsd/arm64 -L`pwd`/../buildroot/netbsd/arm64/usr/lib -fuse-ld=${HOME}/llvm/bin/ld.lld -Wno-unused-command-line-argument"&&cd -
	GODEBUG=madvdontneed=1 CGO_CFLAGS=${CGO_DEF} CGO_LDFLAGS="-Ldepends/buildroot/netbsd/arm64/lib -Ldepends/buildroot/netbsd/arm64/usr/lib -Wl,-rpath,/usr/pkg/lib" ALT_CLANG="${HOME}/llvm/bin/clang" CC="`pwd`/depends/buildroot/netbsd/bin/clang -target aarch64--netbsd --sysroot=`pwd`/depends/buildroot/netbsd/arm64 -fuse-ld=${HOME}/llvm/bin/ld.lld -Wno-unused-command-line-argument" CGO_ENABLED=1 GOOS=netbsd GOARCH=arm64 go build -trimpath -ldflags "-s -w" -o build/phoenixbuilder-netbsd-executable-arm64
build/phoenixbuilder-openbsd-executable-x86:
	cd depends/stub&&make clean&&make ZLIB_SOVERSION=7.0 ALT_LD="${HOME}/llvm/bin/ld.lld" CC="${HOME}/llvm/bin/clang -target i386-unknown-openbsd --sysroot=`pwd`/../buildroot/openbsd/i386 -L`pwd`/../buildroot/openbsd/i386/usr/lib -fuse-ld=`pwd`/../buildroot/openbsd/bin/ld.lld -Wno-unused-command-line-argument"&&cd -
	GODEBUG=madvdontneed=1 CGO_LDFLAGS="-Ldepends/buildroot/openbsd/i386/usr/lib -Wl,-rpath,/usr/local/lib" CGO_CFLAGS=${CGO_DEF} ALT_LD="${HOME}/llvm/bin/ld.lld" CC="${HOME}/llvm/bin/clang -target i386-unknown-openbsd --sysroot=`pwd`/depends/buildroot/openbsd/i386 -fuse-ld=`pwd`/depends/buildroot/openbsd/bin/ld.lld -Wno-unused-command-line-argument" CGO_ENABLED=1 GOOS=openbsd GOARCH=386 go build -trimpath -ldflags "-s -w" -o build/phoenixbuilder-openbsd-executable-x86
build/phoenixbuilder-openbsd-executable-x86_64:
	cd depends/stub&&make clean&&make ZLIB_SOVERSION=7.0 ALT_LD="${HOME}/llvm/bin/ld.lld" CC="${HOME}/llvm/bin/clang -target amd64-unknown-openbsd --sysroot=`pwd`/../buildroot/openbsd/amd64 -L`pwd`/../buildroot/openbsd/amd64/usr/lib -fuse-ld=`pwd`/../buildroot/openbsd/bin/ld.lld -Wno-unused-command-line-argument"&&cd -
	GODEBUG=madvdontneed=1 CGO_LDFLAGS="-Ldepends/buildroot/openbsd/amd64/usr/lib -Wl,-rpath,/usr/local/lib" CGO_CFLAGS=${CGO_DEF} ALT_LD="${HOME}/llvm/bin/ld.lld" CC="${HOME}/llvm/bin/clang -target amd64-unknown-openbsd --sysroot=`pwd`/depends/buildroot/openbsd/amd64 -fuse-ld=`pwd`/depends/buildroot/openbsd/bin/ld.lld -Wno-unused-command-line-argument" CGO_ENABLED=1 GOOS=openbsd GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o build/phoenixbuilder-openbsd-executable-x86_64
build/phoenixbuilder-openbsd-executable-arm64:
	cd depends/stub&&make clean&&make ZLIB_SOVERSION=7.0 ALT_LD="${HOME}/llvm/bin/ld.lld" CC="${HOME}/llvm/bin/clang -target arm64-unknown-openbsd --sysroot=`pwd`/../buildroot/openbsd/arm64 -L`pwd`/../buildroot/openbsd/arm64/usr/lib -fuse-ld=`pwd`/../buildroot/openbsd/bin/ld.lld -Wno-unused-command-line-argument"&&cd -
	GODEBUG=madvdontneed=1 CGO_LDFLAGS="-Ldepends/buildroot/openbsd/arm64/usr/lib -Wl,-rpath,/usr/local/lib" CGO_CFLAGS=${CGO_DEF} ALT_LD="${HOME}/llvm/bin/ld.lld" CC="${HOME}/llvm/bin/clang -target arm64-unknown-openbsd --sysroot=`pwd`/depends/buildroot/openbsd/arm64 -fuse-ld=`pwd`/depends/buildroot/openbsd/bin/ld.lld -Wno-unused-command-line-argument" CGO_ENABLED=1 GOOS=openbsd GOARCH=arm64 go build -trimpath -ldflags "-s -w" -o build/phoenixbuilder-openbsd-executable-arm64
#build/phoenixbuilder-openbsd-executable-armv7:
#	GODEBUG=madvdontneed=1 CGO_LDFLAGS="-Ldepends/buildroot/openbsd/armv7/usr/lib -Wl,-rpath,/usr/local/lib" CGO_CFLAGS=${CGO_DEF} ALT_LD="${HOME}/llvm/bin/ld.lld" CC="${HOME}/llvm/bin/clang -target armv7-unknown-openbsd --sysroot=`pwd`/depends/buildroot/openbsd/armv7 -fuse-ld=`pwd`/depends/buildroot/openbsd/bin/ld.lld -Wno-unused-command-line-argument" CGO_ENABLED=1 GOOS=openbsd GOARCH=arm GOARM=7 go build -trimpath -ldflags "-s -w" -o build/phoenixbuilder-openbsd-executable-armv7
#build/phoenixbuilder-openbsd-executable-mips64:
#	GODEBUG=madvdontneed=1 CGO_LDFLAGS="-Ldepends/buildroot/openbsd/mips64/usr/lib -Wl,-rpath,/usr/local/lib" CGO_CFLAGS=${CGO_DEF} ALT_LD="${HOME}/llvm/bin/ld.lld" CC="${HOME}/llvm/bin/clang -target mips64-unknown-openbsd --sysroot=`pwd`/depends/buildroot/openbsd/mips64 -fuse-ld=`pwd`/depends/buildroot/openbsd/bin/ld.lld -Wno-unused-command-line-argument" CGO_ENABLED=1 GOOS=openbsd GOARCH=mips64 go build -trimpath -ldflags "-s -w" -o build/phoenixbuilder-openbsd-executable-mips64
build/phoenixbuilder-openwrt-mt7620-mipsel_24kc: build/ ${HOME}/openwrt-sdk-21.02.2-ramips-mt7620_gcc-8.4.0_musl.Linux-x86_64/ ${SRCS_GO}
	cd depends/stub&&make clean&&make CC=${HOME}/openwrt-sdk-21.02.2-ramips-mt7620_gcc-8.4.0_musl.Linux-x86_64/staging_dir/toolchain-mipsel_24kc_gcc-8.4.0_musl/bin/mipsel-openwrt-linux-gcc&&cd -
	GODEBUG=madvdontneed=1 CGO_CFLAGS=${CGO_DEF} STAGING_DIR=${HOME}/openwrt-sdk-21.02.2-ramips-mt7620_gcc-8.4.0_musl.Linux-x86_64/staging_dir CC=${HOME}/openwrt-sdk-21.02.2-ramips-mt7620_gcc-8.4.0_musl.Linux-x86_64/staging_dir/toolchain-mipsel_24kc_gcc-8.4.0_musl/bin/mipsel-openwrt-linux-gcc CXX=${HOME}/openwrt-sdk-21.02.2-ramips-mt7620_gcc-8.4.0_musl.Linux-x86_64/staging_dir/toolchain-mipsel_24kc_gcc-8.4.0_musl/bin/mipsel-openwrt-linux-g++ GOARCH=mipsle CGO_ENABLED=1 go build -trimpath -tags openwrt_readline -ldflags "-s -w" -o build/phoenixbuilder-openwrt-mt7620-mipsel_24kc
#build/phoenixbuilder-v8-windows-executable-x86_64.exe: build/ /usr/bin/x86_64-w64-mingw32-gcc ${SRCS_GO}
#	CGO_CFLAGS=${CGO_DEF}" -DWITH_V8" CC=/usr/bin/x86_64-w64-mingw32-gcc CXX=/usr/bin/x86_64-w64-mingw32-g++ GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build -tags with_v8 -trimpath -ldflags "-s -w" -o build/phoenixbuilder-v8-windows-executable-x86_64.exe
#build/phoenixbuilder-windows-shared.dll: build/ /usr/bin/x86_64-w64-mingw32-gcc ${SRCS_GO}
#	CGO_CFLAGS="-Wl,--enable-stdcall-fixup -luser32 -lcomdlg32 -Wno-pointer-to-int-cast -mwindows -m64 -march=x86-64 -luser32 -lkernel32 -lgdi32 -lwinmm -lcomctl32 -ladvapi32 -lshell32 -lpsapi -nodefaultlibs -nostdlib -lmsvcrt -D_UCRT=1" CC=/usr/bin/x86_64-w64-mingw32-gcc CGO_LDFLAGS="--enable-stdcall-fixup" GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build -buildmode=c-shared -trimpath -o build/phoenixbuilder-windows-shared.dll
build/phoenixbuilder-openwrt-ipq40xx-generic-armv7: build/ ${HOME}/openwrt-sdk-22.03.0-rc4-ipq40xx-generic_gcc-11.2.0_musl_eabi.Linux-x86_64/ ${SRCS_GO}
	cd depends/stub&&make clean&&make CC=${HOME}/openwrt-sdk-22.03.0-rc4-ipq40xx-generic_gcc-11.2.0_musl_eabi.Linux-x86_64/staging_dir/toolchain-arm_cortex-a7+neon-vfpv4_gcc-11.2.0_musl_eabi/bin/arm-openwrt-linux-gcc&&cd -
	GODEBUG=madvdontneed=1 CGO_CFLAGS=${CGO_DEF} STAGING_DIR=${HOME}/openwrt-sdk-22.03.0-rc4-ipq40xx-generic_gcc-11.2.0_musl_eabi.Linux-x86_64/staging_dir CC=${HOME}/openwrt-sdk-22.03.0-rc4-ipq40xx-generic_gcc-11.2.0_musl_eabi.Linux-x86_64/staging_dir/toolchain-arm_cortex-a7+neon-vfpv4_gcc-11.2.0_musl_eabi/bin/arm-openwrt-linux-gcc CXX=${HOME}/openwrt-sdk-22.03.0-rc4-ipq40xx-generic_gcc-11.2.0_musl_eabi.Linux-x86_64/staging_dir/toolchain-arm_cortex-a7+neon-vfpv4_gcc-11.2.0_musl_eabi/bin/arm-openwrt-linux-g++ GOARCH=arm GOARM=7 CGO_ENABLED=1 go build -trimpath -ldflags "-s -w" -tags no_readline -o build/phoenixbuilder-openwrt-ipq40xx-generic-armv7
build/phoenixbuilder-openwrt-mt7622-arm64: build/ ${HOME}/openwrt-sdk-22.03.0-rc4-mediatek-mt7622_gcc-11.2.0_musl.Linux-x86_64/ ${SRCS_GO}
	cd depends/stub&&make clean&&make CC=${HOME}/openwrt-sdk-22.03.0-rc4-mediatek-mt7622_gcc-11.2.0_musl.Linux-x86_64/staging_dir/toolchain-aarch64_cortex-a53_gcc-11.2.0_musl/bin/aarch64-openwrt-linux-gcc&&cd -
	GODEBUG=madvdontneed=1 CGO_CFLAGS=${CGO_DEF} STAGING_DIR=${HOME}/openwrt-sdk-22.03.0-rc4-mediatek-mt7622_gcc-11.2.0_musl.Linux-x86_64/staging_dir CC=${HOME}/openwrt-sdk-22.03.0-rc4-mediatek-mt7622_gcc-11.2.0_musl.Linux-x86_64/staging_dir/toolchain-aarch64_cortex-a53_gcc-11.2.0_musl/bin/aarch64-openwrt-linux-gcc CXX=${HOME}/openwrt-sdk-22.03.0-rc4-mediatek-mt7622_gcc-11.2.0_musl.Linux-x86_64/staging_dir/toolchain-aarch64_cortex-a53_gcc-11.2.0_musl/bin/aarch64-openwrt-linux-g++ GOARCH=arm64 CGO_ENABLED=1 go build -trimpath -ldflags "-s -w" -tags no_readline -o build/phoenixbuilder-openwrt-mt7622-arm64
build/hashes.json: build genhash.js ${TARGETS}
	node genhash.js
	cp version build/version

package/ios: build/phoenixbuilder-ios-executable release/
	mkdir -p release/phoenixbuilder-iphoneos/usr/local/bin release/phoenixbuilder-iphoneos/DEBIAN
	cp build/phoenixbuilder-ios-executable release/phoenixbuilder-iphoneos/usr/local/bin/fastbuilder
	printf "Package: pro.fastbuilder.phoenix\n\
	Name: FastBuilder Phoenix (Alpha)\n\
	Version: $(VERSION)\n\
	Architecture: iphoneos-arm\n\
	Maintainer: Ruphane\n\
	Author: Bouldev <admin@boul.dev>\n\
	Section: Games\n\
	Priority: optional\n\
	Enhances: mterminal | openssh | ws.hbang.newterm2\n\
	Homepage: https://fastbuilder.pro\n\
	Depiction: https://apt.boul.dev/info/fastbuilder\n\
	Description: Modern Minecraft structuring tool\n" > release/phoenixbuilder-iphoneos/DEBIAN/control
	dpkg-deb -Zxz -b release/phoenixbuilder-iphoneos release/
package/android: package/android-armv7 package/android-arm64
package/android-armv7: build/phoenixbuilder-android-executable-armv7 release/
	mkdir -p release/phoenixbuilder-android-armv7/data/data/com.termux/files/usr/bin release/phoenixbuilder-android-armv7/DEBIAN
	cp build/phoenixbuilder-android-executable-armv7 release/phoenixbuilder-android-armv7/data/data/com.termux/files/usr/bin/fastbuilder
	printf "Package: pro.fastbuilder.phoenix-android\n\
	Name: FastBuilder Phoenix (Alpha)\n\
	Version: $(VERSION)\n\
	Architecture: arm\n\
	Depends: libreadline8 | readline (>= 8.0.0), zlib | zlib1g\n\
	Maintainer: Ruphane\n\
	Author: Bouldev <admin@boul.dev>\n\
	Section: Games\n\
	Priority: optional\n\
	Homepage: https://fastbuilder.pro\n\
	Description: Modern Minecraft structuring tool\n" > release/phoenixbuilder-android-armv7/DEBIAN/control
	dpkg-deb -Zxz -b release/phoenixbuilder-android-armv7 release/
package/android-arm64: build/phoenixbuilder-android-executable-arm64 release/
	mkdir -p release/phoenixbuilder-android-arm64/data/data/com.termux/files/usr/bin release/phoenixbuilder-android-arm64/DEBIAN
	cp build/phoenixbuilder-android-executable-arm64 release/phoenixbuilder-android-arm64/data/data/com.termux/files/usr/bin/fastbuilder
	printf "Package: pro.fastbuilder.phoenix-android\n\
	Name: FastBuilder Phoenix (Alpha)\n\
	Version: $(VERSION)\n\
	Architecture: aarch64\n\
	Depends: libreadline8 | readline (>= 8.0.0), zlib | zlib1g\n\
	Maintainer: Ruphane\n\
	Author: Bouldev <admin@boul.dev>\n\
	Section: Games\n\
	Priority: optional\n\
	Homepage: https://fastbuilder.pro\n\
	Description: Modern Minecraft structuring tool\n" > release/phoenixbuilder-android-arm64/DEBIAN/control
	dpkg-deb -Zxz -b release/phoenixbuilder-android-arm64 release/
clean:
	rm -f build/phoenixbuilder*
