// Copyright lowRISC contributors (OpenTitan project).
// Licensed under the Apache License, Version 2.0, see LICENSE for details.
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"crypto/elliptic"
	"crypto/x509"
	"encoding/asn1"
	"fmt"
	"math/big"
	"strings"

	"github.com/lowRISC/ot-provisioning/src/pk11"
)

// namedCurve tries to parse name into an elliptic.Curve value.
func namedCurve(name string) (c elliptic.Curve, ok bool) {
	ok = true
	switch strings.ToLower(name) {
	case "p256", "p-256", "secp256r1":
		c = elliptic.P256()
	default:
		ok = false
	}
	return
}

// ecdsaCommands loads ECDSA-specific commands.
func (s *State) ecdsaCommands() {
	s.Define(&Command{
		Name:         "gen-ecdsa",
		Usage:        "<curve>",
		Help:         "generates a new ECDSA key on the given curve; returns the public key",
		Args:         []ArgTy{ArgBytes},
		NeedsSession: true,

		Run: func(args []any, state *State) (any, error) {
			curveName := string(args[0].([]byte))
			curve, ok := namedCurve(curveName)
			if !ok {
				return nil, fmt.Errorf("unknown curve %q", curveName)
			}

			k, err := offload(fmt.Sprintf("generating new ECDSA key on curve %s", curve), func() (any, error) {
				return state.s.GenerateECDSA(curve, nil)
			})
			if err != nil {
				return nil, err
			}

			return k.(pk11.KeyPair).PublicKey, nil
		},
	})

	s.Define(&Command{
		Name:         "import-ecdsa",
		Usage:        "<priv-key> <sensitive>",
		Help:         "imports an ECDSA private key in SEC 1 DER form; returns the private key\nif sensitive is set, the key will not be exportable",
		Args:         []ArgTy{ArgBytes, ArgBool},
		NeedsSession: true,

		Run: func(args []any, state *State) (any, error) {
			key, err := x509.ParseECPrivateKey(args[0].([]byte))
			if err != nil {
				return nil, err
			}

			sensitive := args[1].(bool)
			return state.s.ImportKey(key, &pk11.KeyOptions{Extractable: !sensitive})
		},
	})

	s.Define(&Command{
		Name:         "sign-ecdsa",
		Usage:        "<hash> <key> <message> <format>",
		Help:         "constructs an ECDSA signature using the given hash and key for message\nformat may be fixed or asn1, default fixed",
		Args:         []ArgTy{ArgBytes, ArgPrivate, ArgBytes, ArgBytes | ArgOptional},
		NeedsSession: true,

		Run: func(args []any, state *State) (any, error) {
			hash, err := parseHash(string(args[0].([]byte)))
			if err != nil {
				return nil, err
			}

			key := args[1].(pk11.PrivateKey)
			msg := args[2].([]byte)

			var format string
			if args[3] != nil {
				format = string(args[3].([]byte))
			}

			var isASN1 bool
			switch strings.ToLower(format) {
			case "", "fixed":
				isASN1 = false
			case "asn1":
				isASN1 = true
			default:
				return nil, fmt.Errorf("unknown format %q", args[3])
			}

			var r, s []byte
			_, err = offload("generating new ECDSA signature", func() (any, error) {
				var err error
				r, s, err = key.SignECDSA(hash, msg)
				return nil, err
			})
			if err != nil {
				return nil, err
			}

			if isASN1 {
				type ecdsaASN1 struct {
					R, S *big.Int
				}

				sig := ecdsaASN1{new(big.Int), new(big.Int)}
				sig.R.SetBytes(r)
				sig.S.SetBytes(s)

				return asn1.Marshal(sig)
			} else {
				buf := make([]byte, len(r)+len(s))
				copy(buf[:len(r)], r)
				copy(buf[len(r):], s)
				return buf, nil
			}
		},
	})
}
