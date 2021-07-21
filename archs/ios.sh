#! /bin/bash
export IPHONEOS_DEPLOYMENT_TARGET=9.3
exec $THEOS/toolchain/linux/iphone/bin/clang -F$THEOS/vendor/lib -target arm64-apple-ios9.0 -target arm64-apple-darwin -isysroot $THEOS/sdks/iPhoneOS14.0.sdk $@