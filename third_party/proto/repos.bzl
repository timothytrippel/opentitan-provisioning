# Copyright lowRISC contributors (OpenTitan project).
# Licensed under the Apache License, Version 2.0, see LICENSE for details.
# SPDX-License-Identifier: Apache-2.0

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def proto_repos():
    http_archive(
        name = "build_stack_rules_proto",
        sha256 = "ee7a11d66e7bbc5b0f7a35ca3e960cb9a5f8a314b22252e19912dfbc6e22782d",
        strip_prefix = "rules_proto-3.1.0",
        urls = ["https://github.com/stackb/rules_proto/archive/v3.1.0.tar.gz"],
    )
