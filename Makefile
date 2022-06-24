.PHONY: all current current-v8 current-arm64-executable ios-executable ios-v8-executable ios-lib macos android-executable-armv7 android-executable-arm64 android-executable-x86_64 android-executable-x86 windows-executable windows-executable-x86 windows-executable-x86_64
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
ifneq (${THEOS},)
	TARGETS:=${TARGETS} ios-executable ios-lib macos ios-v8-executable
	PACKAGETARGETS:=${PACKAGETARGETS} package/ios
endif
ifneq ($(wildcard ${HOME}/android-ndk-r20b),)
	TARGETS:=${TARGETS} android-v8-executable-64 android-executable-armv7 android-executable-arm64 android-executable-x86_64 android-executable-x86
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
ifneq ($(wildcard `pwd`/openwrt-sdk-21.02.2-ramips-mt7620_gcc-8.4.0_musl.Linux-x86_64),)
	TARGETS:=${TARGETS} openwrt-mt7620-mipsel_24kc
endif

VERSION=$(shell cat version)

SRCS_GO := $(foreach dir, $(shell find . -type d), $(wildcard $(dir)/*.go $(dir)/*.c))

CGO_DEF := "-DFB_VERSION=\"$(VERSION)\" -DFB_COMMIT=\"$(shell git log -1 --format=format:"%h")\" -DFB_COMMIT_LONG=\"$(shell git log -1 --format=format:"%H")\""

all: ${TARGETS} build/hashes.json
current: build/phoenixbuilder
current-v8: build/phoenixbuilder-v8
current-arm64-executable: build/phoenixbuilder-aarch64
ios-executable: build/phoenixbuilder-ios-executable
ios-v8-executable: build/phoenixbuilder-v8-ios-executable
ios-lib: build/phoenixbuilder-ios-static.a
macos: build/phoenixbuilder-macos
android-executable-armv7: build/phoenixbuilder-android-executable-armv7
android-executable-arm64: build/phoenixbuilder-android-executable-arm64
android-v8-executable-64: build/phoenixbuilder-v8-android-executable-arm64
android-executable-x86_64: build/phoenixbuilder-android-executable-x86_64
android-executable-x86: build/phoenixbuilder-android-executable-x86
windows-executable: windows-executable-x86 windows-executable-x86_64
windows-executable-x86: build/phoenixbuilder-windows-executable-x86.exe
windows-executable-x86_64: build/phoenixbuilder-windows-executable-x86_64.exe
openwrt-mt7620-mipsel_24kc: build/phoenixbuilder-openwrt-mt7620-mipsel_24kc
#windows-v8-executable-x86_64: build/phoenixbuilder-v8-windows-executable-x86_64.exe
#windows-shared: build/phoenixbuilder-windows-shared.dll

package: ${PACKAGETARGETS}
release/:
	mkdir -p release
build/:
	mkdir build
build/phoenixbuilder: build/ ${SRCS_GO}
	CGO_CFLAGS=${CGO_DEF} CGO_ENABLED=1  go build -trimpath -ldflags "-s -w" -o build/phoenixbuilder
build/phoenixbuilder-v8: build/ ${SRCS_GO}
	CGO_CFLAGS=${CGO_DEF}" -DWITH_V8" CGO_ENABLED=1  go build -tags with_v8 -trimpath -ldflags "-s -w" -o build/phoenixbuilder-v8
build/libexternal_functions_provider.so: build/ io/external_functions_provider/provider.c
	gcc -shared io/external_functions_provider/provider.c -o build/libexternal_functions_provider.so
build/phoenixbuilder-static.a: build/ build/libexternal_functions_provider.so ${SRCS_GO}
	CGO_CFLAGS=${CGO_DEF} CGO_LDFLAGS="-Lbuild -lexternal_functions_provider" CGO_ENABLED=1  go build -trimpath -buildmode=c-archive -ldflags "-s -w" -tags no_readline,is_tweak -o build/phoenixbuilder-static.a
build/phoenixbuilder-aarch64: build/ ${SRCS_GO}
	CGO_CFLAGS=${CGO_DEF} CC=/usr/bin/aarch64-linux-gnu-gcc CGO_ENABLED=1 GOARCH=arm64 go build -tags no_readline -trimpath -ldflags "-s -w" -o build/phoenixbuilder-aarch64
build/phoenixbuilder-ios-executable: build/ ${SRCS_GO}
	CGO_CFLAGS=${CGO_DEF} CC=`pwd`/archs/ios.sh CXX=`pwd`/archs/ios.sh CGO_ENABLED=1 GOOS=ios GOARCH=arm64 go build -trimpath -ldflags "-s -w" -o build/phoenixbuilder-ios-executable
	${IOS_STRIP} build/phoenixbuilder-ios-executable
	${LDID} -Sios-ent.xml build/phoenixbuilder-ios-executable
build/phoenixbuilder-v8-ios-executable: build/ ${SRCS_GO}
	CGO_CFLAGS=${CGO_DEF}" -DWITH_V8" CC=`pwd`/archs/ios.sh CXX=`pwd`/archs/ios.sh CGO_ENABLED=1 GOOS=ios GOARCH=arm64 go build -tags with_v8 -trimpath -ldflags "-s -w" -o build/phoenixbuilder-v8-ios-executable
	${IOS_STRIP} build/phoenixbuilder-v8-ios-executable
	${LDID} -Sios-ent.xml build/phoenixbuilder-v8-ios-executable
build/libexternal_functions_provider.dylib: build/ io/external_functions_provider/provider.c
	`pwd`/archs/ios.sh io/external_functions_provider/provider.c -shared -o build/libexternal_functions_provider.dylib
build/phoenixbuilder-ios-static.a: build/ build/libexternal_functions_provider.dylib ${SRCS_GO}
	CGO_CFLAGS=${CGO_DEF} CGO_LDFLAGS="-Lbuild -lexternal_functions_provider" CC=`pwd`/archs/ios.sh CGO_ENABLED=1 GOOS=ios GOARCH=arm64 go build -buildmode=c-archive -trimpath -ldflags "-s -w -extar ${THEOS}/toolchain/linux/iphone/bin/ar" -tags is_tweak,no_readline -o build/phoenixbuilder-ios-static.a
build/phoenixbuilder-macos-x86_64: build/ ${SRCS_GO}
	CGO_CFLAGS=${CGO_DEF} CC=`pwd`/archs/macos.sh CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o build/phoenixbuilder-macos-x86_64
build/phoenixbuilder-macos-arm64: build/ ${SRCS_GO}
	CGO_CFLAGS=${CGO_DEF} CC=`pwd`/archs/macos.sh CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags "-s -w" -o build/phoenixbuilder-macos-arm64
build/phoenixbuilder-macos: build/ build/phoenixbuilder-macos-x86_64 build/phoenixbuilder-macos-arm64 ${SRCS_GO}
	${LIPO} -create build/phoenixbuilder-macos-x86_64 build/phoenixbuilder-macos-arm64 -output build/phoenixbuilder-macos
build/phoenixbuilder-v8-macos-x86_64: build/ ${SRCS_GO}
	CGO_CFLAGS=${CGO_DEF}" -DWITH_V8" CC=`pwd`/archs/macos.sh CXX=`pwd`/archs/macos.sh CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -tags with_v8 -trimpath -ldflags "-s -w" -o build/phoenixbuilder-v8-macos-x86_64
build/phoenixbuilder-v8-macos-arm64: build/ ${SRCS_GO}
	CGO_CFLAGS=${CGO_DEF}" -DWITH_V8" CC=`pwd`/archs/macos.sh CXX=`pwd`/archs/macos.sh CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -tags with_v8 -trimpath -ldflags "-s -w" -o build/phoenixbuilder-v8-macos-arm64
build/phoenixbuilder-v8-macos: build/ build/phoenixbuilder-v8-macos-x86_64 build/phoenixbuilder-v8-macos-arm64 ${SRCS_GO}
	${LIPO} -create build/phoenixbuilder-v8-macos-x86_64 build/phoenixbuilder-v8-macos-arm64 -output build/phoenixbuilder-v8-macos
build/phoenixbuilder-android-executable-armv7: build/ ${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/armv7a-linux-androideabi16-clang ${SRCS_GO}
	CGO_CFLAGS=${CGO_DEF} CC=${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/armv7a-linux-androideabi16-clang GOOS=android GOARCH=arm CGO_ENABLED=1 go build -trimpath -ldflags "-s -w" -o build/phoenixbuilder-android-executable-armv7
build/phoenixbuilder-v8-android-executable-arm64: build/ ${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/aarch64-linux-android21-clang ${SRCS_GO}
	CGO_CFLAGS=${CGO_DEF}" -DWITH_V8" CC=${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/aarch64-linux-android21-clang CXX=${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/aarch64-linux-android21-clang++ GOOS=android GOARCH=arm64 CGO_ENABLED=1 go build -tags with_v8 -trimpath -ldflags "-s -w -extldflags -static-libstdc++" -o build/phoenixbuilder-v8-android-executable-arm64
build/phoenixbuilder-android-executable-arm64: build/ ${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/aarch64-linux-android21-clang ${SRCS_GO}
	CGO_CFLAGS=${CGO_DEF} CC=${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/aarch64-linux-android21-clang CXX=${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/aarch64-linux-android21-clang++ GOOS=android GOARCH=arm64 CGO_ENABLED=1 go build -trimpath -ldflags "-s -w -extldflags -static-libstdc++" -o build/phoenixbuilder-android-executable-arm64
build/phoenixbuilder-android-executable-x86: build/ ${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/i686-linux-android21-clang ${SRCS_GO}
	CGO_CFLAGS=${CGO_DEF} CC=${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/i686-linux-android21-clang GOOS=android GOARCH=386 CGO_ENABLED=1 go build -trimpath -ldflags "-s -w" -o build/phoenixbuilder-android-executable-x86
build/phoenixbuilder-android-executable-x86_64: build/ ${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/x86_64-linux-android21-clang ${SRCS_GO}
	CGO_CFLAGS=${CGO_DEF} CC=${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/x86_64-linux-android21-clang GOOS=android GOARCH=amd64 CGO_ENABLED=1 go build -trimpath -ldflags "-s -w" -o build/phoenixbuilder-android-executable-x86_64
build/phoenixbuilder-windows-executable-x86.exe: build/ /usr/bin/i686-w64-mingw32-gcc ${SRCS_GO}
	CGO_CFLAGS=${CGO_DEF} CC=/usr/bin/i686-w64-mingw32-gcc GOOS=windows GOARCH=386 CGO_ENABLED=1 go build -trimpath -ldflags "-s -w" -o build/phoenixbuilder-windows-executable-x86.exe
build/phoenixbuilder-windows-executable-x86_64.exe: build/ /usr/bin/x86_64-w64-mingw32-gcc ${SRCS_GO}
	CGO_CFLAGS=${CGO_DEF} CC=/usr/bin/x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build -trimpath -ldflags "-s -w" -o build/phoenixbuilder-windows-executable-x86_64.exe
build/phoenixbuilder-openwrt-mt7620-mipsel_24kc: build/ openwrt-sdk-21.02.2-ramips-mt7620_gcc-8.4.0_musl.Linux-x86_64/ ${SRCS_GO}
	CGO_CFLAGS=${CGO_DEF} CC=`pwd`/openwrt-sdk-21.02.2-ramips-mt7620_gcc-8.4.0_musl.Linux-x86_64/staging_dir/toolchain-mipsel_24kc_gcc-8.4.0_musl/bin/mipsel-openwrt-linux-gcc CXX=`pwd`/openwrt-sdk-21.02.2-ramips-mt7620_gcc-8.4.0_musl.Linux-x86_64/staging_dir/toolchain-mipsel_24kc_gcc-8.4.0_musl/bin/mipsel-openwrt-linux-g++ GOARCH=mipsle CGO_ENABLED=1 go build -trimpath -tags no_readline -ldflags "-s -w" -o build/phoenixbuilder-openwrt-mt7620-mipsel_24kc
#build/phoenixbuilder-v8-windows-executable-x86_64.exe: build/ /usr/bin/x86_64-w64-mingw32-gcc ${SRCS_GO}
#	CGO_CFLAGS=${CGO_DEF}" -DWITH_V8" CC=/usr/bin/x86_64-w64-mingw32-gcc CXX=/usr/bin/x86_64-w64-mingw32-g++ GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build -tags with_v8 -trimpath -ldflags "-s -w" -o build/phoenixbuilder-v8-windows-executable-x86_64.exe
#build/phoenixbuilder-windows-shared.dll: build/ /usr/bin/x86_64-w64-mingw32-gcc ${SRCS_GO}
#	CGO_CFLAGS="-Wl,--enable-stdcall-fixup -luser32 -lcomdlg32 -Wno-pointer-to-int-cast -mwindows -m64 -march=x86-64 -luser32 -lkernel32 -lgdi32 -lwinmm -lcomctl32 -ladvapi32 -lshell32 -lpsapi -nodefaultlibs -nostdlib -lmsvcrt -D_UCRT=1" CC=/usr/bin/x86_64-w64-mingw32-gcc CGO_LDFLAGS="--enable-stdcall-fixup" GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build -buildmode=c-shared -trimpath -o build/phoenixbuilder-windows-shared.dll
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
	dpkg -b release/phoenixbuilder-iphoneos release/
package/android: package/android-armv7 package/android-arm64
package/android-armv7: build/phoenixbuilder-android-executable-armv7 release/
	mkdir -p release/phoenixbuilder-android-armv7/data/data/com.termux/files/usr/bin release/phoenixbuilder-android-armv7/DEBIAN
	cp build/phoenixbuilder-android-executable-armv7 release/phoenixbuilder-android-armv7/data/data/com.termux/files/usr/bin/fastbuilder
	printf "Package: pro.fastbuilder.phoenix-android\n\
	Name: FastBuilder Phoenix (Alpha)\n\
	Version: $(VERSION)\n\
	Architecture: arm\n\
	Maintainer: Ruphane\n\
	Author: Bouldev <admin@boul.dev>\n\
	Section: Games\n\
	Priority: optional\n\
	Homepage: https://fastbuilder.pro\n\
	Description: Modern Minecraft structuring tool\n" > release/phoenixbuilder-android-armv7/DEBIAN/control
	dpkg -b release/phoenixbuilder-android-armv7 release/
package/android-arm64: build/phoenixbuilder-android-executable-arm64 release/
	mkdir -p release/phoenixbuilder-android-arm64/data/data/com.termux/files/usr/bin release/phoenixbuilder-android-arm64/DEBIAN
	cp build/phoenixbuilder-android-executable-arm64 release/phoenixbuilder-android-arm64/data/data/com.termux/files/usr/bin/fastbuilder
	printf "Package: pro.fastbuilder.phoenix-android\n\
	Name: FastBuilder Phoenix (Alpha)\n\
	Version: $(VERSION)\n\
	Architecture: aarch64\n\
	Maintainer: Ruphane\n\
	Author: Bouldev <admin@boul.dev>\n\
	Section: Games\n\
	Priority: optional\n\
	Homepage: https://fastbuilder.pro\n\
	Description: Modern Minecraft structuring tool\n" > release/phoenixbuilder-android-arm64/DEBIAN/control
	dpkg -b release/phoenixbuilder-android-arm64 release/
clean:
	rm -f build/phoenixbuilder*
