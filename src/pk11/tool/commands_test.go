// Copyright lowRISC contributors (OpenTitan project).
// Licensed under the Apache License, Version 2.0, see LICENSE for details.
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"bytes"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"math/big"
	"math/rand"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/bazelbuild/rules_go/go/tools/bazel"
	"github.com/lowRISC/ot-provisioning/src/pk11"
	ts "github.com/lowRISC/ot-provisioning/src/pk11/test_support"
	"github.com/lowRISC/ot-provisioning/src/pk11/tool/lex"
)

func newFile(t *testing.T, contents []byte) string {
	t.Helper()
	f, err := os.CreateTemp(bazel.TestTmpDir(), "pk11-tool-*")
	if err != nil {
		t.Fatalf("could not create tmp file: %s", err)
	}
	defer f.Close()

	for len(contents) != 0 {
		len, err := f.Write(contents)
		if err != nil {
			t.Fatalf("could not create tmp file: %s", err)
		}
		contents = contents[len:]
	}

	return f.Name()
}

var states sync.Map

func getState(t *testing.T) *State {
	t.Helper()

	// Pull out the first component of a test name, since REPL states are
	// per-test, not per-subtest.
	name := t.Name()
	if idx := strings.Index(name, "/"); idx >= 0 {
		name = name[:idx]
	}

	if state, ok := states.Load(name); ok {
		return state.(*State)
	}

	state := fromMod(ts.GetMod())
	states.Store(name, state)
	return state
}

func openSession(t *testing.T) {
	t.Helper()
	_ = getState(t)
	run(t, `open-session %d`, ts.GetSlot(t))
}

func getVar(t *testing.T, name string) any {
	t.Helper()

	v, ok := getState(t).vars[name]
	if !ok {
		t.Fatalf("could not load var %q", name)
	}
	return v
}

func run(t *testing.T, cmd string, args ...any) any {
	t.Helper()

	v, err := tryRun(t, cmd, args...)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func tryRun(t *testing.T, cmd string, args ...any) (v any, err error) {
	t.Helper()

	line := fmt.Sprintf(cmd, args...)
	lexer := lex.New(strings.NewReader(line))
	state := getState(t)
	var eof bool
	for !eof {
		toks, ret, errs := state.Interpret(lexer)
		t.Log("pk11>", lex.StringTokens(toks[:len(toks)-1]))
		if len(toks) > 0 {
			_, eof = toks[len(toks)-1].Value.(lex.EOF)
		}

		for i, e := range errs {
			if i == len(errs)-1 {
				if err != nil {
					t.Error(err)
				}
				err = e
			} else {
				t.Error(err)
			}
		}
		v = ret
	}
	return
}

func TestRead(t *testing.T) {
	path := newFile(t, []byte("this is my file!"))
	text := run(t, `set my_var read %q`, path).([]byte)
	if string(text) != "this is my file!" {
		t.Fatalf("got back wrong file contents: %q", text)
	}
}

func TestWrite(t *testing.T) {
	path := newFile(t, nil)
	run(t, `string my_var "this is my other file!"; write $my_var %q`, path)

	text, err := os.ReadFile(path)
	ts.Check(t, err)
	if string(text) != "this is my other file!" {
		t.Fatalf("got back wrong file contents: %q", text)
	}
}

func TestLogin(t *testing.T) {
	openSession(t)
	run(t, `login %q`, ts.UserPin)
}

func TestFind(t *testing.T) {
	openSession(t)
	run(t, `login %q`, ts.UserPin)

	obj := run(t, `set obj gen-aes 128 n`).(pk11.Object)
	uid, err := obj.UID()
	ts.Check(t, err)

	uid2 := run(t, `uid $obj`).([]byte)
	if !bytes.Equal(uid, uid2) {
		t.Errorf("uid (via uid $obj) mismatch: %x != %x", uid, uid2)
	}

	uid2, err = run(t, `object sym $obj`).(pk11.Object).UID()
	ts.Check(t, err)
	if !bytes.Equal(uid, uid2) {
		t.Errorf("uid (via object sym $obj) mismatch: %x != %x", uid, uid2)
	}

	uid2, err = run(t, `object sym %q`, uid).(pk11.Object).UID()
	ts.Check(t, err)
	if !bytes.Equal(uid, uid2) {
		t.Errorf("uid (via object sym \"...\") mismatch: %x != %x", uid, uid2)
	}
}

func TestRSAKeygen(t *testing.T) {
	tests := []struct{ m, e uint }{
		{3072, 0x10001},
		{4096, 0x10001},
	}
	hashes := []struct {
		name string
		hash crypto.Hash
	}{
		{"sha-256", crypto.SHA256},
		{"sha-384", crypto.SHA384},
		{"sha-512", crypto.SHA512},
	}

	openSession(t)
	run(t, `login %q`, ts.UserPin)

	for _, tt := range tests {
		name := fmt.Sprint(tt)
		t.Run(name, func(t *testing.T) {
			export := run(t, `set key gen-rsa %d %d; export $key`, tt.m, tt.e).([]byte)

			ki, err := x509.ParsePKIXPublicKey(export)
			ts.Check(t, err)
			key := ki.(*rsa.PublicKey)

			run(t, `set skey object priv $key`)
			for _, h := range hashes {
				t.Run(h.name, func(t *testing.T) {
					hash := ts.MakeHash(h.hash, []byte(name))
					t.Run("pkcs1", func(t *testing.T) {
						sig := run(t, `sign-rsa-pkcs1 %s $skey %q`, h.name, name).([]byte)
						if err := rsa.VerifyPKCS1v15(key, h.hash, hash, sig); err != nil {
							t.Fatalf("could not verify signature: %s", err)
						}
					})

					t.Run("pss", func(t *testing.T) {
						sig := run(t, `sign-rsa-pss 12 %s $skey %q`, h.name, name).([]byte)
						if err := rsa.VerifyPSS(key, h.hash, hash, sig, &rsa.PSSOptions{SaltLength: 12}); err != nil {
							t.Fatalf("could not verify signature: %s", err)
						}
					})
				})
			}
		})
	}
}

func TestRSAImport(t *testing.T) {
	tests := []struct{ m uint }{{3072}, {4096}}
	hashes := []struct {
		name string
		hash crypto.Hash
	}{
		{"sha-256", crypto.SHA256},
		{"sha-384", crypto.SHA384},
		{"sha-512", crypto.SHA512},
	}

	openSession(t)
	run(t, `login %q`, ts.UserPin)

	for _, tt := range tests {
		name := fmt.Sprint(tt)
		t.Run(name, func(t *testing.T) {
			// This does not need to be secure randomness.
			rand := rand.New(rand.NewSource(0))

			kp, err := rsa.GenerateKey(rand, int(tt.m))
			ts.Check(t, err)
			key := &kp.PublicKey

			pk1 := x509.MarshalPKCS1PrivateKey(kp)
			keyFile := newFile(t, pk1)
			run(t, `set key-file read %q; set skey import-rsa $key-file yes`, keyFile)

			for _, h := range hashes {
				t.Run(h.name, func(t *testing.T) {
					hash := ts.MakeHash(h.hash, []byte(name))
					t.Run("pkcs1", func(t *testing.T) {
						sig := run(t, `sign-rsa-pkcs1 %s $skey %q`, h.name, name).([]byte)
						if err := rsa.VerifyPKCS1v15(key, h.hash, hash, sig); err != nil {
							t.Fatalf("could not verify signature: %s", err)
						}
					})

					t.Run("pss", func(t *testing.T) {
						sig := run(t, `sign-rsa-pss 12 %s $skey %q`, h.name, name).([]byte)
						if err := rsa.VerifyPSS(key, h.hash, hash, sig, &rsa.PSSOptions{SaltLength: 12}); err != nil {
							t.Fatalf("could not verify signature: %s", err)
						}
					})
				})
			}
		})
	}
}

func TestECDSAKeygen(t *testing.T) {
	curves := []struct {
		name string
		elliptic.Curve
	}{
		{"p-256", elliptic.P256()},
	}
	hashes := []struct {
		name string
		hash crypto.Hash
	}{
		{"sha-256", crypto.SHA256},
		{"sha-384", crypto.SHA384},
		{"sha-512", crypto.SHA512},
	}

	verifyScalars := func(pub *ecdsa.PublicKey, hash, sig []byte) bool {
		var r, s big.Int
		r.SetBytes(sig[:len(sig)/2])
		s.SetBytes(sig[len(sig)/2:])
		return ecdsa.Verify(pub, hash, &r, &s)
	}

	formats := []struct {
		format string
		verify func(pub *ecdsa.PublicKey, hash, sig []byte) bool
	}{
		{"", verifyScalars},
		{"fixed", verifyScalars},
		{"asn1", ecdsa.VerifyASN1},
	}

	openSession(t)
	run(t, `login %q`, ts.UserPin)

	for _, c := range curves {
		t.Run(c.name, func(t *testing.T) {
			export := run(t, `set key gen-ecdsa %s; export $key`, c.name).([]byte)

			ki, err := x509.ParsePKIXPublicKey(export)
			ts.Check(t, err)
			key := ki.(*ecdsa.PublicKey)

			if key.Curve != c.Curve {
				t.Fatalf("curve mismatch: want %s, got %s", c.Curve.Params().Name, key.Curve.Params().Name)
			}

			run(t, `set skey object priv $key`)
			for _, h := range hashes {
				t.Run(h.name, func(t *testing.T) {
					hash := ts.MakeHash(h.hash, []byte(c.name))
					for _, f := range formats {
						t.Run(fmt.Sprint(f.format), func(t *testing.T) {
							sig := run(t, `sign-ecdsa %s $skey %q %s`, h.name, c.name, f.format).([]byte)
							if !f.verify(key, hash, sig) {
								t.Fatal("could not verify signature")
							}
						})
					}
				})
			}
		})
	}
}

func TestECDSAImport(t *testing.T) {
	curves := []struct {
		name string
		elliptic.Curve
	}{
		{"p-256", elliptic.P256()},
	}
	hashes := []struct {
		name string
		hash crypto.Hash
	}{
		{"sha-256", crypto.SHA256},
		{"sha-384", crypto.SHA384},
		{"sha-512", crypto.SHA512},
	}

	verifyScalars := func(pub *ecdsa.PublicKey, hash, sig []byte) bool {
		var r, s big.Int
		r.SetBytes(sig[:len(sig)/2])
		s.SetBytes(sig[len(sig)/2:])
		return ecdsa.Verify(pub, hash, &r, &s)
	}

	formats := []struct {
		format string
		verify func(pub *ecdsa.PublicKey, hash, sig []byte) bool
	}{
		{"", verifyScalars},
		{"fixed", verifyScalars},
		{"asn1", ecdsa.VerifyASN1},
	}

	openSession(t)
	run(t, `login %q`, ts.UserPin)

	for _, c := range curves {
		t.Run(c.name, func(t *testing.T) {
			// This does not need to be secure randomness.
			rand := rand.New(rand.NewSource(0))

			kp, err := ecdsa.GenerateKey(c.Curve, rand)
			ts.Check(t, err)
			key := &kp.PublicKey

			pk1, err := x509.MarshalECPrivateKey(kp)
			ts.Check(t, err)
			keyFile := newFile(t, pk1)

			run(t, `set key-file read %q; set skey import-ecdsa $key-file yes`, keyFile)

			for _, h := range hashes {
				t.Run(h.name, func(t *testing.T) {
					hash := ts.MakeHash(h.hash, []byte(c.name))
					for _, f := range formats {
						t.Run(fmt.Sprint(f.format), func(t *testing.T) {
							sig := run(t, `sign-ecdsa %s $skey %q %s`, h.name, c.name, f.format).([]byte)
							if !f.verify(key, hash, sig) {
								t.Fatal("could not verify signature")
							}
						})
					}
				})
			}
		})
	}
}

const (
	iv  = "it's unique!"
	aad = "some extra stuff"
)

func TestAESKeygen(t *testing.T) {
	tests := []struct{ bits uint }{{128}, {256}}

	openSession(t)
	run(t, `login %q`, ts.UserPin)

	for _, tt := range tests {
		name := fmt.Sprint(tt)
		t.Run(name, func(t *testing.T) {
			key := run(t, `set key gen-aes %d n; export $key`, tt.bits).([]byte)

			aes, err := aes.NewCipher(key)
			ts.Check(t, err)
			gcm, err := cipher.NewGCM(aes)
			ts.Check(t, err)

			t.Run("seal", func(t *testing.T) {
				aead := run(t, `seal-aes-gcm $key %q %q 128 %q`, iv, aad, name).([]byte)

				msg, err := gcm.Open(nil, []byte(iv), aead, []byte(aad))
				ts.Check(t, err)

				if string(msg) != name {
					t.Fatal("decryption failure")
				}
			})

			t.Run("unseal", func(t *testing.T) {
				aead := gcm.Seal(nil, []byte(iv), []byte(name), []byte(aad))

				path := newFile(t, aead)
				msg := run(t, `set aead read %q; unseal-aes-gcm $key %q %q 128 $aead`, path, iv, aad).([]byte)

				if string(msg) != name {
					t.Fatal("decryption failure")
				}
			})
		})
	}
}

func TestAESImport(t *testing.T) {
	tests := []struct{ bits uint }{{128}, {256}}

	openSession(t)
	run(t, `login %q`, ts.UserPin)

	for _, tt := range tests {
		name := fmt.Sprint(tt)
		t.Run(name, func(t *testing.T) {
			key := make([]byte, tt.bits/8)
			for i := range key {
				key[i] = byte(i)
			}

			keyFile := newFile(t, key)
			run(t, `set key-file read %q; set key import-aes-raw $key-file yes`, keyFile)
			aes, err := aes.NewCipher(key)
			ts.Check(t, err)
			gcm, err := cipher.NewGCM(aes)
			ts.Check(t, err)

			t.Run("seal", func(t *testing.T) {
				aead := run(t, `seal-aes-gcm $key %q %q 128 %q`, iv, aad, name).([]byte)

				msg, err := gcm.Open(nil, []byte(iv), aead, []byte(aad))
				ts.Check(t, err)

				if string(msg) != name {
					t.Fatal("decryption failure")
				}
			})

			t.Run("unseal", func(t *testing.T) {
				aead := gcm.Seal(nil, []byte(iv), []byte(name), []byte(aad))

				path := newFile(t, aead)
				msg := run(t, `set aead read %q; unseal-aes-gcm $key %q %q 128 $aead`, path, iv, aad).([]byte)

				if string(msg) != name {
					t.Fatal("decryption failure")
				}
			})
		})
	}
}
