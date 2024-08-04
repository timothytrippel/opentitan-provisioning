[//]: # (Copyright lowRISC contributors \(OpenTitan project\).)
[//]: # (Licensed under the Apache License, Version 2.0, see LICENSE for details.)
[//]: # (SPDX-License-Identifier: Apache-2.0)

# Airgapped Bazel Builds

The Bazel WORKSPACE for this repository includes several external dependencies
that are fetched over the network using the `http_archive` repository rule. By
default, this provides an obstacle to building artifacts on an airgapped
machine.

However, Bazel provides a mechanism to still allow for airgapped builds.
Specifically, you can:
1. prepare a directory containing pre-downloaded external dependencies on a
machine with network access,
1. manually move this directory to your airgapped machine, and
1. inform Bazel of its path at build time.

# Pre-downloading Bazel Depedencies

To automatically prepare the directory containing pre-downloaded Bazel
dependencies, a shell script is provided:
`util/airgapped_builds/prep-bazel-airgapped-build.sh`. It may be invoked either:
1. without flags, if it is the first time calling the script, or
1. with the `-f` flag, if you would like to overwrite an existing directory.

Invoking the script above generates a directory called `bazel-airgapped`, that
contains:
1. a `bazel` executable compiled for the target platform,
1. a Bazel [`distdir`](https://bazel.build/run/build#distribution-directory)
subdirectory containing a set of _implicit_ workspace dependencies, and
1. a Bazel [`repository cache`](https://bazel.build/run/build#repository-cache)
subdirectory containing a set of _explicit_ workspace dependencies.

This `bazel-airgapped` directory can then be moved to your airgapped machine,
and Bazel builds can be invoked as described below in [Performing Airgapped
Builds](#Performing-Airgapped-Builds).

# Performing Airgapped Builds
Once the `bazel-airgapped` directory has been moved to your airgapped machine,
Bazel builds can be performed by invoking `bazel` with the following flags:
```sh
export BAZEL_AIRGAPPED_DIR="path/to/bazel-airgapped"

${BAZEL_AIRGAPPED_DIR}/bazel build \
  --distdir=${BAZEL_AIRGAPPED_DIR}/bazel-distdir \
  --repository_cache=${BAZEL_AIRGAPPED_DIR}/bazel-cache <label>
```
where `<label>` is the target label you want to build.

For example, to build all targets, you would use:
```sh
export BAZEL_AIRGAPPED_DIR="path/to/bazel-airgapped"

${BAZEL_AIRGAPPED_DIR}/bazel build \
  --distdir=${BAZEL_AIRGAPPED_DIR}/bazel-distdir \
  --repository_cache=${BAZEL_AIRGAPPED_DIR}/bazel-cache //...
```

# Testing Airgapped Builds (Linux only)
To test airgapped builds without having to prepare directory of pre-downloaded
dependencies and manually move it to an airgapped machine, a test script is also
provided to simulate an airgapped build environment on a network-connected host
using Linux network namespaces.

This test script can be invoked with:
```sh
./util/airgapped_builds/test-airgapped-build.sh`
```

The script performs the following actions (some of which require `sudo` so be
prepared to entery your password if required):
1. calls `util/airgapped_builds/prep-bazel-airgapped-build.sh` to build the
directory of pre-downloaded dependencies,
2. setups up a network namespace that only allows access to the loopback link
(needed to run Bazel, as Bazel spawns a server that communicates over a socket),
3. attempts to build all targets in this repository, and lastly
4. deletes the airgapped network namespace.
