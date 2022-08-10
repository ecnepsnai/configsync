#!/bin/bash
set -e
set -x

pwd
dpkg-deb --build --root-owner-group configsync
mv ./configsync.deb ./configsync/configsync-${CONFIGSYNC_VERSION}.${ARCH}.deb
