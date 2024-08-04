// Copyright lowRISC contributors (OpenTitan project).
// Licensed under the Apache License, Version 2.0, see LICENSE for details.
// SPDX-License-Identifier: Apache-2.0

// Package tpm implements a TPM certificate template.
package tpm

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"

	"github.com/lowRISC/ot-provisioning/src/cert/signer"
)

type builder struct{}

// New creates a new instance of the tpm certificate template builder.
func New() signer.Template {
	return new(builder)
}

// Build creates the tpm certificate template.
func (b *builder) Build(p *signer.Params) (*x509.Certificate, error) {
	serialNumber := big.NewInt(0)
	serialNumber.SetBytes(p.SerialNumber)

	return &x509.Certificate{
		SerialNumber:       serialNumber,
		NotBefore:          p.NotBefore,
		NotAfter:           p.NotAfter,
		Subject:            p.Subject,
		UnknownExtKeyUsage: p.ExtKeyUsage,
		Issuer:             p.Issuer,

		// Basic constraints with extension id: 2.5.29.19
		BasicConstraintsValid: p.BasicConstraintsValid,
		IsCA:                  p.IsCA,
		MaxPathLenZero:        false,
		KeyUsage:              p.KeyUsage,
		IssuingCertificateURL: p.IssuingCertificateURL,
		ExtraExtensions: []pkix.Extension{
			p.SubjectAltName,
		},
	}, nil
}
