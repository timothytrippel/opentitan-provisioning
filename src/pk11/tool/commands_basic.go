// Copyright lowRISC contributors (OpenTitan project).
// Licensed under the Apache License, Version 2.0, see LICENSE for details.
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/lowRISC/ot-provisioning/src/pk11/tool/lex"
)

// basicCommands loads the basic, non-PKCS#11-specific commands into an interpreter.
func (s *State) basicCommands() {
	s.Define(&Command{
		Name:  "help",
		Usage: "",
		Help:  "prints this list of commands",
		Args:  []ArgTy{},

		Run: func(args []any, state *State) (any, error) {
			var names []string
			for k := range state.cmds {
				names = append(names, k)
			}

			sort.Strings(names)
			for _, k := range names {
				v := state.cmds[k]
				fmt.Printf("%s %s\n", v.Name, v.Usage)
				for _, line := range strings.Split(v.Help, "\n") {
					fmt.Printf("  %s\n", line)
				}
			}
			return nil, nil
		},
	})

	s.Define(&Command{
		Name:  "set",
		Usage: "<var> <cmd...>",
		Help:  "runs cmd and places its result (if any) in var",
		Args:  []ArgTy{ArgBytes, ArgTokens},

		Run: func(args []any, state *State) (any, error) {
			if len(args) < 2 {
				return nil, errors.New("expected at least two arguments")
			}
			name := string(args[0].([]byte))
			toks := args[1].([]lex.Token)
			val, err := state.Run(toks...)
			if err != nil {
				return nil, err
			} else if val == nil {
				var cmd strings.Builder
				for i, tok := range toks {
					if i != 0 {
						fmt.Fprint(&cmd, " ")
					}
					fmt.Fprint(&cmd, tok)
				}
				return nil, fmt.Errorf("command %q did not produce a value", cmd.String())
			}

			state.vars[name] = val
			return val, nil // Return val so it gets printed.
		},
	})

	s.Define(&Command{
		Name:  "string",
		Usage: "<var> <string>",
		Help:  "sets var to the contents of string",
		Args:  []ArgTy{ArgBytes, ArgBytes},

		Run: func(args []any, state *State) (any, error) {
			state.vars[string(args[0].([]byte))] = args[1].([]byte)
			return nil, nil
		},
	})

	s.Define(&Command{
		Name:  "read",
		Usage: "<file>",
		Help:  "reads a file as byte array",
		Args:  []ArgTy{ArgBytes},

		Run: func(args []any, state *State) (any, error) {
			return os.ReadFile(string(args[0].([]byte)))
		},
	})

	s.Define(&Command{
		Name:  "write",
		Usage: "<bytes> <file>",
		Help:  "writes a byte array variable to a file",
		Args:  []ArgTy{ArgBytes, ArgBytes},

		Run: func(args []any, state *State) (any, error) {
			// Return the byte array so it gets printed.
			return args[0].([]byte), os.WriteFile(string(args[1].([]byte)), args[0].([]byte), 0777)
		},
	})
}
