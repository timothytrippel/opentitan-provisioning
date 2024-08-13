// Copyright lowRISC contributors (OpenTitan project).
// Licensed under the Apache License, Version 2.0, see LICENSE for details.
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"errors"
	"fmt"
	"strings"

	"github.com/lowRISC/opentitan-provisioning/src/pk11"
)

// pk11Commands loads general PCKS#11 commands.
func (s *State) pk11Commands() {
	s.Define(&Command{
		Name:  "dump",
		Usage: "",
		Help:  "dumps information about the PKCS#11 plugin",
		Args:  []ArgTy{},

		Run: func(args []any, state *State) (any, error) {
			fmt.Println(state.m.Dump())
			return nil, nil
		},
	})

	s.Define(&Command{
		Name:  "open-session",
		Usage: "<slot>",
		Help:  "opens a session on the nth token; cannot be used if a session is open",
		Args:  []ArgTy{ArgInt},

		Run: func(args []any, state *State) (any, error) {
			if state.s != nil {
				return nil, errors.New("open session already exists")
			}

			toks, err := state.m.Tokens()
			if err != nil {
				return nil, err
			}

			slot := int(args[0].(int64))
			if slot < 0 || slot >= len(toks) {
				return nil, fmt.Errorf("slot index out of range; want 0..%d, got %d", len(toks), slot)
			}

			sess, err := toks[slot].OpenSession()
			if err != nil {
				return nil, err
			}

			state.s = sess
			return nil, nil
		},
	})

	s.Define(&Command{
		Name:         "login",
		Usage:        "<pin>",
		Help:         "logs into the current session as the normal user",
		Args:         []ArgTy{ArgBytes},
		NeedsSession: true,

		Run: func(args []any, state *State) (any, error) {
			return nil, state.s.Login(pk11.NormalUser, string(args[0].([]byte)))
		},
	})

	s.Define(&Command{
		Name:         "object",
		Usage:        "<class> <uid or object>",
		Help:         "looks up a PKCS#11 object of the given class, with the given UID or the given object's UID\nclass must be one of pub, priv, or sym",
		Args:         []ArgTy{ArgBytes, ArgBytes | ArgObj},
		NeedsSession: true,

		Run: func(args []any, state *State) (any, error) {
			var uid []byte
			switch v := args[1].(type) {
			case []byte:
				uid = v
			case pk11.Object:
				var err error
				uid, err = v.UID()
				if err != nil {
					return nil, err
				}
			}

			switch strings.ToLower(string(args[0].([]byte))) {
			case "pub":
				return state.s.FindPublicKey(uid)
			case "priv":
				return state.s.FindPrivateKey(uid)
			case "sym":
				return state.s.FindSecretKey(uid)
			default:
				return nil, fmt.Errorf("unknown object class %q", args[0])
			}
		},
	})

	s.Define(&Command{
		Name:         "uid",
		Usage:        "<object>",
		Help:         "gets the UID of a PKCS#11 object",
		Args:         []ArgTy{ArgObj},
		NeedsSession: true,

		Run: func(args []any, state *State) (any, error) {
			return args[0].(pk11.Object).UID()
		},
	})

	s.Define(&Command{
		Name:         "keys",
		Help:         "print the ids of all PKCS#11 keys visible in the current session",
		Args:         []ArgTy{},
		NeedsSession: true,

		Run: func(args []any, state *State) (any, error) {
			objs, err := state.s.FindAllKeys()
			if err != nil {
				return nil, err
			}

			for _, o := range objs {
				switch o.(type) {
				case pk11.PublicKey:
					fmt.Print("pub  ")
				case pk11.PrivateKey:
					fmt.Print("priv ")
				case pk11.SecretKey:
					fmt.Print("sym  ")
				}
				s, err := Stringify(o)
				if err != nil {
					fmt.Println(err)
				}
				fmt.Println(s)
			}

			return nil, nil
		},
	})
}
