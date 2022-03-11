#!/bin/bash
set -e
# trigger by:
# bash docker/download_deps.sh
# docker build -t cma2401pt/phoenixbuilder docker
# docker run --name="builder" --rm --volume $PWD:/work --volume $PWD/docker/cache:/root/go -e HOST_UID=`id -u $USER` -e HOST_GID=`id -g $USER` -e HOST_USER=$USER cma2401pt/phoenixbuilder:latest  /bin/bash /work/docker/start.sh

echo 'env:'
echo HOST_USER=$HOST_USER
echo HOST_GID=$HOST_GID
echo HOST_UID=$HOST_UID

cd /work

source /etc/profile

echo ""
echo "Pre-Build & configure go-raknet"
make current
make clean
chmod 0644 ~/go/pkg/mod/github.com/sandertv/go-raknet@v1.9.1/conn.go
sed "s/urrentProtocol byte = 10/urrentProtocol byte = 8/g" ~/go/pkg/mod/github.com/sandertv/go-raknet@v1.9.1/conn.go>~/conn.go
cp -f ~/conn.go ~/go/pkg/mod/github.com/sandertv/go-raknet@v1.9.1/conn.go

echo ""
echo "Build"
export THEOS=/theos
make ios-executable

#echo ""
#echo "Package for specific platforms"
#export THEOS=/theos
#make package
#
#echo ""
#echo "Build GUI applications"
#export THEOS=/theos
#cd fyne-gui
#cd android_wrapper
#go build
#cd ..
#make
#node simpleappend.js
#ls -lh build/>release.txt
#mv release.txt build/release-list.txt
#cp version build/version.txt
#mv build ../build/gui
#cd ..
#
#echo ""
#echo "Pack binaries"
#mv release/*.deb build/
#rm build/phoenixbuilder-macos-arm64
#rm build/phoenixbuilder-macos-x86_64
#ls -lh build/>release.txt
#mv release.txt build/release-list.txt
#cp version build/version.txt
#tar -czf fb-linux-binaries-v2.tar.gz build/*