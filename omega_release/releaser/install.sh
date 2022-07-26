#!/bin/bash 
set -e 
PLANTFORM="Unknown"
STOARGE_REPO="https://omega.fastbuilder.pro/binary"
# /bin/bash -c "$(curl -fsSL https://omega.fastbuilder.pro/binary/install.sh)"


skip_updatec_check=0
working_dir=$PWD
executable="$PWD/omega-launcher"

function EXIT_FAILURE(){
    exit -1
}

function get_hash(){
    file_name=$1
    if [[ $PLANTFORM == "Macos_x86_64" ]]; then
        out_hash="$(md5 -q $file_name)"
    else
        out_str="$(md5sum $file_name)"
        out_hash="$(echo $out_str | cut -d' ' -f1)"
    fi 
    echo $out_hash
}

function yellow_line(){
    printf "\033[33m$1\033[0m\n"
}

function red_line(){
    printf "\033[31m$1\033[0m\n"
}

function green_line(){
    printf "\033[32m$1\033[0m\n"
}

function download_exec(){
    case ${PLANTFORM} in
        "Linux_x86_64")
        url="$STOARGE_REPO/launcher-linux"
        hash_url="$STOARGE_REPO/launcher-linux.hash"
        ;;
        "Andorid_armv8")
        url="$STOARGE_REPO/launcher-android"
        hash_url="$STOARGE_REPO/launcher-android.hash"
        ;;
        "Macos_x86_64")
        url="$STOARGE_REPO/launcher-macos"
        hash_url="$STOARGE_REPO/launcher-macos.hash"
        ;;
        *)
        echo "不支持的平台${PLANTFORM}"
        EXIT_FAILURE
        ;;
    esac
    current_url=""
    target_hash=$(curl "$hash_url")
   
    if [ -e $executable ]; then 
        current_hash=$(get_hash $executable)
    fi 
    echo $target_hash $current_hash
    if [[ $target_hash == $current_hash ]]; then 
        echo -e ""
    else
        yellow_line "开始下载启动器...请耐心等待"
        curl $url -o $executable
    fi 
    chmod 777 $executable
}

if [[ $(uname) == "Darwin" ]]; then
    PLANTFORM="Macos_x86_64"
elif [[ $(uname -o) == "GNU/Linux" ]] || [[ $(uname -o) == "GNU/Linux" ]]; then 
    PLANTFORM="Linux_x86_64"
    if [[ $(uname -m) != "x86_64" ]]; then
        echo "不支持非64位的Linux系统"
        EXIT_FAILURE
    fi 
elif [[ $(uname -o) == "Android" ]]; then 
    PLANTFORM="Andorid_armv8"
    if [[ $(uname -m) == "armv7" ]]; then
        echo "不支持armv7的Andorid系统"
        EXIT_FAILURE
    fi 
    echo "检测文件权限中..."
    if [ ! -x "/sdcard/Download" ]; then 
        echo "请给予 termux 文件权限 ~"
        sleep 2
        termux-setup-storage
    fi 
    if [ -x "/sdcard/Download" ]; then 
        echo -e ""
        # green_line "太好了，omega将被保存到downloads文件夹下，你可以从任何文件管理器中找到它了"
        # working_dir="/sdcard/Download"
        # executable="/sdcard/Download/fastbuilder"
    else 
        red_line "不行啊，没给权限"
        EXIT_FAILURE
    fi 
else
    echo "不支持该系统，你的系统是"
    uname -a 
fi 
download_exec
cd $working_dir
echo $PWD 
echo "启动中..."
$executable