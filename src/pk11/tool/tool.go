// Copyright lowRISC contributors (OpenTitan project).
// Licensed under the Apache License, Version 2.0, see LICENSE for details.
// SPDX-License-Identifier: Apache-2.0

// Binary //src/crypto/pk11/tool implements a REPL for interacting with
// the pk11 package, primarily aimed at making it easy to test the library
// against particular HSM devices, or debugging the library itself.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/bazelbuild/rules_go/go/tools/bazel"

	"github.com/lowRISC/opentitan-provisioning/src/pk11/tool/commands"
	"github.com/lowRISC/opentitan-provisioning/src/pk11/tool/lex"
	"github.com/lowRISC/opentitan-provisioning/third_party/softhsm2/test_config"
)

const openscSO = "/usr/lib/x86_64-linux-gnu/pkcs11/opensc-pkcs11.so"

var (
	plugin  = flag.String("plugin", "", "path to a PKCS#11 plugin library")
	softhsm = flag.Bool("softhsm", false, "use a SoftHSM sandbox instead of a harware token; see //third_party/softhsm:test_config.go")
	slot    = flag.Int("slot", -1, "automatically open a session on startup on the given slot")
	pin     = flag.String("pin", "", "automatically log into a session (see -slot) as the regular user with the given pin")
	script  = flag.String("script", "", "path to a script to run; if not set, drops into an interactive session")
)

func main() {
	flag.Parse()

	if *plugin == "" {
		*plugin = openscSO
	}

	if *softhsm {
		config, err := test_config.MakeSandbox()
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
		os.Setenv(test_config.EnvVar, config)
		fmt.Printf("created SoftHSM config at %q\n", config)

		softhsmSoPath, err := bazel.Runfile("softhsm2/lib/softhsm/libsofthsm2.so")
		if err != nil {
			fmt.Printf("could not find libsofthsm2.so: %s\n", err)
			os.Exit(2)
		}
		*plugin = softhsmSoPath
	}

	// Switch to $REPO_TOP at this point; this means that we will read and write
	// files outside of the sandbox.
	if err := os.Chdir(os.Getenv("BUILD_WORKSPACE_DIRECTORY")); err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	state, err := commands.New(*plugin)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	if *slot >= 0 {
		cmd := fmt.Sprintf("open-session %d", *slot)
		lexer := lex.New(strings.NewReader(cmd))
		_, _, errs := state.Interpret(lexer)

		if len(errs) != 0 {
			fmt.Print("could not open session: ")
			for _, err := range errs {
				fmt.Printf("%s", err)
			}
			fmt.Println()
			os.Exit(2)
		}
	}

	if *pin != "" {
		cmd := fmt.Sprintf("login %q", *pin)
		lexer := lex.New(strings.NewReader(cmd))
		_, _, errs := state.Interpret(lexer)

		if len(errs) != 0 {
			fmt.Print("could not log in: ")
			for _, err := range errs {
				fmt.Printf("%s", err)
			}
			fmt.Println()
			os.Exit(2)
		}
	}

	input := os.Stdin
	isScript := *script != ""
	if isScript {
		f, err := os.Open(*script)
		if err != nil {
			fmt.Printf("could not open script file %q: %s\n", *script, err)
			os.Exit(2)
		}
		defer f.Close()
		input = f
	}

	lexer := lex.New(input)
	var eof bool
	for !eof {
		if !isScript {
			fmt.Print("pk11> ")
		}
		toks, ret, errs := state.Interpret(lexer)
		if len(toks) > 0 {
			_, eof = toks[len(toks)-1].Value.(lex.EOF)
		}

		for _, err := range errs {
			if isScript {
				fmt.Printf("could not execute command `%s`: %s\n", lex.StringTokens(toks), err)
				os.Exit(2)
			} else {
				fmt.Printf("# error: %s\n", err)
			}
		}
		if len(errs) != 0 {
			continue
		}

		if !isScript {
			s, err := commands.Stringify(ret)
			if err != nil {
				fmt.Printf("# error: %s\n", err)
				continue
			}
			if s != "" {
				fmt.Println(s)
			}
		}
	}
}
