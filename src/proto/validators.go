// Copyright lowRISC contributors (OpenTitan project).
// Licensed under the Apache License, Version 2.0, see LICENSE for details.
// SPDX-License-Identifier: Apache-2.0

// Package validators provides validation routines for OT provisioning proto validators.
//
// See:
//   - https://docs.google.com/document/d/1dE7vR791Atp7Wu7Ss90K1MvdyoroouSHPdq_RXQ2R8I#bookmark=id.n9feo7yvyhle
//     FIXME: Replace above with a pointer to markdown TBD.
//   - https://docs.opentitan.org/doc/security/specs/identities_and_root_keys#device-identifier
package validators

import (
	"fmt"

	dpb "github.com/lowRISC/opentitan-provisioning/src/proto/device_id_go_pb"
)

const (
	DeviceIdSkuSpecificLenInBytes  = 16
	MaxDeviceDataPayloadLenInBytes = 8192
)

// Checks that a uint32 fits into 16 bits.
func validate16Bits(val uint32) error {
	if val != uint32(uint16(val)) {
		return fmt.Errorf("Value outside 16-bit range: %v", val)
	}
	return nil
}

// Checks a SiliconCreatorId value for validity.
func ValidateSiliconCreatorId(sc dpb.SiliconCreatorId) error {
	if err := validate16Bits(uint32(sc)); err != nil {
		return err
	}
	switch sc {
	case dpb.SiliconCreatorId_SILICON_CREATOR_ID_OPENSOURCE:
		fallthrough
	case dpb.SiliconCreatorId_SILICON_CREATOR_ID_NUVOTON:
		return nil
	}
	return fmt.Errorf("Invalid SiliconCreatorId: %v", sc)
}

// Checks a ProductId value for validity.
func ValidateProductId(pi dpb.ProductId) error {
	if err := validate16Bits(uint32(pi)); err != nil {
		return err
	}
	switch pi {
	case dpb.ProductId_PRODUCT_ID_EARLGREY_Z1:
		fallthrough
	case dpb.ProductId_PRODUCT_ID_EARLGREY_A1:
		return nil
	}
	return fmt.Errorf("Invalid ProductId: %v", pi)
}

// Performs invariant checks for a HardwareOrigin that protobuf syntax cannot capture.
func ValidateHardwareOrigin(ho *dpb.HardwareOrigin) error {
	if err := ValidateSiliconCreatorId(ho.SiliconCreatorId); err != nil {
		return err
	}
	if err := ValidateProductId(ho.ProductId); err != nil {
		return err
	}
	// TODO: Validate ho.DeviceIdentificationNumber
	return nil
}

// Performs invariant checks for a DeviceId that protobuf syntax cannot capture.
func ValidateDeviceId(di *dpb.DeviceId) error {
	if err := ValidateHardwareOrigin(di.HardwareOrigin); err != nil {
		return err
	}

	// len(di.SkuSpecific) == 0 ==> (optional) field not supplied,
	// which is considered valid.
	if len(di.SkuSpecific) != 0 && len(di.SkuSpecific) != DeviceIdSkuSpecificLenInBytes {
		return fmt.Errorf("Invalid SkuSpecific string length: %v", len(di.SkuSpecific))
	}

	// TODO: Validate di.crc32
	return nil
}

// DeviceIdToString injectively converts a (valid!) DeviceId into a deterministic string.
func DeviceIdToString(di *dpb.DeviceId) string {
	return fmt.Sprintf("DeviceId:%x:%x:%x:%x",
		di.HardwareOrigin.SiliconCreatorId,
		di.HardwareOrigin.ProductId,
		di.HardwareOrigin.DeviceIdentificationNumber,
		di.SkuSpecific)
}

// Checks the length of the payload object ([]byte).  Since a payload is optional,
// 0-length is considered valid.
func validatePayload(payload []byte) error {
	l := len(payload)
	if l > MaxDeviceDataPayloadLenInBytes {
		return fmt.Errorf("Invalid Payload length: %v", l)
	}

	return nil
}

// ValidateDeviceLifeCycle checks a life cycle value for validity.
func ValidateDeviceLifeCycle(lc dpb.DeviceLifeCycle) error {
	switch lc {
	case
		dpb.DeviceLifeCycle_DEVICE_LIFE_CYCLE_RAW,
		dpb.DeviceLifeCycle_DEVICE_LIFE_CYCLE_TEST_LOCKED,
		dpb.DeviceLifeCycle_DEVICE_LIFE_CYCLE_TEST_UNLOCKED,
		dpb.DeviceLifeCycle_DEVICE_LIFE_CYCLE_DEV,
		dpb.DeviceLifeCycle_DEVICE_LIFE_CYCLE_PROD,
		dpb.DeviceLifeCycle_DEVICE_LIFE_CYCLE_PROD_END,
		dpb.DeviceLifeCycle_DEVICE_LIFE_CYCLE_RMA,
		dpb.DeviceLifeCycle_DEVICE_LIFE_CYCLE_SCRAP:
		return nil
	default:
		return fmt.Errorf("Invalid DeviceLifeCycle: %v", lc)
	}
}

// ValidateDeviceData performs invariant checks for a DeviceData that
// protobuf syntax cannot capture.
func ValidateDeviceData(dd *dpb.DeviceData) error {
	if err := validatePayload(dd.Payload); err != nil {
		return err
	}

	// TODO: Validate metadata.

	return ValidateDeviceLifeCycle(dd.DeviceLifeCycle)
}
