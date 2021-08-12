#!/bin/bash

# Bouldev 2021
# This script is for auto select FastBuilder release prebuilts,
# not for native compiling.
#
# If you did not found any matched release version for your
# operating systems or machines, please contact us by email:
# <support at boul dot dev>

#Planned support: macOS, iOS, Android, Linux (Debian, Ubuntu)

#=============================================================#

# I keep meeting bugs when testing this fucked-up script
# Temp disable macOS for now, I will try to fix it later
if [[ $(uname) == "Darwin" ]] && [[ $(uname -m) == "arm64" ]] || [[ $(uname -m) == "x86_64" ]]; then
  echo "macOS not supported by installer at that time!"
  echo "Please wait for updates or install FastBuilder manually."
  exit 1
fi
# Define a function to properly exit
# This were designed to delete temp files after script ends
function quit_installer() {
  rm -r ${PREFIX}/./fastbuilder-temp
  exit ${1}
}

# Start
SCRIPT_VERSION="0.0.1"
printf "\033[33mFastBuilder Phoenix Installer v${SCRIPT_VERSION}\033[0m\n"
printf "\033[33mBouldev 2021, Copyrights Reserved.\033[0m\n"
printf "\033[32mStarting installation progress...\033[0m\n"

# Check permissions and prefix
echo "Checking permissons..."
if [ ${PREFIX} ]; then
  printf "\033[33mFound prefix preset in your environment: ${PREFIX}\033[0m\n"
else
  PREFIX="/usr/local"
fi
BINDIR="${PREFIX}/bin"
ROOT_REQUIRED="1"
if [[ ${1} ]] && [[ ${1} == "local" ]]; then
  printf "\033[32mUser required to run install script with non-root access\033[0m\n"
  printf "A folder named \"fastbuilder\" will be created under ${HOME}\n"
  PREFIX="${HOME}/fastbuilder"
  ROOT_REQUIRED="0"
elif [ $(id -u) == 0 ]; then
  if [ ${SUDO_UID} ]; then
    printf "\033[32mRunning under sudo priviledges\033[0m\n"
  else
    printf "\033[33mIt is dangerous to run under root directly, but the\033[0m"
    printf "\033[33m install script would proceed anyway (sudo suggested).\033[0m\n"
  fi
else
  printf "\033[31mRoot priviledge required!\033[0m\n"
  printf "\033[31mPlease run the installer using this command:\033[0m\n"
  printf "\033[33m  sudo sh ${0}\033[0m\n"
  exit 1
fi

# Basic informations
echo "Fetching basic info..."
MACHINE=$(uname -m)
SYSTEM_NAME=$(uname)
KERNEL_VERSION=$(uname -r)
# The reason we do not use "uname -m"/"uname -p" to identify arch
# is that they may return unexpected values.
# e.g. "uname -m" returns device model name when on iOS
arch_format() {
  ARCH=$(arch)
  if [ $(echo ${ARCH} | grep -E "armv8|aarch64" | echo $?) -eq 0 ]; then
    ARCH="arm64"
  elif [ $(echo ${ARCH} | grep -E "x64|amd64" | echo $?) -eq 0 ]; then
    ARCH="x86_64"
  elif [[ ${ARCH} == "arm" ]] || [[ ${ARCH} = "arm32" ]]; then
    ARCH="armv7"
  fi
  printf ${ARCH}
}
ARCH="$(arch_format)"
echo "Your device and OS: ${SYSTEM_NAME}, ${KERNEL_VERSION} ${ARCH}"

# Check if any CLI tools that can be used to download files
# Use cURL by default
echo "Finding downloaders..."
DL_TOOL=""
for i in "curl" "wget" "axel" "aria2c"; do
  which ${i} >/dev/null 2>&1
  if [ $? -eq 0 ]; then
    echo "Found ${i}: $(which ${i})"
    DL_TOOL=$(which ${i})
    break
  fi
done
if [ ${DL_TOOL} == "" ]; then
  printf "\033[31mInstall curl before using this script!\033[0m\n"
  exit 1
fi

# Check if "install" command exists
INSTALL=""
# GNU install is preferred, BSD install is okay though
# On macOS, GNU install were installed using brew with name "ginstall"
for i in "ginstall" "install"; do
  which ${i} >/dev/null 2>&1
  if [ $? -eq 0 ]; then
    printf "\033[32mFastBuilder will be installed by using ${i}: \033[0m"
    printf "\033[32m$(which ${i})\033[0m\n"
    INSTALL="${i} -m 755"
    break
  fi
done
if [ ${INSTALL} == "" ]; then
  printf "\033[33mThis script prefers to install files by using \033[0m"
  printf "\033[33mGNU/BSD coreutils but you do not have it. Skipping.\033[0m"
  INSTALL="cp -f"
fi

printf "\033[32mAll basic checks complete! Proceeding to install...\033[0m"
# FastBuilder Presets
# You should not change these contents
FB_DOMAIN="https://fastbuilder.pro/"
FB_LOCATION_ROOT="downloads/phoenix/"
FB_SUFFIX="phoenixbuilder"
FB_LINK=${FB_DOMAIN}${FB_LOCATION_ROOT}${FB_SUFFIX}
FB_VER=""

# Further system detection
FILE_TYPE=""
FILE_ARCH=""
if [[ ${SYSTEM_NAME} == "Linux" ]] && [[ $(uname -o) == "Android" ]]; then
  if [[ ${MACHINE} != "arm" ]] && [[ ${MACHINE} != "i386" ]] && [[ $(dpkg -L pro.fastbuilder.phoenix-android | echo $?) -eq 0 ]]; then
    #printf "\033[31mYou have already installed FastBuilder through APT!\nPlease uninstall \"pro.fastbuilder.phoenix-android\" before running this script.\033[0m\n"
    #printf "\033[32mOr, download latest FastBuilder's deb package from the user center.\033[0m\n"
    printf "\033[32mFound previous installed FastBuilder\033[0m\n"
    #exit 1
  elif [ $(echo ${ARCH} | grep -E "arm64|armv7" | echo $?) -eq 0 ]; then
    echo "Downloading FastBuilder Phoenix for Android..."
    FB_SUFFIX="pro.fastbuilder.phoenix-android"
    FILE_TYPE=".deb"
    if [ $(echo ${ARCH} | grep "arm64" | echo $?) -eq 0 ]; then
      FILE_ARCH="aarch64"
    else
      FILE_ARCH="arm"
    fi
  else
    printf "\033[31mFastBuilder no longer support ${ARCH} Android! Stopping.\033[0m\n"
    exit 1
  fi
elif [ $(echo ${MACHINE} | grep -E "iPhone|iPad|iPod" | echo $?) == "0" ]; then
  echo ${MACHINE}
  if [ $(dpkg -L pro.fastbuilder.phoenix | echo $?) == "0" ]; then
    #printf "\033[31mYou have already installed FastBuilder through APT!\nPlease uninstall \"pro.fastbuilder.phoenix\" before running this script.\033[0m\n"
    #printf "\033[32mOr, download latest FastBuilder's deb package from your package manager (Cydia, Sileo, etc.).\033[0m\n"
    printf "\033[32mFound previous installed FastBuilder\033[0m\n"
    #exit 1
  elif [ ${ARCH} != "arm64" ]; then
    printf "\033[31mFastBuilder no longer support ${ARCH} iOS! Stopping.\033[0m\n"
    exit 1
  fi
    echo "Downloading FastBuilder Phoenix for iOS..."
    FB_SUFFIX="pro.fastbuilder.phoenix"
    FILE_TYPE=".deb"
    # iOS does not seperate architectures, iphoneos-arm for all
    FILE_ARCH="iphoneos-arm"
else
  echo fuck
fi

# Download now
if [ ${FILE_ARCH} == "iphoneos-arm" ]; then
  # Install APT source for iOS. This would allow users to upgrade FastBuilder from Cydia
  if [[ $(grep "apt.boul.dev" -rl /etc/apt/sources.list.d | echo $?) -eq 1 ]] && [[ ${ROOT_REQUIRED} == "1" ]]; then
    printf "\033[32mAdding apt.boul.dev to your repo list...\033[0m\n"
    echo "deb https://apt.boul.dev/ ./" >/etc/apt/sources.list.d/apt.boul.dev.list
  else
    printf "\033[32mUser already added apt.boul.dev to repo list.\033[0m\n"
  fi
fi

mkdir -p fastbuilder-temp ${HOME}/./fastbuilder
if [[ ${SYSTEM_NAME} == "Linux" ]] && [[ $(uname -o) != "Android" ]]; then
  # We have not provide Linux distribution packages currently, so binaries only
  printf "Downloading FastBuilder binary..."
  ${DL_TOOL} -o fastbuilder-temp/fastbuilder ${FB_LINK}
  if [ $? -eq 0 ]; then
    printf "\033[32mSuccessfully downloaded FastBuilder (x86_64)\033[0m\n"
  else
    printf "\033[31mDownload failure! Please check your connections.\nStopping.\033[0m\n"
    quit_installer 1
  fi
  if [ ${ROOT_REQUIRED} == "1" ]; then
    ${INSTALL} fastbuilder-temp/fastbuilder ${BINDIR}
  else
    ${INSTALL} fastbuilder-temp/fastbuilder ${PREFIX}/
  fi
elif [[ ${SYSTEM_NAME} == "Darwin" ]] && [[ ${FILE_ARCH} != "iphoneos-arm" ]]; then
  printf "Downloading FastBuilder binary..."
  ${DL_TOOL} -o fastbuilder-temp/fastbuilder ${FB_LINK}-macos
  if [ $? -eq 0 ]; then
    printf "\033[32mSuccessfully downloaded FastBuilder (Universal)\033[0m\n"
  else
    printf "\033[31mDownload failure! Please check your connections.\nStopping.\033[0m\n"
    quit_installer 1
  fi
  if [ ${ROOT_REQUIRED} == "1" ]; then
    ${INSTALL} fastbuilder-temp/fastbuilder ${BINDIR}
  else
    ${INSTALL} fastbuilder-temp/fastbuilder ${PREFIX}/
  fi
else
  # Download a file contains the latest version num for FastBuilder distros
  printf "Getting latest version of FastBuilder..."
  ${DL_TOOL} -o fastbuilder-temp/version ${FB_DOMAIN}${FB_LOCATION_ROOT}version
  if [ $? -eq 0 ]; then
    FB_VER=$(cat fastbuilder-temp/version | sed -n -e 'H;${x;s/\n//g;p;}')
  else
    printf "\033[31mDownload failure! Please check your connections.\nStopping.\033[0m\n"
    quit_installer 1
  fi
  printf "Downloading FastBuilder package...\n"
  ${DL_TOOL} -o fastbuilder-temp/fastbuilder.deb ${FB_LINK}_${FB_VER}_${FILE_ARCH}${FILE_TYPE}
  if [ $? -eq 0 ]; then
    printf "\033[32mSuccessfully downloaded FastBuilder\033[0m\n"
  else
    printf "\033[031Download failure! Please check your connections.\nStopping.\033[0m\n"
    quit_installer 1
  fi
  # When installer.sh have root priviledges, it will install packages directly through dpkg
  # If not, it will unpack it and export the FastBuilder executable to PATH
  if [ ${ROOT_REQUIRED} == "1" ]; then
    printf "Installing deb package...\n"
    dpkg -i fastbuilder-temp/fastbuilder.deb
    if [ $? != "0" ]; then
      printf "\033[31mSome errors occured when calling Debian Packager.\nYou may want to run \"dpkg --configure -a\" to fix some problems.\033[0m\n"
      quit_installer 1
    fi
  else
    printf "Installing FastBuilder to specified path: ${PREFIX}\n"
    mkdir -p ${PREFIX}
    dpkg -x fastbuilder-temp/fastbuilder.deb fastbuilder/
    if [ $(uname -o | grep "Android" | echo $?) -eq 0 ]; then
      mv ${PREFIX}/data/data/com.termux/files/usr/bin/fastbuilder ${PREFIX}/
      # Remember to add a dot in front of the target directory
      # to prevent some unexpected behavior
      rm -r ${PREFIX}/./data
    else
      mv ${PREFIX}/usr/local/bin/fastbuilder ${PREFIX}/
      rm -r ${PREFIX}/./usr
    fi
    if [ $(cat ${HOME}/.profile | grep "export \${HOME}/fastbuilder:\$PATH" | echo $?) -eq 1 ]; then
      echo "Adding ${HOME}/fastbuilder to your \$PATH"
      echo "export \${HOME}/fastbuilder:\$PATH" >${HOME}/.profile
    fi
  fi
fi

# Yay, everything done!
printf "\033[32mFastBuilder has been successfully installed on your device!\nUse command \"fastbuilder\" to launch it.\033[0m\n"
quit_installer 0
