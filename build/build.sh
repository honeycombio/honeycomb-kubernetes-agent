#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail
set -x

if [ -z "${VERSION}" ]; then
    echo "VERSION must be set"
    exit 1
fi

export CGO_ENABLED=0

export KO_DOCKER_REPO=${KO_DOCKER_REPO:-ko.local}
# See https://github.com/google/ko for options
# --base-import-paths avoids us having a weird -$(md5sum) at the end of the
# image name
ko publish --tags head,${VERSION} --base-import-paths --platform=linux/amd64,linux/arm64 .
