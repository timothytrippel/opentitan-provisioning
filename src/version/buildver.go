// Copyright lowRISC contributors (OpenTitan project).
// Licensed under the Apache License, Version 2.0, see LICENSE for details.
// SPDX-License-Identifier: Apache-2.0

// buildver package provides access to build version variables and utilities
// to generate formatted version strings.
package buildver

import (
	"fmt"
)

var (
	// The following variables are set by Bazel via x_defs parameters. Any
	// variable name changes need to be replicated in the Bazel build target.

	// BuildHost contains the build hostname injected by Bazel.
	BuildHost = "unkown"

	// BuildUser contains the build user injected by Bazel.
	BuildUser = "unkown"

	// BuildTimestamp contains the build timestamp injected by Bazel.
	BuildTimestamp = "0"

	// BuildSCMRevision contains the repository release tag or commit hash
	// injected by Bazel.
	BuildSCMRevision = "unkown"

	// BuildSCMStatus contains the status of the repository injected by Bazel.
	BuildSCMStatus = "unkown"
)

// FormattedStr returns a formatted string version which can be used to
// reference the target release.
func FormattedStr() string {
	return fmt.Sprintf("Version: %s-%s Host: %q User: %q Timestamp: %s", BuildSCMRevision, BuildSCMStatus, BuildHost, BuildUser, BuildTimestamp)
}
