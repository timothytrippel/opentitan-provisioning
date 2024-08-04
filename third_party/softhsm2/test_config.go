// Copyright lowRISC contributors (OpenTitan project).
// Licensed under the Apache License, Version 2.0, see LICENSE for details.
// SPDX-License-Identifier: Apache-2.0

// Package test_config provides utilities for creating a SoftHSM sandbox in a
// specific location.
//
// The primary purpose of this library is creating a sandbox that can have
// tokens scribbled into it for testing, either as part of a unit test or
// manual testing, which is not tied to the global SoftHSM pool (which is
// usually owned by the root user, if it exists at all!).
package test_config

import (
	"os"
	"path/filepath"
	"text/template"
)

const (
	// The environment variable used by SoftHSM for finding its configuration files.
	//
	// When spawning a subprocess or opening the PKCS#11 plugin, the environment should
	// be set to this location.
	EnvVar = "SOFTHSM2_CONF"

	// The environment variable used by this library to find the desired SoftHSM sandbox
	// location, if specified.
	DirVar = "SOFTHSM2_DIR"
)

// The default path to place a sandbox into.
var DefaultPath = filepath.Join(os.Getenv("BUILD_WORKSPACE_DIRECTORY"), ".softhsm2")

var confTmpl = template.Must(template.New("softhsm.conf").Parse(`
directories.tokendir = {{.}}
objectstore.backend = file
objectstore.umask = 0077

log.level = DEBUG
slots.removable = false
slots.mechanisms = ALL
library.reset_on_fork = false
`))

// MakeSandbox configures the sandbox that SoftHSM writes its configuration files to.
//
// The sandbox will be placed in $SOFTHSM2_DIR, if that environment variable is set,
// or otherwise $REPO_TOP/.softhsm2
//
// This function creates the appropriate files and directories in the sandbox.
func MakeSandbox() (string, error) {
	path, ok := os.LookupEnv(DirVar)
	if !ok {
		path = DefaultPath
	}
	return MakeSandboxIn(path)
}

// MakeSandboxIn is like MakeSandbox but with an explicit path.
func MakeSandboxIn(sandboxPath string) (string, error) {
	sandboxPath, err := filepath.Abs(sandboxPath)
	if err != nil {
		return "", err
	}

	var (
		rwConfPath = filepath.Join(sandboxPath, "softhsm.conf")
		rwTokenDir = filepath.Join(sandboxPath, "tokens")
	)

	if err := os.MkdirAll(sandboxPath, 0777); err != nil {
		return "", err
	}

	out, err := os.Create(rwConfPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	err = confTmpl.Execute(out, rwTokenDir)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(rwTokenDir, 0777); err != nil {
		return "", err
	}

	return rwConfPath, nil
}
