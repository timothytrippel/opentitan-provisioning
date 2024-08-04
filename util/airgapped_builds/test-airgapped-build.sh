#!/bin/bash
# Copyright lowRISC contributors (OpenTitan project).
# Licensed under the Apache License, Version 2.0, see LICENSE for details.
# SPDX-License-Identifier: Apache-2.0

set -ex

: "${REPO_TOP:=$(git rev-parse --show-toplevel)}"

# Prefetch bazel airgapped dependencies.
. ${REPO_TOP}/util/airgapped_builds/prep-bazel-airgapped-build.sh -f

# Remove the airgapped network namespace.
remove_airgapped_netns() {
  sudo ip netns delete airgapped
}
trap remove_airgapped_netns EXIT

# Set up a network namespace named "airgapped" with access to loopback.
sudo ip netns add airgapped
sudo ip netns exec airgapped ip addr add 127.0.0.1/8 dev lo
sudo ip netns exec airgapped ip link set dev lo up

# Enter the network namespace and perform several builds.
sudo ip netns exec airgapped sudo -u "$USER" bash -c \
  "TARGET_PATTERN_FILE=\$(mktemp)
  echo //... > \"\${TARGET_PATTERN_FILE}\"
  bazel-airgapped/bazel cquery \
    --distdir=$(pwd)/bazel-airgapped/bazel-distdir \
    --repository_cache=$(pwd)/bazel-airgapped/bazel-cache \
    --noinclude_aspects \
    --output=starlark \
    --starlark:expr='\"-{}\".format(target.label)' \
    --define DISABLE_VERILATOR_BUILD=true \
    -- \"rdeps(//..., kind(bitstream_splice, //...))\" \
    >> \"\${TARGET_PATTERN_FILE}\"
  echo Building target pattern:
  cat \"\${TARGET_PATTERN_FILE}\"
  bazel-airgapped/bazel build \
    --distdir=$(pwd)/bazel-airgapped/bazel-distdir \
    --repository_cache=$(pwd)/bazel-airgapped/bazel-cache \
    --define DISABLE_VERILATOR_BUILD=true \
    --target_pattern_file=\"\${TARGET_PATTERN_FILE}\""
exit 0
