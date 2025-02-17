#!/bin/bash

/usr/bin/rpmbuild --target $(uname -m) -ba --define "_version ${CONFIGSYNC_VERSION}" --define "_date ${BUILD_DATE}" configsync.spec