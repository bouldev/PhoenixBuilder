#!/bin/bash
set -e
# trigger by:
# bash docker/download_deps.sh
# docker build -t cma2401pt/phoenixbuilder docker
# docker run --name="builder" --rm --volume $PWD:/work --volume $PWD/docker/cache:/root/go -e HOST_UID=`id -u $USER` -e HOST_GID=`id -g $USER` -e HOST_USER=$USER cma2401pt/phoenixbuilder:latest  /bin/bash /work/docker/build.sh

echo 'env:'
echo HOST_USER=$HOST_USER
echo HOST_GID=$HOST_GID
echo HOST_UID=$HOST_UID

cd /work

source /etc/profile

make current
make clean
chmod 0644 ~/go/pkg/mod/github.com/sandertv/go-raknet@v1.9.1/conn.go
sed "s/urrentProtocol byte = 10/urrentProtocol byte = 8/g" ~/go/pkg/mod/github.com/sandertv/go-raknet@v1.9.1/conn.go>~/conn.go
cp -f ~/conn.go ~/go/pkg/mod/github.com/sandertv/go-raknet@v1.9.1/conn.go