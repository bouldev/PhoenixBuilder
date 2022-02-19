#!/bin/bash
export IPHONEOS_DEPLOYMENT_TARGET=9.3
if [ $(uname) == "Darwin" ] && [ $(uname -m | grep -E "x86_64|arm64" > /dev/null ; echo ${?}) -eq 0 ]
then
  exec "$(xcrun --sdk iphoneos -f clang)" -target arm64-apple-ios9.0 -target arm64-apple-darwin -isysroot "$(xcrun --sdk iphoneos --show-sdk-path)" $@
elif [ $(uname) == "Darwin" ] && [ $(uname -m | grep -E "iPhone|iPad|iPod" > /dev/null ; echo ${?}) -eq 0 ]
then
  exec /usr/bin/clang -target arm64-apple-ios9.0 -target arm64-apple-darwin $@
else
  exec $THEOS/toolchain/linux/iphone/bin/clang -F$THEOS/vendor/lib -target arm64-apple-ios14.0 -target arm64-apple-darwin -isysroot $THEOS/sdks/iPhoneOS14.4.sdk $@
fi
