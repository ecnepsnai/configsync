#!/bin/bash

function build {
    GOOS=$1
    GOARCH=$2
    NAME="configsync_${GOOS}_${GOARCH}.tar.gz"

    rm -f configsync
    go build -ldflags="-s -w"
    tar -czf ${NAME} configsync
    rm -f configsync
    mv ${NAME} artifacts/
}

rm -rf artifacts
mkdir -p artifacts

for ARCH in 'amd64' 'arm64'; do
    for OS in 'linux' 'freebsd' 'openbsd' 'netbsd' 'darwin'; do
        build ${OS} ${ARCH}
    done
done
