// Copyright lowRISC contributors (OpenTitan project).
// Licensed under the Apache License, Version 2.0, see LICENSE for details.
// SPDX-License-Identifier: Apache-2.0
#include "src/transport/service_credentials.h"

#include <grpcpp/grpcpp.h>

namespace provisioning {
namespace transport {

using grpc::Status;

constexpr char kCredentialsKey[] = "x-opentitan-auth-token";

ServiceCredentials::ServiceCredentials(
    const std::vector<std::string>& sku_tokens)
    : sku_tokens_(sku_tokens) {}

Status ServiceCredentials::GetMetadata(
    grpc::string_ref service_url, grpc::string_ref method_name,
    const grpc::AuthContext& channel_auth_context,
    std::multimap<grpc::string, grpc::string>* metadata) {
  for (const std::string& sku_token : sku_tokens_) {
    metadata->emplace(kCredentialsKey, sku_token);
  }
  return Status::OK;
}

}  // namespace transport
}  // namespace provisioning
