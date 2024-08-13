// Copyright lowRISC contributors (OpenTitan project).
// Licensed under the Apache License, Version 2.0, see LICENSE for details.
// SPDX-License-Identifier: Apache-2.0

package commands

import "github.com/lowRISC/opentitan-provisioning/src/pk11"

// ecdsaCommands loads ECDSA-specific commands.
func (s *State) aesCommands() {
	s.Define(&Command{
		Name:         "gen-aes",
		Usage:        "<bits> <sensitive>",
		Help:         "generates a new AES key on the given number of bits\nif sensitive is set, the key will not be exportable",
		Args:         []ArgTy{ArgInt, ArgBool},
		NeedsSession: true,

		Run: func(args []any, state *State) (any, error) {
			bits := args[0].(int64)
			sensitive := args[1].(bool)
			return offload("generating new AES key", func() (any, error) {
				return state.s.GenerateAES(uint(bits), &pk11.KeyOptions{Extractable: !sensitive})
			})
		},
	})

	s.Define(&Command{
		Name:         "import-aes-raw",
		Usage:        "<key> <sensitive>",
		Help:         "imports raw AES key bytes\nif sensitive is set, the key will not be exportable",
		Args:         []ArgTy{ArgBytes, ArgBool},
		NeedsSession: true,

		Run: func(args []any, state *State) (any, error) {
			kBytes := args[0].([]byte)
			sensitive := args[1].(bool)
			return state.s.ImportKey(pk11.AESKey(kBytes), &pk11.KeyOptions{Extractable: !sensitive})
		},
	})

	s.Define(&Command{
		Name:         "seal-aes-gcm",
		Usage:        "<key> <iv> <aad> <tagbits> <message>",
		Help:         "seals an AES-GCM AEAD with the given parameters",
		Args:         []ArgTy{ArgSecret, ArgBytes, ArgBytes, ArgInt, ArgBytes},
		NeedsSession: true,

		Run: func(args []any, state *State) (any, error) {
			key := args[0].(pk11.SecretKey)
			iv := args[1].([]byte)
			aad := args[2].([]byte)
			tagbits := int(args[3].(int64))
			msg := args[4].([]byte)

			return offload("sealing new AES-GCM AEAD", func() (any, error) {
				var err error
				// Drop the potentially-HSM-provided iv on the ground for now.
				aead, _, err := key.SealAESGCM(iv, aad, tagbits, msg)
				return aead, err
			})
		},
	})

	s.Define(&Command{
		Name:         "unseal-aes-gcm",
		Usage:        "<key> <iv> <aad> <tagbits> <aead>",
		Help:         "unseals an AES-GCM AEAD with the given parameters",
		Args:         []ArgTy{ArgSecret, ArgBytes, ArgBytes, ArgInt, ArgBytes},
		NeedsSession: true,

		Run: func(args []any, state *State) (any, error) {
			key := args[0].(pk11.SecretKey)
			iv := args[1].([]byte)
			aad := args[2].([]byte)
			tagbits := int(args[3].(int64))
			aead := args[4].([]byte)

			return offload("unsealing AES-GCM AEAD", func() (any, error) {
				return key.UnsealAESGCM(iv, aad, tagbits, aead)
			})
		},
	})
}
