// Copyright lowRISC contributors (OpenTitan project).
// Licensed under the Apache License, Version 2.0, see LICENSE for details.
// SPDX-License-Identifier: Apache-2.0

package signer

import (
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"time"
)

// Params contains parameters used to populate the certificate template at
// build time.
type Params struct {
	Version                           int
	SerialNumber                      []byte
	Issuer, Subject, BasicConstraints pkix.Name
	NotBefore, NotAfter               time.Time
	KeyUsage                          x509.KeyUsage
	ExtKeyUsage                       []asn1.ObjectIdentifier
	BasicConstraintsValid             bool
	IsCA                              bool
	SignatureAlgorithm                x509.SignatureAlgorithm
	Extension                         []pkix.Extension
	AuthorityKeyId                    pkix.Extension
	SubjectAltName                    pkix.Extension
	IssuingCertificateURL             []string
}

// Template defines a certificate build interface.
type Template interface {
	Build(*Params) (*x509.Certificate, error)
}

// CreateCertificate creates a certificate from an x509 template endorsing the
// provided pub key, with a signature generated using priv key. The provided
// parent certificate must endorse the public version of priv key.
//
// The priv key must implement the crypto.Signer interface.
func CreateCertificate(template, parent *x509.Certificate, pub, priv any) ([]byte, error) {
	cert, err := x509.CreateCertificate(rand.Reader, template, parent, pub, priv)
	if err != nil {
		return nil, err
	}
	return cert, nil
}
