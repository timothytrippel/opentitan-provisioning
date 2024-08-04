// Copyright lowRISC contributors (OpenTitan project).
// Licensed under the Apache License, Version 2.0, see LICENSE for details.
// SPDX-License-Identifier: Apache-2.0
#include "src/transport/service_credentials.h"

#include <gmock/gmock.h>
#include <grpcpp/grpcpp.h>
#include <grpcpp/security/auth_context.h>
#include <grpcpp/security/credentials.h>
#include <grpcpp/support/config.h>
#include <grpcpp/support/string_ref.h>
#include <gtest/gtest.h>

#include <map>
#include <string>
#include <vector>

namespace provisioning {
namespace transport {
namespace {

using testing::ContainerEq;
using testing::IsTrue;
using testing::StrictMock;

class MockAuthContext : public grpc::AuthContext {
  MOCK_METHOD(bool, IsPeerAuthenticated, (), (const, override));
  MOCK_METHOD(std::vector<grpc::string_ref>, GetPeerIdentity, (),
              (const, override));
  MOCK_METHOD(std::string, GetPeerIdentityPropertyName, (), (const, override));
  MOCK_METHOD(std::vector<grpc::string_ref>, FindPropertyValues,
              (const std::string&), (const, override));
  MOCK_METHOD(grpc::AuthPropertyIterator, begin, (), (const, override));
  MOCK_METHOD(grpc::AuthPropertyIterator, end, (), (const, override));
  MOCK_METHOD(void, AddProperty, (const std::string&, const grpc::string_ref&),
              (override));
  MOCK_METHOD(bool, SetPeerIdentityPropertyName, (const std::string&),
              (override));
};

TEST(ServiceCredentialsTest, Type) {
  std::vector<std::string> sku_tokens;
  ServiceCredentials credentials(sku_tokens);
  EXPECT_EQ("OpenTitanAuthToken", credentials.GetType());
}

TEST(ServiceCredentialsTest, DebugString) {
  std::vector<std::string> sku_tokens;
  ServiceCredentials credentials(sku_tokens);
  EXPECT_EQ("OpenTitanAuthToken", credentials.DebugString());
}

TEST(ServiceCredentialsTest, GetMetadataOk) {
  std::vector<std::string> sku_tokens = {"TokenSkuA", "TokenSkuB"};
  ServiceCredentials credentials(sku_tokens);

  const std::string credentials_key = "x-opentitan-auth-token";

  std::multimap<std::string, std::string> expected;
  for (const std::string& token : sku_tokens) {
    expected.emplace(credentials_key, token);
  }

  std::multimap<std::string, std::string> metadata;
  EXPECT_THAT(credentials
                  .GetMetadata(/*service_url=*/"", /*method_name=*/"",
                               StrictMock<MockAuthContext>(), &metadata)
                  .ok(),
              IsTrue());
  EXPECT_THAT(metadata, ContainerEq(expected));
}

}  // namespace
}  // namespace transport
}  // namespace provisioning
