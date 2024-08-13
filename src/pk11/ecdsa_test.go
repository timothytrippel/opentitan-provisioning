// Copyright lowRISC contributors (OpenTitan project).
// Licensed under the Apache License, Version 2.0, see LICENSE for details.
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"fmt"
	"math/big"
	"math/rand"
	"testing"

	"github.com/lowRISC/opentitan-provisioning/src/pk11"
	ts "github.com/lowRISC/opentitan-provisioning/src/pk11/test_support"
)

func TestECDSA(t *testing.T) {
	tests := []struct {
		curve elliptic.Curve
		hash  crypto.Hash
	}{
		{elliptic.P256(), crypto.SHA256},
	}

	s := ts.GetSession(t)
	ts.Check(t, s.Login(pk11.NormalUser, ts.UserPin))

	for _, test := range tests {
		name := fmt.Sprintf("%s-%s", test.curve.Params().Name, test.hash)
		t.Run(name, func(t *testing.T) {
			kp, err := s.GenerateECDSA(test.curve, nil)
			ts.Check(t, err)

			var r, s big.Int
			rBytes, sBytes, err := kp.SignECDSA(test.hash, []byte(name))
			ts.Check(t, err)
			r.SetBytes(rBytes)
			s.SetBytes(sBytes)

			pub, err := kp.PublicKey.ExportKey()
			ts.Check(t, err)

			hash := ts.MakeHash(test.hash, []byte(name))
			if !ecdsa.Verify(pub.(*ecdsa.PublicKey), hash, &r, &s) {
				t.Fatal("verification failed")
			}
		})
	}
}

func TestECDSASigner(t *testing.T) {
	tests := []struct {
		curve elliptic.Curve
		hash  crypto.Hash
	}{
		{elliptic.P256(), crypto.SHA256},
	}

	s := ts.GetSession(t)
	ts.Check(t, s.Login(pk11.NormalUser, ts.UserPin))

	for _, test := range tests {
		name := fmt.Sprintf("%s-%s", test.curve.Params().Name, test.hash)
		t.Run(name, func(t *testing.T) {
			kp, err := s.GenerateECDSA(test.curve, nil)
			ts.Check(t, err)

			signer, err := kp.Signer()
			ts.Check(t, err)

			hash := ts.MakeHash(test.hash, []byte(name))
			sig, err := signer.Sign(nil, hash, test.hash)
			ts.Check(t, err)
			if !ecdsa.VerifyASN1(signer.Public().(*ecdsa.PublicKey), hash, sig) {
				t.Fatal("verification failed")
			}
		})
	}
}

func TestECDSAImport(t *testing.T) {
	tests := []struct {
		curve elliptic.Curve
		hash  crypto.Hash
	}{
		{elliptic.P256(), crypto.SHA256},
	}

	s := ts.GetSession(t)
	ts.Check(t, s.Login(pk11.NormalUser, ts.UserPin))

	for _, test := range tests {
		name := fmt.Sprintf("%s-%s", test.curve.Params().Name, test.hash)
		t.Run(name, func(t *testing.T) {
			// This does not need to be secure randomness.
			rand := rand.New(rand.NewSource(0))
			key, err := ecdsa.GenerateKey(test.curve, rand)
			ts.Check(t, err)

			ki, err := s.ImportKey(key, nil)
			ts.Check(t, err)
			ko := ki.(pk11.PrivateKey)

			var r, s big.Int
			rBytes, sBytes, err := ko.SignECDSA(test.hash, []byte(name))
			ts.Check(t, err)
			r.SetBytes(rBytes)
			s.SetBytes(sBytes)

			hash := ts.MakeHash(test.hash, []byte(name))
			if !ecdsa.Verify(&key.PublicKey, hash, &r, &s) {
				t.Fatal("verification failed")
			}
		})
	}
}

func TestECDSALookup(t *testing.T) {
	tests := []struct {
		curve elliptic.Curve
		hash  crypto.Hash
	}{
		{elliptic.P256(), crypto.SHA256},
	}

	s := ts.GetSession(t)
	ts.Check(t, s.Login(pk11.NormalUser, ts.UserPin))

	for _, test := range tests {
		name := fmt.Sprintf("%s-%s", test.curve.Params().Name, test.hash)
		t.Run(name, func(t *testing.T) {
			kp, err := s.GenerateECDSA(test.curve, nil)
			ts.Check(t, err)

			uid, err := kp.PublicKey.UID()
			ts.Check(t, err)

			kp, err = s.FindKeyPair(uid)
			ts.Check(t, err)

			var r, s big.Int
			rBytes, sBytes, err := kp.SignECDSA(test.hash, []byte(name))
			ts.Check(t, err)
			r.SetBytes(rBytes)
			s.SetBytes(sBytes)

			pub, err := kp.PublicKey.ExportKey()
			ts.Check(t, err)

			hash := ts.MakeHash(test.hash, []byte(name))
			if !ecdsa.Verify(pub.(*ecdsa.PublicKey), hash, &r, &s) {
				t.Fatal("verification failed")
			}
		})
	}
}
