// Copyright lowRISC contributors (OpenTitan project).
// Licensed under the Apache License, Version 2.0, see LICENSE for details.
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"crypto/rsa"
	"crypto/x509"
	"fmt"

	"github.com/lowRISC/ot-provisioning/src/pk11"
)

// rsaCommands loads RSA-specific commands.
func (s *State) rsaCommands() {
	s.Define(&Command{
		Name:         "gen-rsa",
		Usage:        "<mod-bits> <exponent>",
		Help:         "generates a new RSA key with the given parameters; returns the public key",
		Args:         []ArgTy{ArgInt, ArgInt},
		NeedsSession: true,

		Run: func(args []any, state *State) (any, error) {
			m := uint(args[0].(int64))
			e := uint(args[1].(int64))

			k, err := offload(fmt.Sprintf("generating new RSA key; len(m)=%d, e=%d", m, e), func() (any, error) {
				return state.s.GenerateRSA(m, e, nil)
			})
			if err != nil {
				return nil, err
			}

			return k.(pk11.KeyPair).PublicKey, nil
		},
	})

	s.Define(&Command{
		Name:         "import-rsa",
		Usage:        "<priv-key> <sensitive>",
		Help:         "imports an ECDSA private key in PKCS#1 DER form; returns the private key\nif sensitive is set, the key will not be exportable",
		Args:         []ArgTy{ArgBytes, ArgBool},
		NeedsSession: true,

		Run: func(args []any, state *State) (any, error) {
			kBytes := args[0].([]byte)
			key, err := x509.ParsePKCS1PrivateKey(kBytes)
			if err != nil {
				return nil, err
			}

			sensitive := args[1].(bool)
			return state.s.ImportKey(key, &pk11.KeyOptions{Extractable: !sensitive})
		},
	})

	s.Define(&Command{
		Name:         "sign-rsa-pkcs1",
		Usage:        "<hash> <key> <message>",
		Help:         "constructs an PKCS#1 v1.5 RSA signature using the given hash and key for message",
		Args:         []ArgTy{ArgBytes, ArgPrivate, ArgBytes},
		NeedsSession: true,

		Run: func(args []any, state *State) (any, error) {
			hash, err := parseHash(string(args[0].([]byte)))
			if err != nil {
				return nil, err
			}

			key := args[1].(pk11.PrivateKey)
			msg := args[2].([]byte)

			return offload("generating new RSA signature", func() (any, error) {
				return key.SignRSAPKCS1v15(hash, msg)
			})
		},
	})

	s.Define(&Command{
		Name:         "sign-rsa-pss",
		Usage:        "<salt-len> <hash> <key> <message>",
		Help:         "constructs an RSA-PSS signature using the given hash, salt length (in bytes), and key for message",
		Args:         []ArgTy{ArgInt, ArgBytes, ArgPrivate, ArgBytes},
		NeedsSession: true,

		Run: func(args []any, state *State) (any, error) {
			hash, err := parseHash(string(args[1].([]byte)))
			if err != nil {
				return nil, err
			}

			saltLen := int(args[0].(int64))
			key := args[2].(pk11.PrivateKey)
			msg := args[3].([]byte)

			return offload("generating new RSA signature", func() (any, error) {
				return key.SignRSAPSS(&rsa.PSSOptions{saltLen, hash}, msg)
			})
		},
	})
}
