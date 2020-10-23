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
build linux amd64
build netbsd amd64
build freebsd amd64
build openbsd amd64
build darwin amd64
