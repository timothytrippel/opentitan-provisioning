// Copyright lowRISC contributors (OpenTitan project).
// Licensed under the Apache License, Version 2.0, see LICENSE for details.
// SPDX-License-Identifier: Apache-2.0

// Package device_data contains device variables.
package device_testdata

import (
	dpb "github.com/lowRISC/ot-provisioning/src/proto/device_id_go_pb"
)

const (
	DeviceIdSkuSpecificLen = 128
)

var (
	hwOriginOk = dpb.HardwareOrigin{
		DeviceType: &dpb.DeviceType{
			SiliconCreator:    dpb.SiliconCreator_SILICON_CREATOR_TEST,
			ProductIdentifier: 0,
		},
		DeviceIdentificationNumber: 0,
	}
	hwOriginBadCreator = dpb.HardwareOrigin{
		DeviceType: &dpb.DeviceType{
			SiliconCreator:    dpb.SiliconCreator_SILICON_CREATOR_UNSPECIFIED,
			ProductIdentifier: 0,
		},
		DeviceIdentificationNumber: 0,
	}
)

func NewDeviceID() *dpb.DeviceId {
	return &dpb.DeviceId{
		HardwareOrigin: &hwOriginOk,
		SkuSpecific:    make([]byte, DeviceIdSkuSpecificLen),
	}
}

func NewDeviceIDSkuTooLong() *dpb.DeviceId {
	return &dpb.DeviceId{
		HardwareOrigin: &hwOriginOk,
		SkuSpecific:    make([]byte, DeviceIdSkuSpecificLen+1),
	}
}

func NewDeviceIDMissingSku() *dpb.DeviceId {
	return &dpb.DeviceId{
		HardwareOrigin: &hwOriginOk,
		SkuSpecific:    nil, // Empty SkuSpecific is OK.
	}
}

func NewDeviceIdBadOrigin() *dpb.DeviceId {
	return &dpb.DeviceId{
		HardwareOrigin: &hwOriginBadCreator,
		SkuSpecific:    make([]byte, DeviceIdSkuSpecificLen),
	}
}
