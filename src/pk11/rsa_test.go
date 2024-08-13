// Copyright lowRISC contributors (OpenTitan project).
// Licensed under the Apache License, Version 2.0, see LICENSE for details.
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"crypto"
	"crypto/rsa"
	"fmt"
	"math/rand"
	"testing"

	"github.com/lowRISC/opentitan-provisioning/src/pk11"
	ts "github.com/lowRISC/opentitan-provisioning/src/pk11/test_support"
)

func TestRSA(t *testing.T) {
	type tt struct {
		m, e uint
	}
	var tests []tt

	for _, m := range []uint{2048, 3072, 4096} {
		for _, e := range []uint{3, 0x010001} {
			tests = append(tests, tt{m, e})
		}
	}

	s := ts.GetSession(t)
	ts.Check(t, s.Login(pk11.NormalUser, ts.UserPin))

	for _, test := range tests {
		name := fmt.Sprintf("%v", test)
		t.Run(name, func(t *testing.T) {
			// RSA keygen is hilariously slow. To speed up these tests, we only
			// have one test for each m, e pair.
			kp, err := s.GenerateRSA(test.m, test.e, nil)
			ts.Check(t, err)

			pub, err := kp.PublicKey.ExportKey()
			ts.Check(t, err)
			pubkey := pub.(*rsa.PublicKey)

			hashes := []crypto.Hash{crypto.SHA256, crypto.SHA384, crypto.SHA512}
			for _, h := range hashes {
				t.Run(h.String(), func(t *testing.T) {
					hash := ts.MakeHash(h, []byte(name))

					t.Run("pkcs#1", func(t *testing.T) {
						sig, err := kp.SignRSAPKCS1v15(h, []byte(name))
						ts.Check(t, err)
						ts.Check(t, rsa.VerifyPKCS1v15(pubkey, h, hash, sig))
					})

					t.Run("pss", func(t *testing.T) {
						opts := &rsa.PSSOptions{rsa.PSSSaltLengthAuto, h}
						sig, err := kp.SignRSAPSS(opts, []byte(name))
						ts.Check(t, err)
						ts.Check(t, rsa.VerifyPSS(pubkey, h, hash, sig, opts))
					})

					t.Run("pkcs#1-crypto.Signer", func(t *testing.T) {
						signer, err := kp.Signer()
						ts.Check(t, err)

						sig, err := signer.Sign(nil, hash, h)
						ts.Check(t, err)
						ts.Check(t, rsa.VerifyPKCS1v15(pubkey, h, hash, sig))
					})

					t.Run("pss-crypto.Signer", func(t *testing.T) {
						opts := &rsa.PSSOptions{rsa.PSSSaltLengthAuto, h}
						signer, err := kp.Signer()
						ts.Check(t, err)

						sig, err := signer.Sign(nil, hash, opts)
						ts.Check(t, err)
						ts.Check(t, rsa.VerifyPSS(pubkey, h, hash, sig, opts))
					})
				})
			}
		})
	}
}

func TestRSAImport(t *testing.T) {
	tests := []int{2048, 3072, 4096}

	s := ts.GetSession(t)
	ts.Check(t, s.Login(pk11.NormalUser, ts.UserPin))

	for _, test := range tests {
		name := fmt.Sprintf("%v", test)
		t.Run(name, func(t *testing.T) {
			// This does not need to be secure randomness.
			rand := rand.New(rand.NewSource(0))
			// RSA keygen is hilariously slow. To speed up these tests, we only
			// have one test for each m, e pair.
			key, err := rsa.GenerateKey(rand, test)
			ts.Check(t, err)

			ki, err := s.ImportKey(key, nil)
			ts.Check(t, err)
			ko := ki.(pk11.PrivateKey)

			hashes := []crypto.Hash{crypto.SHA256, crypto.SHA384, crypto.SHA512}
			for _, h := range hashes {
				t.Run(h.String(), func(t *testing.T) {
					hash := ts.MakeHash(h, []byte(name))

					t.Run("pkcs#1", func(t *testing.T) {
						sig, err := ko.SignRSAPKCS1v15(h, []byte(name))
						ts.Check(t, err)
						ts.Check(t, rsa.VerifyPKCS1v15(&key.PublicKey, h, hash, sig))
					})

					t.Run("pss", func(t *testing.T) {
						opts := &rsa.PSSOptions{rsa.PSSSaltLengthAuto, h}
						sig, err := ko.SignRSAPSS(opts, []byte(name))
						ts.Check(t, err)
						ts.Check(t, rsa.VerifyPSS(&key.PublicKey, h, hash, sig, opts))
					})

					t.Run("pkcs#1-crypto.Signer", func(t *testing.T) {
						signer := pk11.RSASigner{&key.PublicKey, ko}

						sig, err := signer.Sign(nil, hash, h)
						ts.Check(t, err)
						ts.Check(t, rsa.VerifyPKCS1v15(&key.PublicKey, h, hash, sig))
					})

					t.Run("pss-crypto.Signer", func(t *testing.T) {
						opts := &rsa.PSSOptions{rsa.PSSSaltLengthAuto, h}
						signer := pk11.RSASigner{&key.PublicKey, ko}

						sig, err := signer.Sign(nil, hash, opts)
						ts.Check(t, err)
						ts.Check(t, rsa.VerifyPSS(&key.PublicKey, h, hash, sig, opts))
					})
				})
			}
		})
	}
}

func TestRSALookup(t *testing.T) {
	tests := []struct {
		m, e uint
		hash crypto.Hash
	}{
		{2048, 3, crypto.SHA256},
		{2048, 0x010001, crypto.SHA256},
	}

	s := ts.GetSession(t)
	ts.Check(t, s.Login(pk11.NormalUser, ts.UserPin))

	for _, test := range tests {
		name := fmt.Sprintf("%v", test)
		t.Run(name, func(t *testing.T) {
			kp, err := s.GenerateRSA(test.m, test.e, nil)
			ts.Check(t, err)

			uid, err := kp.PublicKey.UID()
			ts.Check(t, err)

			kp, err = s.FindKeyPair(uid)
			ts.Check(t, err)

			pubIface, err := kp.PublicKey.ExportKey()
			ts.Check(t, err)
			pubkey := pubIface.(*rsa.PublicKey)

			hash := ts.MakeHash(test.hash, []byte(name))
			t.Run("pkcs#1", func(t *testing.T) {
				sig, err := kp.SignRSAPKCS1v15(test.hash, []byte(name))
				ts.Check(t, err)
				ts.Check(t, rsa.VerifyPKCS1v15(pubkey, test.hash, hash, sig))
			})

			t.Run("pss", func(t *testing.T) {
				opts := &rsa.PSSOptions{rsa.PSSSaltLengthAuto, test.hash}
				sig, err := kp.SignRSAPSS(opts, []byte(name))
				ts.Check(t, err)
				ts.Check(t, rsa.VerifyPSS(pubkey, test.hash, hash, sig, opts))
			})

			t.Run("pkcs#1-crypto.Signer", func(t *testing.T) {
				signer, err := kp.Signer()
				ts.Check(t, err)

				sig, err := signer.Sign(nil, hash, test.hash)
				ts.Check(t, err)
				ts.Check(t, rsa.VerifyPKCS1v15(pubkey, test.hash, hash, sig))
			})

			t.Run("pss-crypto.Signer", func(t *testing.T) {
				opts := &rsa.PSSOptions{rsa.PSSSaltLengthAuto, test.hash}
				signer, err := kp.Signer()
				ts.Check(t, err)

				sig, err := signer.Sign(nil, hash, opts)
				ts.Check(t, err)
				ts.Check(t, rsa.VerifyPSS(pubkey, test.hash, hash, sig, opts))
			})
		})
	}
}
