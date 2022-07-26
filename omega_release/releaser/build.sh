cd `dirname $0`
cd ..
rm -rf binary
set -e 
mkdir binary

PHOENIX_BUILDER_DIR=".."
TIME_STAMP=$(date '+%m%d%H%M')

make -C ${PHOENIX_BUILDER_DIR} clean 
make -C ${PHOENIX_BUILDER_DIR} build/phoenixbuilder build/phoenixbuilder-windows-executable-x86_64.exe build/phoenixbuilder-android-executable-arm64 build/phoenixbuilder-macos-x86_64 -j4
cp ${PHOENIX_BUILDER_DIR}/build/phoenixbuilder ./binary/fastbuilder-linux
cp ${PHOENIX_BUILDER_DIR}/build/phoenixbuilder-windows-executable-x86_64.exe ./binary/fastbuilder-windows.exe
cp ${PHOENIX_BUILDER_DIR}/build/phoenixbuilder-macos-x86_64 ./binary/fastbuilder-macos
cp ${PHOENIX_BUILDER_DIR}/build/phoenixbuilder-android-executable-arm64 ./binary/fastbuilder-android

function get_hash(){
    fileName=$1
    outstr="$(md5sum $fileName)"
    hashStr="$(echo $outstr | cut -d' ' -f1)"
    echo "$hashStr"
}

echo $(get_hash ./binary/fastbuilder-linux) > ./binary/fastbuilder-linux.hash
echo $(get_hash ./binary/fastbuilder-windows.exe) > ./binary/fastbuilder-windows.hash
echo $(get_hash ./binary/fastbuilder-macos) > ./binary/fastbuilder-macos.hash
echo $(get_hash ./binary/fastbuilder-android) > ./binary/fastbuilder-android.hash
cat ./binary/*.hash > ./binary/all.hash

go run ./compressor/main.go -in "\
        ./binary/fastbuilder-linux,\
        ./binary/fastbuilder-windows.exe,\
        ./binary/fastbuilder-macos,\
        ./binary/fastbuilder-android\
    " -out "\
        ./binary/fastbuilder-linux.brotli,\
        ./binary/fastbuilder-windows.exe.brotli,\
        ./binary/fastbuilder-macos.brotli,\
        ./binary/fastbuilder-android.brotli\
    "

# cp ./releaser/更新日志.txt ./binary
cp ./releaser/install.sh ./binary
cp ./releaser/dockerfile ./binary
cp ./releaser/docker_bootstrap.sh ./binary
echo "$TIME_STAMP" >> ./binary/TIME_STAMP

make -C ./launcher clean
make -C ./launcher all -j6
cp ./launcher/build/* ./binary
echo $(get_hash ./binary/launcher-linux-mcsm) > ./binary/launcher-linux-mcsm.hash
echo $(get_hash ./binary/launcher-linux) > ./binary/launcher-linux.hash
echo $(get_hash ./binary/launcher-macos) > ./binary/launcher-macos.hash
echo $(get_hash ./binary/launcher-android) > ./binary/launcher-android.hash

rsync -avP --delete ./binary/* FBOmega:/var/www/omega-storage/binary/