#!/bin/bash
export IPHONEOS_DEPLOYMENT_TARGET=12.0
if [ $(uname) == "Darwin" ] && [ $(uname -m | grep -E "x86_64|arm64" > /dev/null ; echo ${?}) -eq 0 ]
then
  exec "$(xcrun --sdk iphoneos -f clang)" -target arm64-apple-ios12.0 -mios-version-min=12.0 -isysroot "$(xcrun --sdk iphoneos --show-sdk-path)" $@
elif [ $(uname) == "Darwin" ] && [ $(uname -m | grep -E "iPhone|iPad|iPod" > /dev/null ; echo ${?}) -eq 0 ]
then
  exec /usr/bin/clang -target arm64-apple-ios12.0 $@
else
  if [ -e $THEOS/sdks/iPhoneOS12.2.sdk ]; then
    export SDK_PATH=$THEOS/sdks/iPhoneOS12.2.sdk
  elif [ -e $THEOS/sdks/iPhoneOS12.4.sdk ]; then
    export SDK_PATH=$THEOS/sdks/iPhoneOS12.4.sdk
  else
    echo SDK not found
    exit 1
  fi
  exec $THEOS/toolchain/linux/iphone/bin/clang -F$THEOS/vendor/lib -target arm64-apple-ios12.0 -mios-version-min=12.0 -isysroot $SDK_PATH -F$SDK_PATH/System/Library/PrivateFrameworks -framework Foundation -framework CoreFoundation -framework UIKit -framework CoreGraphics -framework CoreUI -framework AVFoundation -Wno-unused-command-line-argument -lSystem $@
fi
