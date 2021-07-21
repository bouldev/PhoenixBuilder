all: build current ios-executable ios-lib android-executable-v7 android-executable-64 windows-executable
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
	${THEOS}/toolchain/linux/iphone/bin/ldid -S build/phoenixbuilder-ios-executable
ios-lib:
	rm -f build/phoenixbuilder-ios-lib.a
	CC=`pwd`/archs/ios.sh CGO_ENABLED=1 GOOS=ios GOARCH=arm64 go build -buildmode=c-archive -trimpath -ldflags "-s -w" -o build/phoenixbuilder-ios-static.a
	#node symbolkiller.js build/phoenixbuilder-ios-static.a
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