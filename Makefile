ifeq ($(shell uname | grep "Darwin" > /dev/null ; echo $${?}),0)
ifeq ($(shell uname -m | grep -E "iPhone|iPad|iPod" > /dev/null ; echo $${?}),0)
IOS_STRIP=/usr/bin/strip
LDID=/usr/bin/ldid
else
IOS_STRIP=$(shell xcrun --sdk iphoneos strip)
LDID=ldid2
endif
else
IOS_STRIP=false
LDID=$${THEOS}/toolchain/linux/iphone/bin/ldid
endif

all: build current ios-executable ios-lib macos android-executable-v7 android-executable-64 windows-executable
build:
	mkdir build
current:
	rm -f build/phoenixbuilder
	CGO_ENABLED=1 go build -trimpath -ldflags "-s -w" -o build/phoenixbuilder
	#node symbolkiller.js build/phoenixbuilder
ios-executable:
	rm -f build/phoenixbuilder-ios-executable
	CC=`pwd`/archs/ios.sh CGO_ENABLED=1 GOOS=ios GOARCH=arm64 go build -trimpath -ldflags "-s -w" -o build/phoenixbuilder-ios-executable
	#node symbolkiller.js build/phoenixbuilder-ios-executable
	$(IOS_STRIP) build/phoenixbuilder-ios-executable
	$(LDID) -S build/phoenixbuilder-ios-executable
ios-lib:
	rm -f build/phoenixbuilder-ios-lib.a
	CC=`pwd`/archs/ios.sh CGO_ENABLED=1 GOOS=ios GOARCH=arm64 go build -buildmode=c-archive -trimpath -ldflags "-s -w" -o build/phoenixbuilder-ios-static.a
	#node symbolkiller.js build/phoenixbuilder-ios-static.a
macos: 
	rm -f build/phoenixbuilder-macos*
	CC=`pwd`/archs/macos.sh CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o build/phoenixbuilder-macos-x86_64
	CC=`pwd`/archs/macos.sh CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags "-s -w" -o build/phoenixbuilder-macos-arm64
	lipo -create build/phoenixbuilder-macos-x86_64 build/phoenixbuilder-macos-arm64 -output build/phoenixbuilder-macos
android-executable-v7:
	rm -f build/phoenixbuilder-android-executable-armv7
	CC=${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/armv7a-linux-androideabi16-clang GOOS=android GOARCH=arm CGO_ENABLED=1 go build -trimpath -ldflags "-s -w" -o build/phoenixbuilder-android-executable-armv7
	#node symbolkiller.js build/phoenixbuilder-android-executable-armv7
android-executable-64:
	rm -f build/phoenixbuilder-android-executable-arm64
	CC=${HOME}/android-ndk-r20b/toolchains/llvm/prebuilt/linux-x86_64/bin/aarch64-linux-android21-clang GOOS=android GOARCH=arm64 CGO_ENABLED=1 go build -trimpath -ldflags "-s -w" -o build/phoenixbuilder-android-executable-arm64
	#node symbolkiller.js build/phoenixbuilder-android-executable-arm64
windows-executable:
	rm -f build/phoenixbuilder-windows-executable.exe
	CC=/usr/bin/i686-w64-mingw32-gcc GOOS=windows GOARCH=386 CGO_ENABLED=1 go build -trimpath -ldflags "-s -w" -o build/phoenixbuilder-windows-executable.exe
clean:
	rm -rf build