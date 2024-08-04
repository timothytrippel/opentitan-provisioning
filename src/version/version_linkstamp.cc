// Copyright lowRISC contributors (OpenTitan project).
// Licensed under the Apache License, Version 2.0, see LICENSE for details.
// SPDX-License-Identifier: Apache-2.0

// This file is compiled and linked at the same time to introduce timestamp
// information into the binary. Because this is done using the `linkstamp` Bazel
// option, it is not possible to add dependencies to this file.

#ifndef BUILD_HOST
#define BUILD_HOST "unkown"
#endif

#ifndef BUILD_USER
#define BUILD_USER "unknown"
#endif

#ifndef BUILD_TIMESTAMP
#define BUILD_TIMESTAMP "0"
#endif

#ifndef BUILD_SCM_REVISION
#define BUILD_SCM_REVISION "unknown"
#endif

#ifndef BUILD_SCM_STATUS
#define BUILD_SCM_STATUS "unknown"
#endif

#define TO_STRING2(x) #x
#define TO_STRING(x) TO_STRING2(x)

extern "C" const char kBuildHost[];
const char kBuildHost[] = BUILD_HOST;

extern "C" const char kBuildUser[];
const char kBuildUser[] = BUILD_USER;

extern "C" const char kBuildTimestamp[];
const char kBuildTimestamp[] = TO_STRING(BUILD_TIMESTAMP);

extern "C" const char kBuildRevision[];
const char kBuildRevision[] = BUILD_SCM_REVISION;

extern "C" const char kBuildStatus[];
const char kBuildStatus[] = BUILD_SCM_STATUS;
