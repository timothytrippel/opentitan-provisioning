// Copyright lowRISC contributors (OpenTitan project).
// Licensed under the Apache License, Version 2.0, see LICENSE for details.
// SPDX-License-Identifier: Apache-2.0

#include "src/version/version.h"

#include <iostream>

#include "gmock/gmock.h"
#include "gtest/gtest.h"

namespace provisioning {
namespace {
using provisioning::VersionFormatted;
using testing::MatchesRegex;

TEST(VersionTest, FormattedVersionOk) {
  std::cout << "version = " << VersionFormatted() << std::endl;

  // Sanity check to make sure the Formatted version string contains non-empty
  // param values.
  EXPECT_THAT(VersionFormatted(),
              MatchesRegex("Version:\\s.+?-.+?\\sHost:\\s.+?\\sUser:\\s.+?"
                           "\\sTimestamp:\\s\\S+?\\s"));
}

}  // namespace
}  // namespace provisioning
