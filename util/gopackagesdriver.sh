#!/bin/bash
# Copyright lowRISC contributors (OpenTitan project).
# Licensed under the Apache License, Version 2.0, see LICENSE for details.
# SPDX-License-Identifier: Apache-2.0
set -e
exec bazelisk run -- @io_bazel_rules_go//go/tools/gopackagesdriver "${@}"