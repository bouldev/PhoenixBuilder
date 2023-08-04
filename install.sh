#!/bin/bash

# Bouldev 2023
# This script is for auto selecting PhoenixBuilder release prebuilts,
# not for native compiling.
#
# If you did not find any matched release version for your
# operating systems or machines, please contact us by email:
# <support at boul dot dev>

# Planned support: macOS, iOS, Android, Linux (Debian, Ubuntu)
#=============================================================#

# Define functions to exit properly
# This is designed to delete temp files after script ends
trap ctrl_c INT

function quit_installer() {
  rm -rf "${PREFIX}"/./fastbuilder-temp
  exit "${1}"
}

function ctrl_c() {
  printf "\n\033[33mUser forced to exit, performing cleanup steps...\033[0m\n"
  quit_installer 1
}

# Start
SCRIPT_VERSION="0.0.3"
printf "\033[33mFastBuilder Phoenix Installer v%s\033[0m\n" "${SCRIPT_VERSION}"
printf "\033[33mBouldev 2022, Copyrighted.\033[0m\n"
printf "\033[32mStarting installation progress...\033[0m\n"

# Some distro does not provide `which` by default
WHICH_CMD=""
for which_prog in "which" "which.debianutils" "command"; do
  if [[ $which_prog == "command" ]]; then
    if command -v apt >> /dev/null 2>&1; then
      WHICH_CMD="command -v"
    fi
  else
    # Try to search for themself by emself
    if $which_prog $which_prog >> /dev/null 2>&1; then
      WHICH_CMD="$which_prog"
      break
    fi
  fi
done
if [[ "${WHICH_CMD}" == "" ]]; then
  printf "\033[33mWarning: Unable to identify absolute location of executables.\033[0m\n"
fi

# Check whether uname(1) GNU or BSD
UNAME_GET_OSNAME="uname -s"
for uname_prog in "uname" "guname"; do
  ${WHICH_CMD} ${uname_prog} > /dev/null 2>&1
  if [ $? == 0 ]; then
    if [ $(${uname_prog} --version &> /dev/null; echo $?) == 0 ]; then
      UNAME_GET_OSNAME="${uname_prog} -o"
    fi
  fi
done

# TODO: Don't do tmps in prefix, use mktemp(1)
# Check permissions and prefix
echo "Checking permissions..."
if [ "${DESTDIR}" ]; then
  printf "\033[33mFound DESTDIR: %s\033[0m\n" "${DESTDIR}"
  PREFIX="${DESTDIR}"
elif [ "${PREFIX}" ]; then
  printf "\033[33mFound prefix preset in your environment: %s\033[0m\n" "${PREFIX}"
else
  PREFIX="/usr/local"
fi
BINDIR="${PREFIX}/bin"
ROOT_REQUIRED="1"
if [ ${LOCAL} ]; then
  printf "\033[32mUser required to run install script with non-privileged access\033[0m\n"
  printf "A folder named \"fastbuilder\" will be created under %s\n" "${HOME}"
  PREFIX="${HOME}/fastbuilder"
  ROOT_REQUIRED="0"
elif [[ $(${UNAME_GET_OSNAME}) == "Android" ]] && [[ $(apt install &> /dev/null; echo $?) == 0 ]]; then
  # No need of root on Termux
  printf "\033[32mRunning under Android Termux (APT does not require root)\033[0m\n"
  ROOT_REQUIRED="1"
elif [[ $(${UNAME_GET_OSNAME}) == "Android" ]] && [[ $(apt install &> /dev/null; echo $?) != 0 ]]; then
  # What happend?
  printf "\033[31mAPT broken!\033[0m\n"
  printf "\033[31mPlease reconfigure your APT/dpkg to fix current existing problems by running\033[0m\n"
  printf "\033[33m  dpkg --configure -a\033[0m\n"
  printf "\033[31mOr prepend LOCAL=1 before command\033[0m\n"
  printf "\033[31mTo install FastBuilder without APT access.\033[0m\n"
  quit_installer 1
elif [[ $(id -u) == 0 ]]; then
  printf "\033[31mWARNING: Is is not recommended to install things by scripts in a normal *nix, they may mess up your environment.\033[0m\n"
  if [ ${SUDO_UID} ]; then
    printf "\033[32mRunning under sudo privileges\033[0m\n"
  else
    printf "\033[33mIt is dangerous to run under root directly, but the\033[0m"
    printf "\033[33m install script would proceed anyway (sudo suggested).\033[0m\n"
  fi
else
  printf "\033[31mRoot privilege required!\033[0m\n"
  printf "\033[31mPlease run this installer under root permission\033[0m\n"
  printf "\033[31mIs is not recommended to install things by scripts in a normal *nix, they may mess up your environment.\033[0m\n"
  printf "\033[31mOr prepend LOCAL=1 before command\033[0m\n"
  printf "\033[31mTo install FastBuilder without root access.\033[0m\n"
  quit_installer 1
fi

# Basic information
echo "Fetching basic info..."
SYSTEM_NAME=$(uname)
KERNEL_VERSION=$(uname -r)
# The reason we do not use "uname -m"/"uname -p" to identify arch
# is that they may return unexpected values.
# e.g. "uname -m" returns device model name when on iOS
arch_format() {
  ${WHICH_CMD} arch > /dev/null 2>&1
  if [ $? == 0 ]; then
    ARCH=$(arch)
  else
    ARCH=$(uname -m)
  fi

  if [ $(echo ${ARCH} | grep -E "armv8|aarch64" &> /dev/null; echo $?) == 0 ]; then
    ARCH="arm64"
  elif [ $(echo ${ARCH} | grep -E "x64|amd64" &> /dev/null; echo $?) == 0 ]; then
    ARCH="x86_64"
  elif [ $(echo ${ARCH} | grep -E "386|586|686" &> /dev/null; echo $?) == 0 ]; then
    ARCH="x86"
  elif [[ ${ARCH} == "arm" ]] || [[ ${ARCH} == "arm32" ]]; then
    ARCH="armv7"
  fi
  printf ${ARCH}
}
machine_format() {
  MACHINE=$(uname -m)
  if [ $(echo ${MACHINE} | grep -E "armv4|armv5|armv6|armv7" &> /dev/null; echo $?) == 0 ]; then
    MACHINE="arm"
  elif [ $(echo ${MACHINE} | grep -E "386|586|686" &> /dev/null; echo $?) == 0 ]; then
    MACHINE="x86"
  fi
  printf ${MACHINE}
}
ARCH="$(arch_format)"
MACHINE="$(machine_format)"
# Darwin's uname is not reliable, using sw_vers to identify device family if possible
if [ ${SYSTEM_NAME} == "Darwin" ]; then
  ${WHICH_CMD} sw_vers > /dev/null 2>&1
  if [ $? == 0 ]; then
    if [ "$(sw_vers -productName)" == "macOS" ]; then
      MACHINE="macos"
    elif [ "$(sw_vers -productName)" == "iPhone OS" ]; then
      MACHINE="ios"
    else
      printf "\033[31mUnknown Darwin Product %s!\033[0m\n" "$(sw_vers -productName)"
      printf "\033[31mPlease report this issue under \033[0m"
      printf "\033[33mhttps://github.com/LNSSPsd/PhoenixBuilder/issues\033[0m\n"
      exit 1
    fi
  else
    printf "\033[31mRequired command sw_vers(1) not found\033[0m\n"
    printf "\033[31mUsing uname(1) for guessing (That's terrible)\033[0m\n"
    if [ $(echo ${MACHINE} | grep -E "iPhone|iPad|iPod"; echo $?) == 0 ]; then
      MACHINE="ios"
    else
      MACHINE="macos"
    fi
  fi
fi
echo "Your device and OS: ${SYSTEM_NAME} ${KERNEL_VERSION}, ${ARCH}"

# Check if any CLI tools that can be used to download files
# Use cURL by default
echo "Finding downloaders..."
DL_TOOL=""
DL_TOOL_NAME=""
DL_TOOL_OUT_FLAG="-o"
for i in "curl" "wget" "axel" "aria2c"; do
  ${WHICH_CMD} ${i} > /dev/null 2>&1
  if [ $? == 0 ]; then
    echo "Found ${i}: $(which ${i})"
    DL_TOOL=$(which ${i})
    DL_TOOL_NAME="${i}"
    break
  fi
done
if [ ${DL_TOOL} == "" ]; then
  printf "\033[31mInstall curl before using this script!\033[0m\n"
  exit 1
elif [ ${DL_TOOL_NAME} == "wget" ]; then
  DL_TOOL_OUT_FLAG="-O"
elif [ ${DL_TOOL_NAME} == "curl" ]; then
  DL_TOOL_OUT_FLAG="-fSL -o"
fi

# Check if "install" command exists
INSTALL=""
# GNU install is preferred, BSD install is okay though
# On macOS, GNU install were installed using brew with name "ginstall"
for i in "ginstall" "install"; do
  ${WHICH_CMD} ${i} >/dev/null 2>&1
  if [ $? == 0 ]; then
    printf "\033[32mFastBuilder will be installed by using ${i}: \033[0m"
    printf "\033[32m$(${WHICH_CMD} ${i})\033[0m\n"
    INSTALL="${i} -m 0755"
    break
  fi
done
if [ "${INSTALL}" == "" ]; then
  printf "\033[33mThis script prefers to install files by using \033[0m"
  printf "\033[33mGNU/BSD install(1) but you do not have it. Skipping.\033[0m"
  INSTALL="cp -f"
fi

printf "\033[32mAll basic checks complete! Proceeding the installation...\033[0m\n"

# FastBuilder Presets
# You should not change these contents
FB_DOMAIN="https://storage.fastbuilder.pro/"
FB_LOCATION_ROOT=""
FB_PREFIX="phoenixbuilder"
FB_LINK="${FB_DOMAIN}${FB_LOCATION_ROOT}${FB_PREFIX}"
FB_VER=""

# Github Releases download source presets
# Do not use mirror as default, let users choose their own
# The environment variables here are the default and can be overridden by the environment variables set by the export command
GH_DOMAIN=${GH_DOMAIN:="https://github.com"}
GH_USER=${GH_USER:="LNSSPsd"}
GH_REPO=${GH_REPO:="PhoenixBuilder"}
GH_RELEASE_URL=${GH_RELEASE_URL:="releases/download/"}
GH_LINK=${GH_LINK:="${GH_DOMAIN}/${GH_USER}/${GH_REPO}/${GH_RELEASE_URL}"}

# Further system detection
FILE_TYPE=""
FILE_ARCH=""

BINARY_INSTALL="0"

if [[ ${SYSTEM_NAME} == "Linux" ]] && [[ $(${UNAME_GET_OSNAME}) == "Android" ]]; then
  # We do not provide .deb packages for Android X86 currently
  if [[ ${ROOT_REQUIRED} == "1" ]] && [[ ${ARCH} != "x86" ]] && [[ ${ARCH} != "x86_64" ]] && [[ $(dpkg --version &> /dev/null; echo $?) == 0 ]]; then
    if [[ $(dpkg -L pro.fastbuilder.phoenix-android &> /dev/null; echo $?) == 0 ]]; then
      FB_VER=$(dpkg-query --showformat='${Version}' --show pro.fastbuilder.phoenix-android)
      printf "\033[32mFound previously installed FastBuilder, Version: ${FB_VER}\033[0m\n"
    fi
    # Terrible workaround for previous corrupted releases, try not hurt packmans
    apt remove pro.fastbuilder.phoenix-android -y
    echo "Requesting FastBuilder Phoenix for Android ${ARCH}..."
    FB_PREFIX="pro.fastbuilder.phoenix-android"
    FILE_TYPE=".deb"
    # Termux armv8l should run armv7 binaries (That's weird?)
    if [[ $(echo ${ARCH} | grep "arm64" &> /dev/null; echo $?) == 0 ]] && [[ $(dpkg --print-architecture) != "arm" ]]; then
      FILE_ARCH="aarch64"
    else
      FILE_ARCH="arm"
    fi
  else
    # Weird error, some Android may not using Termux and then dpkg is something nonexist
    printf "\033[31mFastBuilder cannot provide .deb for your ${ARCH} Android! Requesting binary executables.\033[0m\n"
    FB_PREFIX="phoenixbuilder-android-executable-"
    FILE_TYPE=""
    FILE_ARCH="${ARCH}"
    BINARY_INSTALL="1"
  fi
elif [ ${MACHINE} == "ios" ]; then
  if [[ ${ROOT_REQUIRED} == "1" ]]; then
    if [[ $(dpkg -L pro.fastbuilder.phoenix &> /dev/null; echo $?) == "0" ]]; then
      FB_VER=$(dpkg-query --showformat='${Version}' --show pro.fastbuilder.phoenix)
      printf "\033[32mFound previously installed FastBuilder, Version: ${FB_VER}\033[0m\n"
    fi
    printf "\033[32mIt is suggested to upgrade FastBuilder from your package manager (Cydia, Sileo, etc.).\033[0m\n"
    printf "\033[32mBut I don't care, proceeding...\033[0m\n"
    echo "Requesting FastBuilder Phoenix for ${ARCH} iOS..."
    FB_PREFIX="pro.fastbuilder.phoenix"
    FILE_TYPE=".deb"
    # iOS does not separate architectures, iphoneos-arm for all
    FILE_ARCH="iphoneos-arm"
  elif [ ${ARCH} != "arm64" ]; then
    printf "\033[31mFastBuilder no longer support ${ARCH} iOS! Stopping.\033[0m\n"
    exit 1
  elif [[ $(dpkg --version &> /dev/null; echo $?) != 0 ]] || [[ ${ROOT_REQUIRED} != "1" ]]; then
    printf "\033[32mWe can't call your Debian Packager, Requesting binary executables.\033[0m\n"
    FB_PREFIX="phoenixbuilder-ios-executable"
    FILE_TYPE=""
    FILE_ARCH=""
    BINARY_INSTALL="1"
  fi
elif [ ${MACHINE} == "macos" ]; then
  # Fat Mach-O contains multiple arches, and yes we did that
  if [[ ${ARCH} == "arm64" ]] || [[ ${ARCH} == "x86_64" ]]; then
    echo "Requesting FastBuilder Phoenix for ${ARCH} macOS..."
    FB_PREFIX="phoenixbuilder-macos"
    FILE_TYPE=""
    FILE_ARCH=""
    BINARY_INSTALL="1"
  else
    printf "\033[31mFastBuilder no longer support ${ARCH} macOS! Stopping.\033[0m\n"
    exit 1
  fi
elif [[ ${SYSTEM_NAME} == "NetBSD" ]] || [[ ${SYSTEM_NAME} == "FreeBSD" ]] || [[ ${SYSTEM_NAME} == "OpenBSD" ]]; then
  echo           "If you met 404 error in further downloading, report it at"
  printf "\033[32m  https://github.com/LNSSPsd/PhoenixBuilder/issues\033[0m\n"
  FB_PREFIX="phoenixbuilder-$(echo ${SYSTEM_NAME} | tr '[:upper:]' '[:lower:]')-executable-"
  FILE_TYPE=""
  FILE_ARCH="${ARCH}"
  BINARY_INSTALL="1"
elif [[ ${SYSTEM_NAME} == "Linux" ]] && [[ $(${UNAME_GET_OSNAME}) != "Android" ]]; then
  # Finally, Linux
  echo     "NOTE: We only provide x86_64 and arm64 executables currently, if"
  echo     "      you need prebuilts for other architectures, issue at"
  printf "\033[32mhttps://github.com/LNSSPsd/PhoenixBuilder/issues\033[0m\n"
  if [[ ${ARCH} != "x86_64" ]] && [[ ${ARCH} == "arm64" ]]; then
    FB_PREFIX="phoenixbuilder-"
    FILE_ARCH="aarch64"
  elif [[ ${ARCH} != "x86_64" ]] && [[ ${ARCH} != "arm64" ]]; then
    FB_PREFIX="phoenixbuilder-"
    FILE_ARCH="${ARCH}"
  fi
  BINARY_INSTALL="1"
fi

# TODO: Put all trash in $TMPDIR
# Download now
if [[ ${MACHINE} == "ios" ]] && [[ ${ROOT_REQUIRED} == "1" ]]; then
  # Install APT source for iOS. This would allow users to upgrade FastBuilder from Cydia
  if [ $(grep "apt.boul.dev" -rl /etc/apt/sources.list.d &> /dev/null; echo $?) != 0 ]; then
    printf "\033[32mAdding apt.boul.dev to your repo list...\033[0m\n"
    echo "deb https://apt.boul.dev/ ./" > /etc/apt/sources.list.d/apt.boul.dev.list
  else
    printf "\033[32mUser already added apt.boul.dev to repo list.\033[0m\n"
  fi
fi

rm -rf "${PREFIX}"/./fastbuilder-temp "${BINDIR}"/./fastbuilder "${HOME}"/./fastbuilder
mkdir -p "${PREFIX}"/./fastbuilder-temp
LAUNCH_CMD=""

report_error() {
  if [ ${DL_TOOL_NAME} == "curl" ]; then
    if [ ${1} == 22 ]; then
      printf "\033[031mDownload failure! Requested resources not exist! (curl: 22)\033[0m\n"
      printf "\033[031m ${FB_LINK}\033[0m\n"
    elif [ ${1} == 3 ]; then
      printf "\033[031mURL malformed. (curl: 3)\033[0m\n"
      printf "\033[031mPlease report this bug!\033[0m\n"
    elif [ ${1} == 23 ]; then
        printf "\033[031mCould not write data to local filesystem! (curl: 23)\033[0m\n"
        printf "\033[031mCheck your r/w permissions before the installation.\033[0m\n"
    else
        printf "\033[031mDownload failure! Please check your connections (curl: ${DL_RET}).\nStopping.\033[0m\n"
    fi
  elif [ ${DL_TOOL_NAME} == "wget" ]; then
    if [ ${1} == 1 ]; then
      printf "\033[031mGeneric error (wget: 1)\nTry using curl?\033[0m\n"
    elif [ ${1} == 2 ]; then
      printf "\033[031mParse error, check your .wgetrc and .netrc (wget: 2)\033[0m\n"
    elif [ ${1} == 3 ]; then
      printf "\033[031mFile I/O error (wget: 3)\033[0m\n"
      printf "\033[031mCheck your r/w permissions before the installation.\033[0m\n"
    elif [ ${1} == 8 ]; then
      printf "\033[031mDownload failure! Requested resources not exist! (wget: 8)\033[0m\n"
      printf "\033[031m ${FB_LINK}\033[0m\n"
    else
      printf "\033[031mDownload failure! Please check your connections (wget: ${1}).\nStopping.\033[0m\n"
    fi
  elif [ ${DL_TOOL_NAME} == "aria2c" ]; then
    if [ ${1} == 1 ]; then
      printf "\033[031mUnknown error occurred (aria2c: 1)\nTry using curl?\033[0m\n"
    elif [ ${1} == 3 ]; then
      printf "\033[031mDownload failure! Requested resources not exist! (aria2c: 3)\033[0m\n"
      printf "\033[031m ${FB_LINK}\033[0m\n"
    elif [ ${1} == 9 ]; then
      printf "\033[031mDisk space not enough. (aria2c: 9)\nCleanup spaces before the installation!\033[0m\n"
    elif [[ ${1} == 15 ]] || [[ ${1} == 16 ]] || [[ ${1} == 17 ]] || [[ ${1} == 18 ]]; then
      printf "\033[031mCould not open/create file or directory (aria2c: ${1})\033[0m\n"
      printf "\033[031mCheck your r/w permissions before the installation.\033[0m\n"
    else
      printf "\033[031mDownload failure! Please check your connections (aria2c: ${1}).\nStopping.\033[0m\n"
    fi
  elif [ ${DL_TOOL_NAME} == "axel" ]; then
    if [ ${1} == 1 ]; then
      printf "\033[031mSomething went wrong (axel: 1)\nTry using curl?\033[0m\n"
    else
      printf "\033[031mDownload failure! Please check your connections (axel: ${DL_RET}).\nStopping.\033[0m\n"
    fi
  else
    printf "\033[031mDownload failure! (${DL_TOOL}: ${DL_RET}).\nStopping.\033[0m\n"
  fi
  quit_installer 1
}

# Download a file contains the latest version num for FastBuilder distros
printf "Getting latest version of FastBuilder...\n"
FB_VERSION_LINK="${FB_DOMAIN}${FB_LOCATION_ROOT}/version"
if [[ ${PB_USE_GH_REPO} == "1" ]]; then
  FB_VERSION_LINK="${GH_DOMAIN}/${GH_USER}/${GH_REPO}/raw/main/version"
fi
${DL_TOOL} ${DL_TOOL_OUT_FLAG} "${PREFIX}"/./fastbuilder-temp/version ${FB_VERSION_LINK}
DL_RET=$?
if [ ${DL_RET} == 0 ]; then
  FB_VER=$(cat "${PREFIX}"/./fastbuilder-temp/version | sed -n -e 'H;${x;s/\n//g;p;}')
  printf "${FB_VER}\n"
else
  report_error ${DL_RET}
fi

if [[ ${BINARY_INSTALL} == "1" ]]; then
  printf "Downloading FastBuilder binary...\n"
  # Repeat FB_LINK
  FB_LINK="${FB_DOMAIN}${FB_LOCATION_ROOT}${FB_PREFIX}${FILE_ARCH}${FILE_TYPE}"
  if [[ ${PB_USE_GH_REPO} == "1" ]]; then
    printf "\033[32mOriginal download link: ${FB_LINK}\033[0m\n"
    FB_LINK="${GH_LINK}v${FB_VER}/${FB_PREFIX}${FILE_ARCH}${FILE_TYPE}"
    printf "\033[32mGithub download link: ${FB_LINK}\033[0m\n"
  fi
  printf "\033[33mIf the official storage does not work for you, you can try to assign environment variable \"PB_USE_GH_REPO=1\" for the script to download stuff from Github.\033[0m\n"
  ${DL_TOOL} ${DL_TOOL_OUT_FLAG} "${PREFIX}/./fastbuilder-temp/fastbuilder" "${FB_LINK}"
  DL_RET=$?
  if [ ${DL_RET} == 0 ]; then
    printf "\033[32mSuccessfully downloaded FastBuilder\033[0m"
    if [ ${MACHINE} == "macos" ]; then
      printf "\033[32m (Universal)\033[0m\n"
    elif [ ${MACHINE} == "ios" ]; then
      printf "\033[32m for iOS\033[0m\n"
    else
      printf "\033[32m (${ARCH})\033[0m\n"
    fi
  else
    report_error ${DL_RET}
  fi
  # Explicitly perform chmod
  chmod +x "${PREFIX}"/./fastbuilder-temp/fastbuilder
  if [ ${ROOT_REQUIRED} == "1" ]; then
    ${INSTALL} "${PREFIX}"/./fastbuilder-temp/fastbuilder ${BINDIR}
    LAUNCH_CMD="fastbuilder"
  else
    ${INSTALL} "${PREFIX}"/./fastbuilder-temp/fastbuilder "${PREFIX}"/
    LAUNCH_CMD="${PREFIX}/fastbuilder"
  fi
else
  printf "Downloading FastBuilder package...\n"
  # Repeat FB_LINK
  FB_LINK="${FB_DOMAIN}${FB_LOCATION_ROOT}${FB_PREFIX}_${FB_VER}_${FILE_ARCH}${FILE_TYPE}"
  if [[ ${PB_USE_GH_REPO} == "1" ]]; then
    printf "\033[32mOriginal download link: ${FB_LINK}\033[0m\n"
    FB_LINK="${GH_LINK}v${FB_VER}/${FB_PREFIX}_${FB_VER}_${FILE_ARCH}${FILE_TYPE}"
    printf "\033[32mGithub download link: ${FB_LINK}\033[0m\n"
  else
    printf "\033[33mIf the official storage does not work for you, you can try to assign environment variable \"PB_USE_GH_REPO=1\" for the script to download stuff from Github.\033[0m\n"
  fi
  ${DL_TOOL} ${DL_TOOL_OUT_FLAG} "${PREFIX}"/./fastbuilder-temp/fastbuilder.deb ${FB_LINK}
  DL_RET=$?
  if [ ${DL_RET} == 0 ]; then
    printf "\033[32mSuccessfully downloaded FastBuilder\033[0m\n"
  else
    report_error ${DL_RET}
  fi
  # When install.sh have root privileges, it will install packages directly through dpkg
  if [ ${ROOT_REQUIRED} == "1" ]; then
    printf "Installing deb package...\n"
    dpkg -i "${PREFIX}"/./fastbuilder-temp/fastbuilder.deb
    if [ $? != 0 ]; then
      printf "\033[31mSome errors occurred when calling Debian Packager.\nYou may want to run \"dpkg --configure -a\" to fix some problems.\033[0m\n"
      quit_installer 1
    fi
    LAUNCH_CMD="fastbuilder"
  else
    mv "${PREFIX}"/./fastbuilder-temp/fastbuilder.deb "${PREFIX}"/./fastbuilder.deb
    printf "\033[32mUnprevileged, A deb file has been downloaded at %s/fastbuilder.deb\033[0m\n" "${PREFIX}"
    printf "\033[32mManually install it by \"dpkg -i %s/fastbuilder.deb\"" "${PREFIX}\033[0m\n"
    quit_installer 0
  fi
fi

# Yay, everything done!
printf "\033[32mFastBuilder has been successfully installed on your device!\nUse command \"%s\" to launch it.\033[0m\n" "${LAUNCH_CMD}"
quit_installer 0
