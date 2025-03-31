#!/bin/bash
set -e

ROOT_PATH=$(realpath ../../)
CONFIGSYNC_VERSION=${1:?Version required}
BUILD_DATE=$(date -u -R)
DOCKER_CMD=${DOCKER:-"podman"}

rm -rf build/
mkdir -p build/

cd ${ROOT_PATH}
tar -czf configsync-${CONFIGSYNC_VERSION}.tar.gz *.go git/*.go cmd/configsync/*.go go.mod go.sum
mv configsync-${CONFIGSYNC_VERSION}.tar.gz package/rpm/build
cd package/rpm
cp Dockerfile configsync.spec entrypoint.sh build/

cd build/
mkdir configsync-${CONFIGSYNC_VERSION}/
mv configsync-${CONFIGSYNC_VERSION}.tar.gz configsync-${CONFIGSYNC_VERSION}/
cd configsync-${CONFIGSYNC_VERSION}/
tar -xzf configsync-${CONFIGSYNC_VERSION}.tar.gz
rm configsync-${CONFIGSYNC_VERSION}.tar.gz
cd ../
tar -czf configsync-${CONFIGSYNC_VERSION}.tar.gz configsync-${CONFIGSYNC_VERSION}/
rm -r configsync-${CONFIGSYNC_VERSION}/

GOLANG_ARCH="amd64"
PODMAN_ARCH="amd64"
if [[ $(uname -m) == 'aarch64' ]]; then
    GOLANG_ARCH="arm64"
    PODMAN_ARCH="arm64"
fi
GOLANG_VERSION=$(curl -sS "https://go.dev/dl/?mode=json" | jq -r '.[0].version' | sed 's/go//')

podman build --platform=linux/${PODMAN_ARCH} --build-arg GOLANG_ARCH=${GOLANG_ARCH} --build-arg GOLANG_VERSION=${GOLANG_VERSION} -t localhost/configsync_build_rpm:${CONFIGSYNC_VERSION} .
ID=$(podman image inspect --format '{{ .Id }}' localhost/configsync_build_rpm:${CONFIGSYNC_VERSION})
rm -rf rpms
mkdir -p rpms
podman run --platform=linux/${PODMAN_ARCH} --user root -v $(readlink -f rpms):/root/rpmbuild/RPMS:Z -e CONFIGSYNC_VERSION=${CONFIGSYNC_VERSION} -e BUILD_DATE="${BUILD_DATE}" -it ${ID}
cp -v rpms/*/*.rpm .
mv *.rpm ../../../artifacts
cd ../
rm -rf build/
