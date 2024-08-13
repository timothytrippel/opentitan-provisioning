// Copyright lowRISC contributors (OpenTitan project).
// Licensed under the Apache License, Version 2.0, see LICENSE for details.
// SPDX-License-Identifier: Apache-2.0

// Binary softhsm2 is a wrapper over the softhsm2-util tool that allows
// for easy control of the location where tokens are stored; SoftHSM makes this
// somewhat more complicated than it needs to be.
//
// This binary simply forwards all arguments to softhsm2-util, using
// $REPO_TOP/.softhsm2 as the SoftHSM directory; this path may be overriden using
// the SOFTHSM2_DIR environment variable.
//
// Because it uses runfiles, this binary should be run using bazelisk run.
package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/bazelbuild/rules_go/go/tools/bazel"
	"github.com/lowRISC/opentitan-provisioning/third_party/softhsm2/test_config"
)

func main() {
	repoTop := os.Getenv("BUILD_WORKSPACE_DIRECTORY")

	configPath, err := test_config.MakeSandbox()
	if err != nil {
		fmt.Printf("could not create softhsm2 sandbox: %s\n", err)
		os.Exit(1)
	}
	os.Setenv(test_config.EnvVar, configPath)
	fmt.Printf("set %s to %q\n", test_config.EnvVar, configPath)

	softhsmUtilPath, err := bazel.Runfile("softhsm2/bin/softhsm2-util")
	if err != nil {
		fmt.Printf("could not find softhsm2-util binary: %s\n", err)
		os.Exit(1)
	}

	softhsmSoPath, err := bazel.Runfile("softhsm2/lib/softhsm/libsofthsm2.so")
	if err != nil {
		fmt.Printf("could not find libsofthsm2.so: %s\n", err)
		os.Exit(1)
	}

	cmd := exec.Command(softhsmUtilPath, "--module", softhsmSoPath)
	cmd.Args = append(cmd.Args, os.Args[1:]...)
	cmd.Dir = repoTop
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if exit, ok := err.(*exec.ExitError); ok {
		os.Exit(exit.ExitCode())
	}

	if err != nil {
		fmt.Printf("could not exec into softhsm2-util: %s\n", err)
		os.Exit(1)
	}
}
