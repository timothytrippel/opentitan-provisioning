// Copyright lowRISC contributors (OpenTitan project).
// Licensed under the Apache License, Version 2.0, see LICENSE for details.
// SPDX-License-Identifier: Apache-2.0

// Package pk11 provides a wrapper over the "github.com/miekg/pkcs11" library
// that exposes a reasonably agnostic interface.
package pk11

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Dumps information about the module and tokens within.
func (m *Mod) Dump() string {
	s := new(strings.Builder)

	info, err := m.Raw().GetInfo()
	fmt.Fprint(s, "m.Info = ")
	if err != nil {
		fmt.Fprintln(s, err)
	} else {
		j, _ := json.MarshalIndent(info, "", "  ")
		fmt.Fprintln(s, string(j))
	}

	slots, err := m.Raw().GetSlotList( /*tokenPresent=*/ false)
	if err != nil {
		fmt.Fprintf(s, "m.Slots = %s\n", err)
	} else {
		for i, sl := range slots {
			fmt.Fprintf(s, "m.Slots[%d] = ", i)
			info, err := m.Raw().GetSlotInfo(sl)
			if err != nil {
				fmt.Fprintln(s, err)
			} else {
				j, _ := json.MarshalIndent(info, "", "  ")
				fmt.Fprintln(s, string(j))
			}

			fmt.Fprintf(s, "m.Slots[%d].Token = ", i)
			tinfo, err := m.Raw().GetTokenInfo(sl)
			if err != nil {
				fmt.Fprintln(s, err)
			} else {
				j, _ := json.MarshalIndent(tinfo, "", "  ")
				fmt.Fprintln(s, string(j))
			}

			mechs, err := m.Raw().GetMechanismList(sl)
			if err != nil {
				fmt.Fprintf(s, "m.Slots[%d].Mechs = %s\n", i, err)
			}
			for i2, mech := range mechs {
				fmt.Fprintf(s, "m.Slots[%d].Mechs[%d] = 0x%x\n", i, i2, mech.Mechanism)
				fmt.Fprintf(s, "m.Slots[%d].Mechs[%d].Info = ", i, i2)
				// GetMechanismInfo ignores all but the first slice element.
				minfo, err := m.Raw().GetMechanismInfo(sl, mechs[i2:])
				if err != nil {
					fmt.Fprintln(s, err)
				} else {
					j, _ := json.MarshalIndent(minfo, "", "  ")
					fmt.Fprintln(s, string(j))
				}
			}
		}
	}

	return s.String()
}
