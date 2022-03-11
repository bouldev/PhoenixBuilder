#!/bin/bash
git clone --recursive https://github.com/theos/theos.git
curl -LO https://github.com/sbingner/llvm-project/releases/latest/download/linux-ios-arm64e-clang-toolchain.tar.lzma
curl -LO https://github.com/theos/sdks/archive/master.zip
curl -LO https://github.com/phracker/MacOSX-SDKs/releases/download/11.3/MacOSX11.0.sdk.tar.xz