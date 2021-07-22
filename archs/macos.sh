#!/bin/bash
export MACOSX_DEPLOYMENT_TARGET=10.12
if [ $(uname) == "Darwin" ] && [ $(uname -m | grep -E "x86_64|arm64" > /dev/null ; echo ${?}) -eq 0 ]
then
  exec "$(xcrun --sdk macosx -f clang)" -target x86_64-apple-darwin -target arm64-apple-darwin -mmacosx-version-min=10.12 -isysroot "$(xcrun --sdk macosx --show-sdk-path)" $@
elif [ $(uname) == "Darwin" ] && [ $(uname -m | grep -E "iPhone|iPad|iPod" > /dev/null ; echo ${?}) -eq 0 ]
then
  exec /usr/bin/clang -target x86_64-apple-darwin -target arm64-apple-darwin -mmacosx-version-min=10.12
else
  exec $THEOS/toolchain/linux/iphone/bin/clang -target x86_64-apple-darwin -target arm64-apple-darwin -mmacosx-version-min=10.12 -isysroot $THEOS/sdks/MacOSX11.0.sdk $@
fi
