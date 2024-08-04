# Copyright lowRISC contributors (OpenTitan project).
# Licensed under the Apache License, Version 2.0, see LICENSE for details.
# SPDX-License-Identifier: Apache-2.0

"""Linting rules for OT Provisioning."""

load("@bazel_skylib//lib:shell.bzl", "shell")

def _gofmt_impl(ctx):
    out_file = ctx.actions.declare_file(ctx.label.name + ".bash")
    exclude_patterns = ["\\! -path {}".format(shell.quote(p)) for p in ctx.attr.exclude_patterns]
    include_patterns = ["-name {}".format(shell.quote(p)) for p in ctx.attr.patterns]
    substitutions = {
        "@@EXCLUDE_PATTERNS@@": " ".join(exclude_patterns),
        "@@INCLUDE_PATTERNS@@": " -o ".join(include_patterns),
        "@@GOFMT@@": shell.quote(ctx.executable.gofmt.short_path),
        "@@DIFF_COMMAND@@": shell.quote(ctx.attr.diff_command),
        "@@MODE@@": shell.quote(ctx.attr.mode),
    }
    ctx.actions.expand_template(
        template = ctx.file._runner,
        output = out_file,
        substitutions = substitutions,
        is_executable = True,
    )

    return DefaultInfo(
        files = depset([out_file]),
        runfiles = ctx.runfiles(files = [ctx.executable.gofmt]),
        executable = out_file,
    )

gofmt = rule(
    implementation = _gofmt_impl,
    attrs = {
        "patterns": attr.string_list(
            default = ["*.go"],
            doc = "Filename patterns for format checking",
        ),
        "exclude_patterns": attr.string_list(
            doc = "Filename patterns to exlucde from format checking",
        ),
        "mode": attr.string(
            default = "diff",
            values = ["diff", "fix"],
            doc = "Execution mode: display diffs or fix formatting",
        ),
        "diff_command": attr.string(
            default = "diff -u",
            doc = "Command to execute to display diffs",
        ),
        "gofmt": attr.label(
            default = "@go_sdk//:bin/gofmt",
            allow_single_file = True,
            cfg = "host",
            executable = True,
            doc = "The gofmt executable",
        ),
        "_runner": attr.label(
            default = "//rules/scripts:gofmt.template.sh",
            allow_single_file = True,
        ),
    },
    executable = True,
)

def _clang_format_impl(ctx):
    out_file = ctx.actions.declare_file(ctx.label.name + ".bash")
    exclude_patterns = ["\\! -path {}".format(shell.quote(p)) for p in ctx.attr.exclude_patterns]
    include_patterns = ["-name {}".format(shell.quote(p)) for p in ctx.attr.patterns]
    substitutions = {
        "@@EXCLUDE_PATTERNS@@": " ".join(exclude_patterns),
        "@@INCLUDE_PATTERNS@@": " -o ".join(include_patterns),
        "@@CLANG_FORMAT@@": shell.quote(ctx.attr.clang_format_command),
        "@@DIFF_COMMAND@@": shell.quote(ctx.attr.diff_command),
        "@@MODE@@": shell.quote(ctx.attr.mode),
    }
    ctx.actions.expand_template(
        template = ctx.file._runner,
        output = out_file,
        substitutions = substitutions,
        is_executable = True,
    )

    return DefaultInfo(
        files = depset([out_file]),
        executable = out_file,
    )

clang_format_check = rule(
    implementation = _clang_format_impl,
    attrs = {
        "patterns": attr.string_list(
            default = ["*.c", "*.h", "*.cc", "*.cpp"],
            doc = "Filename patterns for format checking",
        ),
        "exclude_patterns": attr.string_list(
            doc = "Filename patterns to exlucde from format checking",
        ),
        "mode": attr.string(
            default = "diff",
            values = ["diff", "fix"],
            doc = "Execution mode: display diffs or fix formatting",
        ),
        "diff_command": attr.string(
            default = "diff -u",
            doc = "Command to execute to display diffs",
        ),
        "clang_format_command": attr.string(
            default = "clang-format",
            doc = "The clang-format executable",
        ),
        "_runner": attr.label(
            default = "//rules/scripts:clang_format.template.sh",
            allow_single_file = True,
        ),
    },
    executable = True,
)
