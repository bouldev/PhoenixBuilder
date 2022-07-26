set -e 
sleep 1
mkdir -p omega_docker
HOST_UID=$(id -u)
HOST_GID=$(id -g)
HOST_USER=$(whoami)
OMEGA_SOURECE="https://omega.fastbuilder.pro/binary"
curl ${OMEGA_SOURECE}/dockerfile -o omega_docker/dockerfile
TIME_STAMP="$(curl ${OMEGA_SOURECE}/TIME_STAMP)"
docker build \
    --build-arg HOST_UID=$HOST_UID \
    --build-arg HOST_GID=$HOST_GID \
    --build-arg HOST_USER=$HOST_USER \
    --build-arg TIME_STAMP=$TIME_STAMP \
    --build-arg OMEGA_SOURCE=$OMEGA_SOURECE \
    -t omega/omega:current omega_docker
#docker image prune -f
rm -rf omega_docker
echo "镜像创建/更新成功,镜像名: omega/omega:current "