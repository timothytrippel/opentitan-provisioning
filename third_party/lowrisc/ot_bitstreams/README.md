# Updating the Bitstreams

CW310 and CW340 bitstreams are checked into this repo to enable emulating provisioning flow on an OpenTitan FPGA DUT.
The bitstreams checked into this repo represent the state of a chip as it would be if it were entering CP, meaning the chip has a completely empty OTP and flash, except the lifecycle state is TEST\_UNLOCKED0.
To update these bitstreams with the latest pinned version of the lowrisc\_opentitan repo, do the following:
1. Make sure the `_OPENTITAN_VERSION` tag in `third_party/lowrisc/repos.bzl` matches the `_OT_REPO_BRANCH` tag in `third_party/lowrisc/ot_bitstreams/build-ot-bitstreams.sh`.
1. `./third_party/lowrisc/ot_bitstreams/build-ot-bitstreams.sh <path to opentitan repo>`
1. `git add third_party/lowrisc/ot_bitstreams/cp_cw340.bit`
1. `git add third_party/lowrisc/ot_bitstreams/cp_hyper310.bit`
1. Commit the newly added bitstreams.
