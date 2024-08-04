# Copyright lowRISC contributors (OpenTitan project).
# Licensed under the Apache License, Version 2.0, see LICENSE for details.
# SPDX-License-Identifier: Apache-2.0

load("@bazel_tools//tools/build_defs/repo:utils.bzl", "maybe")
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def lowrisc_repos():
    maybe(
        http_archive,
        name = "lowrisc_misc_linters",
        sha256 = "0b3b7b8f5ceacda50ca74a5b7dfeddcbd5ddfa8ffd1a482878aee2fc02989794",
        strip_prefix = "misc-linters-20220921_01",
        url = "https://github.com/lowRISC/misc-linters/archive/refs/tags/20220921_01.tar.gz",
    )
    maybe(
        http_archive,
        name = "lowrisc_bazel_release",
        sha256 = "c7b0cbdec0a1081a0b0a52eb1ebd942e7eaa218408008661fdb6e8ec3b441a4a",
        strip_prefix = "bazel-release-0.0.3",
        url = "https://github.com/lowRISC/bazel-release/archive/refs/tags/v0.0.3.tar.gz",
    )
