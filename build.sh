#!/bin/bash
set -e

CONFIGSYNC_VERSION=${1:?Version required}
BUILD_DATE=$(date -u -R)

function build {
    GOOS=$1
    GOARCH=$2
    NAME="configsync-${CONFIGSYNC_VERSION}_${GOOS}-${GOARCH}.tar.gz"

    rm -f configsync
    CGO_ENABLED=0 GOAMD64=v2 go build -buildmode=exe -trimpath -ldflags="-s -w -X 'main.Version=${CONFIGSYNC_VERSION}' -X 'main.BuildDate=${BUILD_DATE}'" -v -o configsync
    tar -czf ${NAME} configsync
    rm -f configsync
    mv ${NAME} ../../artifacts/
}

rm -rf artifacts
mkdir -p artifacts

cd cmd/configsync
for ARCH in 'amd64' 'arm64'; do
    for OS in 'linux' 'freebsd' 'openbsd' 'netbsd' 'darwin'; do
        build ${OS} ${ARCH}
    done
done

cd ../../package/rpm
./build.sh ${CONFIGSYNC_VERSION}
cd ../deb
./build.sh ${CONFIGSYNC_VERSION}
