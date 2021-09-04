#!/bin/bash

# Copyright 2016 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

export VERSION=${VERSION:-$(cat "$(dirname "$0")"/../version.txt)}
PLATFORM="${PLATFORM:-linux/amd64,linux/arm64}"

unset GOOS
unset GOARCH
export KO_DOCKER_REPO=${KO_DOCKER_REPO:-ko.local}
# shellcheck disable=SC2086
ko publish \
  --tags "head,${VERSION}" \
  --base-import-paths \
  --platform "${PLATFORM}" \
  ${PUBLISH_ARGS-} \
  .
