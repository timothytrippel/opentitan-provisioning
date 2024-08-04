// Copyright lowRISC contributors (OpenTitan project).
// Licensed under the Apache License, Version 2.0, see LICENSE for details.
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"testing"

	"github.com/lowRISC/ot-provisioning/src/pk11"
	ts "github.com/lowRISC/ot-provisioning/src/pk11/test_support"
)

func TestUserLogin(t *testing.T) {
	s := ts.GetSession(t)
	ts.Check(t, s.Login(pk11.NormalUser, ts.UserPin))
}

func TestSOLogin(t *testing.T) {
	s := ts.GetSession(t)
	ts.Check(t, s.Login(pk11.SecurityOfficerUser, ts.SecOffPin))
}
