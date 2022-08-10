#!/bin/bash

/usr/bin/rpmbuild -ba --define "_version ${CONFIGSYNC_VERSION}" --define "_date ${BUILD_DATE}" configsync.spec