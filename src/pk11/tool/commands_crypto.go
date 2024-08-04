// Copyright lowRISC contributors (OpenTitan project).
// Licensed under the Apache License, Version 2.0, see LICENSE for details.
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"

	"github.com/lowRISC/ot-provisioning/src/pk11"
)

// cryptoCommands loads general cryptography commands.
func (s *State) cryptoCommands() {
	s.Define(&Command{
		Name:         "export",
		Usage:        "<key>",
		Help:         "exports key in an appropriate X.509-compatible format",
		Args:         []ArgTy{ArgKey},
		NeedsSession: true,

		Run: func(args []any, state *State) (any, error) {
			k, err := args[0].(pk11.Key).ExportKey()
			if err != nil {
				return nil, err
			}

			switch k := k.(type) {
			case *rsa.PublicKey, *ecdsa.PublicKey:
				return x509.MarshalPKIXPublicKey(k)
			case *rsa.PrivateKey, *ecdsa.PrivateKey:
				return x509.MarshalPKCS8PrivateKey(k)
			case pk11.AESKey:
				return []byte(k), nil
			default:
				panic("ExportKey() returned unknown key type: this is a bug")
			}
		},
	})
}
