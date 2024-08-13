// Copyright lowRISC contributors (OpenTitan project).
// Licensed under the Apache License, Version 2.0, see LICENSE for details.
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"bytes"
	"crypto"
	"fmt"
	"testing"

	"golang.org/x/crypto/hkdf"

	"github.com/lowRISC/opentitan-provisioning/src/pk11"
	ts "github.com/lowRISC/opentitan-provisioning/src/pk11/test_support"
)

func TestHKDFNoSalt(t *testing.T) {
	hashes := []crypto.Hash{
		crypto.SHA256, crypto.SHA384, crypto.SHA512,
	}
	s := ts.GetSession(t)
	ts.Check(t, s.Login(pk11.NormalUser, ts.UserPin))

	for _, hash := range hashes {
		name := fmt.Sprintf("%v", hash)
		t.Run(name, func(t *testing.T) {
			root, err := s.ImportKeyMaterial([]byte("a very random string"), nil)
			ts.Check(t, err)

			t.Run("multi-step", func(t *testing.T) {
				prk, err := root.HKDFExtract(hash, nil, nil)
				ts.Check(t, err)

				ko, err := prk.HKDFExpandAES(hash, []byte("cool label"), 128, &pk11.KeyOptions{Extractable: true})
				ts.Check(t, err)
				ki, err := ko.ExportKey()
				ts.Check(t, err)
				got := []byte(ki.(pk11.AESKey))

				hkdf := hkdf.New(hash.New, []byte("a very random string"), nil, []byte("cool label"))
				want := make([]byte, 16)
				hkdf.Read(want)

				if !bytes.Equal(got, want) {
					t.Fatalf("got:%x, want:%x", got, want)
				}
			})

			t.Run("oneshot", func(t *testing.T) {
				ko, err := root.HKDFDeriveAES(hash, nil, []byte("cool label"), 128, &pk11.KeyOptions{Extractable: true})
				ts.Check(t, err)
				ki, err := ko.ExportKey()
				ts.Check(t, err)
				got := []byte(ki.(pk11.AESKey))

				hkdf := hkdf.New(hash.New, []byte("a very random string"), nil, []byte("cool label"))
				want := make([]byte, 16)
				hkdf.Read(want)

				if !bytes.Equal(got, want) {
					t.Fatalf("got:%x, want:%x", got, want)
				}
			})
		})
	}
}

func TestHKDFSliceSalt(t *testing.T) {
	hashes := []crypto.Hash{
		crypto.SHA256, crypto.SHA384, crypto.SHA512,
	}
	s := ts.GetSession(t)
	ts.Check(t, s.Login(pk11.NormalUser, ts.UserPin))

	for _, hash := range hashes {
		name := fmt.Sprintf("%v", hash)
		t.Run(name, func(t *testing.T) {
			root, err := s.ImportKeyMaterial([]byte("a very random string"), nil)
			ts.Check(t, err)

			t.Run("multi-step", func(t *testing.T) {
				prk, err := root.HKDFExtract(hash, []byte("nacl"), nil)
				ts.Check(t, err)

				ko, err := prk.HKDFExpandAES(hash, []byte("cool label"), 128, &pk11.KeyOptions{Extractable: true})
				ts.Check(t, err)
				ki, err := ko.ExportKey()
				ts.Check(t, err)
				got := []byte(ki.(pk11.AESKey))

				hkdf := hkdf.New(hash.New, []byte("a very random string"), []byte("nacl"), []byte("cool label"))
				want := make([]byte, 16)
				hkdf.Read(want)

				if !bytes.Equal(got, want) {
					t.Fatalf("got:%x, want:%x", got, want)
				}
			})

			t.Run("oneshot", func(t *testing.T) {
				ko, err := root.HKDFDeriveAES(hash, []byte("nacl"), []byte("cool label"), 128, &pk11.KeyOptions{Extractable: true})
				ts.Check(t, err)
				ki, err := ko.ExportKey()
				ts.Check(t, err)
				got := []byte(ki.(pk11.AESKey))

				hkdf := hkdf.New(hash.New, []byte("a very random string"), []byte("nacl"), []byte("cool label"))
				want := make([]byte, 16)
				hkdf.Read(want)

				if !bytes.Equal(got, want) {
					t.Fatalf("got:%x, want:%x", got, want)
				}
			})
		})
	}
}

func TestHKDFKeySalt(t *testing.T) {
	hashes := []crypto.Hash{
		crypto.SHA256, crypto.SHA384, crypto.SHA512,
	}
	s := ts.GetSession(t)
	ts.Check(t, s.Login(pk11.NormalUser, ts.UserPin))

	for _, hash := range hashes {
		name := fmt.Sprintf("%v", hash)
		t.Run(name, func(t *testing.T) {
			root, err := s.ImportKeyMaterial([]byte("a very random string"), nil)
			ts.Check(t, err)
			salt, err := s.ImportKeyMaterial([]byte("nacl"), nil)
			ts.Check(t, err)

			t.Run("multi-step", func(t *testing.T) {
				prk, err := root.HKDFExtract(hash, salt, nil)
				ts.Check(t, err)

				ko, err := prk.HKDFExpandAES(hash, []byte("cool label"), 128, &pk11.KeyOptions{Extractable: true})
				ts.Check(t, err)
				ki, err := ko.ExportKey()
				ts.Check(t, err)
				got := []byte(ki.(pk11.AESKey))

				hkdf := hkdf.New(hash.New, []byte("a very random string"), []byte("nacl"), []byte("cool label"))
				want := make([]byte, 16)
				hkdf.Read(want)

				if !bytes.Equal(got, want) {
					t.Fatalf("got:%x, want:%x", got, want)
				}
			})

			t.Run("oneshot", func(t *testing.T) {
				ko, err := root.HKDFDeriveAES(hash, salt, []byte("cool label"), 128, &pk11.KeyOptions{Extractable: true})
				ts.Check(t, err)
				ki, err := ko.ExportKey()
				ts.Check(t, err)
				got := []byte(ki.(pk11.AESKey))

				hkdf := hkdf.New(hash.New, []byte("a very random string"), []byte("nacl"), []byte("cool label"))
				want := make([]byte, 16)
				hkdf.Read(want)

				if !bytes.Equal(got, want) {
					t.Fatalf("got:%x, want:%x", got, want)
				}
			})
		})
	}
}
