// Copyright lowRISC contributors (OpenTitan project).
// Licensed under the Apache License, Version 2.0, see LICENSE for details.
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"math/rand"
	"testing"

	"github.com/lowRISC/ot-provisioning/src/pk11"
	ts "github.com/lowRISC/ot-provisioning/src/pk11/test_support"
)

const Plain = `
Muchos años después, frente al pelotón de fusilamiento, el coronel Aureliano
Buendía había de recordar aquella tarde remota en que su padre lo llevó a
conocer el hielo. Macondo era entonces una aldea de veinte casas de barro y
cañabrava construida a la orilla de un río de aguas diáfanas que se precipitaban
por un lecho de piedras pulidas, blancas y enormes como huevos prehistóricos. El
mundo era tan reciente, que muchas cosas carecían de nombre, y para mencionarlas
había que señalarlas con el dedo.
`

func TestAESGCMBadParams(t *testing.T) {
	tests := []struct {
		keyBitLen uint
		tagBits   int
	}{}
	s := ts.GetSession(t)
	ts.Check(t, s.Login(pk11.NormalUser, ts.UserPin))

	for _, test := range tests {
		name := fmt.Sprintf("%v", test)
		t.Run(name, func(t *testing.T) {
			k, err := s.GenerateAES(test.keyBitLen, &pk11.KeyOptions{Extractable: true})
			if err != nil {
				return
			}

			t.Run("Seal", func(t *testing.T) {
				_, _, err := k.SealAESGCM(make([]byte, 12), nil, test.tagBits, []byte(Plain))
				if err == nil {
					t.Fatal("expected an error")
				}
			})

			t.Run("Unseal", func(t *testing.T) {
				_, err := k.UnsealAESGCM(make([]byte, 12), nil, test.tagBits, []byte(Plain))
				if err == nil {
					t.Fatal("expected an error")
				}
			})
		})
	}
}

func TestAESGCMEnc(t *testing.T) {
	tests := []uint{128, 256}

	s := ts.GetSession(t)
	ts.Check(t, s.Login(pk11.NormalUser, ts.UserPin))

	for _, test := range tests {
		name := fmt.Sprintf("%v", test)
		t.Run(name, func(t *testing.T) {
			k, err := s.GenerateAES(test, &pk11.KeyOptions{Extractable: true})
			ts.Check(t, err)

			kIface, err := k.ExportKey()
			ts.Check(t, err)
			kBytes := kIface.(pk11.AESKey)

			ciph, iv, err := k.SealAESGCM(make([]byte, 12), nil, 128, []byte(Plain))
			ts.Check(t, err)

			aes, err := aes.NewCipher([]byte(kBytes))
			ts.Check(t, err)
			aead, err := cipher.NewGCM(aes)
			ts.Check(t, err)

			plain, err := aead.Open(nil, iv, ciph, nil)
			ts.Check(t, err)

			if string(plain) != Plain {
				t.Fatal("Plaintext mismatch")
			}
		})
	}
}

func TestAESGCMDec(t *testing.T) {
	tests := []uint{128, 256}

	s := ts.GetSession(t)
	ts.Check(t, s.Login(pk11.NormalUser, ts.UserPin))

	for _, test := range tests {
		name := fmt.Sprintf("%v", test)
		t.Run(name, func(t *testing.T) {
			k, err := s.GenerateAES(test, &pk11.KeyOptions{Extractable: true})
			ts.Check(t, err)

			kIface, err := k.ExportKey()
			ts.Check(t, err)
			kBytes := kIface.(pk11.AESKey)

			aes, err := aes.NewCipher([]byte(kBytes))
			ts.Check(t, err)
			aead, err := cipher.NewGCM(aes)
			ts.Check(t, err)

			ciph := aead.Seal(nil, make([]byte, 12), []byte(Plain), nil)

			plain, err := k.UnsealAESGCM(make([]byte, 12), nil, 128, ciph)
			ts.Check(t, err)

			if string(plain) != Plain {
				t.Fatal("Plaintext mismatch")
			}
		})
	}
}

func TestAESGCMImport(t *testing.T) {
	tests := []uint{128, 256}

	s := ts.GetSession(t)
	ts.Check(t, s.Login(pk11.NormalUser, ts.UserPin))

	for _, test := range tests {
		name := fmt.Sprintf("%v", test)
		t.Run(name, func(t *testing.T) {
			// This does not need to be secure randomness.
			key := make([]byte, test/8)
			rand.Read(key)

			ki, err := s.ImportKey(pk11.AESKey(key), nil)
			ts.Check(t, err)
			ko := ki.(pk11.SecretKey)

			ciph, iv, err := ko.SealAESGCM(make([]byte, 12), nil, 128, []byte(Plain))
			ts.Check(t, err)

			aes, err := aes.NewCipher(key)
			ts.Check(t, err)
			aead, err := cipher.NewGCM(aes)
			ts.Check(t, err)

			plain, err := aead.Open(nil, iv, ciph, nil)
			ts.Check(t, err)

			if string(plain) != Plain {
				t.Fatal("Plaintext mismatch for HSM encrypt")
			}

			ciph = aead.Seal(nil, make([]byte, 12), []byte(Plain), nil)

			plain, err = ko.UnsealAESGCM(make([]byte, 12), nil, 128, ciph)
			ts.Check(t, err)

			if string(plain) != Plain {
				t.Fatal("Plaintext mismatch for HSM decrypt")
			}
		})
	}
}
