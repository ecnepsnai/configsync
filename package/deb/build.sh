#!/bin/bash
set -e
set -x

ROOT_PATH=$(realpath ../../)
CONFIGSYNC_VERSION=${1:?Version required}
BUILD_DATE=$(date -R)
DOCKER_CMD=${DOCKER:-"podman"}

ARCH="amd64"
if [[ $(uname -m) == 'aarch64' ]]; then
    ARCH="arm64"
fi

rm -rf build
mkdir -p build/DEBIAN
mkdir -p build/usr/sbin

cp ../../artifacts/configsync-${CONFIGSYNC_VERSION}_linux-${ARCH}.tar.gz .
tar -xzf configsync-${CONFIGSYNC_VERSION}_linux-${ARCH}.tar.gz
rm configsync-${CONFIGSYNC_VERSION}_linux-${ARCH}.tar.gz
mv configsync build/usr/sbin/configsync

cp configsync.control.spec build/DEBIAN/control
perl -pi -e "s,%%VERSION%%,${CONFIGSYNC_VERSION},g" build/DEBIAN/control
perl -pi -e "s,%%ARCH%%,${ARCH},g" build/DEBIAN/control

podman build -t configsync_build_deb:${CONFIGSYNC_VERSION} .
podman run --user root -v $(readlink -f build):/configsync:Z -e CONFIGSYNC_VERSION=${CONFIGSYNC_VERSION} -e ARCH=${ARCH} -it configsync_build_deb:${CONFIGSYNC_VERSION}

cp build/*.deb ../../artifacts
rm -rf build
