#!/usr/bin/env bash
# Copyright lowRISC contributors (OpenTitan project).
# Licensed under the Apache License, Version 2.0, see LICENSE for details.
# SPDX-License-Identifier: Apache-2.0

GOFMT=@@GOFMT@@
MODE=@@MODE@@

gofmt=$(readlink "$GOFMT")

if ! cd "$BUILD_WORKSPACE_DIRECTORY"; then
    echo "Unable to change to workspace (BUILD_WORKSPACE_DIRECTORY: ${BUILD_WORKSPACE_DIRECTORY})"
    exit 1
fi

if [[ $# != 0 ]]; then
    FILES="$@"
else
    FILES=$(find . \
        -type f \
        @@EXCLUDE_PATTERNS@@ \
        \( @@INCLUDE_PATTERNS@@ \) \
        -print)
fi

case "$MODE" in
    diff)
        echo "$FILES" | xargs ${gofmt} -s -d
        exit $?
        ;;
    fix)
        echo "$FILES" | xargs ${gofmt} -s -w
        ;;
    *)
        echo "Unknown mode: $MODE"
        exit 2
esac
