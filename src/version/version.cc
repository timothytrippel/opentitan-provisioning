// Copyright lowRISC contributors (OpenTitan project).
// Licensed under the Apache License, Version 2.0, see LICENSE for details.
// SPDX-License-Identifier: Apache-2.0
#include "src/version/version.h"

#include <sstream>

#include "absl/strings/str_cat.h"

extern "C" const char kBuildHost[] __attribute__((weak));
extern "C" const char kBuildUser[] __attribute__((weak));
extern "C" const char kBuildTimestamp[] __attribute__((weak));
extern "C" const char kBuildRevision[] __attribute__((weak));
extern "C" const char kBuildStatus[] __attribute__((weak));

namespace provisioning {

const char* BuildHost() {
  if (&kBuildHost == nullptr) return "not-set";
  return kBuildHost;
}

const char* BuildUser() {
  if (&kBuildUser == nullptr) return "not-set";
  return kBuildUser;
}

const char* BuildTimestamp() {
  if (&kBuildTimestamp == nullptr) return "not-set";
  return kBuildTimestamp;
}

const char* BuildRevision() {
  if (&kBuildRevision == nullptr) return "not-set";
  return kBuildRevision;
}

const char* BuildStatus() {
  if (&kBuildStatus == nullptr) return "not-set";
  return kBuildStatus;
}

std::string VersionFormatted() {
  return absl::StrCat("Version: ", BuildRevision(), "-", BuildStatus(),
                      " Host: ", BuildHost(), " User: ", BuildUser(),
                      " Timestamp: ", BuildTimestamp(), "\n");
}
}  // namespace provisioning
