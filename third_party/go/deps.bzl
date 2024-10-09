# Copyright lowRISC contributors (OpenTitan project).
# Licensed under the Apache License, Version 2.0, see LICENSE for details.
# SPDX-License-Identifier: Apache-2.0

load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies", "go_repository")
load("@io_bazel_rules_go//go:deps.bzl", "go_register_toolchains", "go_rules_dependencies")

def local_go_deps():
    go_packages_()
    go_rules_dependencies()
    go_register_toolchains(
        nogo = "@//:vet",
        version = "1.18.2",
    )
    gazelle_dependencies()

# go_packages_ is an automatically generated macro that lists repositories
# corresponding to //:go.mod's requirements.
#
# Use `bazel run //:update-go-repos` to update it, after changing //:go.mod.
def go_packages_():
    go_repository(
        name = "com_github_cenkalti_backoff_v4",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/cenkalti/backoff/v4",
        sum = "h1:y4OZtCnogmCPw98Zjyt5a6+QwPLGkiQsYW5oUqylYbM=",
        version = "v4.2.1",
    )
    go_repository(
        name = "com_github_elastic_elastic_transport_go_v8",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/elastic/elastic-transport-go/v8",
        sum = "h1:DJGxovyQLXGr62e9nDMPSxRyWION0Bh6d9eCFBriiHo=",
        version = "v8.3.0",
    )
    go_repository(
        name = "com_github_elastic_go_elasticsearch_v8",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/elastic/go-elasticsearch/v8",
        sum = "h1:JJ3i2DimYTsJcUoEGbg6tNB0eehTNdid9c5kTR1TGuI=",
        version = "v8.10.1",
    )
    go_repository(
        name = "com_github_golang_protobuf",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/golang/protobuf",
        sum = "h1:ROPKBNFfQgOUMifHyP+KYbvpjbdoFNs+aK7DXlji0Tw=",
        version = "v1.5.2",
    )
    go_repository(
        name = "com_github_google_go_cmp",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/google/go-cmp",
        sum = "h1:BKbKCqvP6I+rmFHt06ZmyQtvB8xAkWdhFyr0ZUNZcxQ=",
        version = "v0.5.6",
    )
    go_repository(
        name = "com_github_google_tink_go",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/google/tink/go",
        sum = "h1:t7JHqO8Ath2w2ig5vjwQYJzhGEZymedQc90lQXUBa4I=",
        version = "v1.6.1",
    )
    go_repository(
        name = "com_github_miekg_pkcs11",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/miekg/pkcs11",
        sum = "h1:iMwmD7I5225wv84WxIG/bmxz9AXjWvTWIbM/TYHvWtw=",
        version = "v1.0.3",
    )
    go_repository(
        name = "io_etcd_go_etcd",
        build_file_proto_mode = "disable_global",
        importpath = "go.etcd.io/etcd",
        sum = "h1:5hMrpf6REqTHV2LW2OclNpRtxI0k9ZplMemJsMSWju0=",
        version = "v3.3.27+incompatible",
    )
    go_repository(
        name = "io_etcd_go_etcd_api_v3",
        build_file_proto_mode = "disable_global",
        importpath = "go.etcd.io/etcd/api/v3",
        sum = "h1:v28cktvBq+7vGyJXF8G+rWJmj+1XUmMtqcLnH8hDocM=",
        version = "v3.5.1",
    )
    go_repository(
        name = "io_etcd_go_etcd_client_v3",
        build_file_proto_mode = "disable_global",
        importpath = "go.etcd.io/etcd/client/v3",
        sum = "h1:oImGuV5LGKjCqXdjkMHCyWa5OO1gYKCnC/1sgdfj1Uk=",
        version = "v3.5.1",
    )
    go_repository(
        name = "org_golang_google_grpc",
        build_file_proto_mode = "disable_global",
        importpath = "google.golang.org/grpc",
        sum = "h1:f+PlOh7QV4iIJkPrx5NQ7qaNGFQ3OTse67yaDHfju4E=",
        version = "v1.41.0",
    )
    go_repository(
        name = "org_golang_x_crypto",
        build_file_proto_mode = "disable_global",
        importpath = "golang.org/x/crypto",
        sum = "h1:syTAU9FwmvzEoIYMqcPHOcVm4H3U5u90WsvuYgwpETU=",
        version = "v0.0.0-20220307211146-efcb8507fb70",
    )
    go_repository(
        name = "org_golang_x_sync",
        build_file_proto_mode = "disable_global",
        importpath = "golang.org/x/sync",
        sum = "h1:wsuoTGHzEhffawBOhz5CYhcrV4IdKZbEyZjBMuTp12o=",
        version = "v0.1.0",
    )
    go_repository(
        name = "org_golang_x_tools",
        build_file_proto_mode = "disable_global",
        importpath = "golang.org/x/tools",
        sum = "h1:QjFRCZxdOhBJ/UNgnBZLbNV13DlbnK0quyivTnXJM20=",
        version = "v0.1.10",
    )
