#!/bin/bash
set -e

function build {
    GOOS=$1
    GOARCH=$2
    NAME="configsync_${GOOS}_${GOARCH}.tar.gz"

    rm -f configsync
    CGO_ENABLED=0 go build -ldflags="-s -w"
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
