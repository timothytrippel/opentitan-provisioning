// Copyright lowRISC contributors (OpenTitan project).
// Licensed under the Apache License, Version 2.0, see LICENSE for details.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/lowRISC/opentitan-provisioning/src/ate"

	dipb "github.com/lowRISC/opentitan-provisioning/src/proto/device_id_go_pb"
	rrpb "github.com/lowRISC/opentitan-provisioning/src/proto/registry_record_go_pb"
	proxybufferpb "github.com/lowRISC/opentitan-provisioning/src/proxy_buffer/proto/proxy_buffer_go_pb"
)

const LineLimit = 100

func printPersoBlob(persoBlobBytes []byte) {
	_, err := ate.UnpackPersoBlob(persoBlobBytes)
	if err != nil {
		fmt.Println("Error parsing perso blob:", err)
		return
	}

	// TODO(moidx): print X509 + CWT certs to console here.
	// TODO(moidx): print X509 certs to a file.
}

func printRegistryRecord(rr *rrpb.RegistryRecord) {
	// Parse device data from from the registry record.
	deviceData := &dipb.DeviceData{}
	proto.Unmarshal(rr.Data, deviceData)

	// Print each field of the registry record.
	fmt.Println(strings.Repeat("-", LineLimit))
	fmt.Println("Registry Record: ")
	fmt.Println(strings.Repeat("-", LineLimit))
	fmt.Printf("SKU:        %s\n", rr.Sku)
	fmt.Printf("Version:    %d\n", rr.Version)
	fmt.Printf("Device ID:  %s\n", rr.DeviceId)
	fmt.Println(strings.Repeat("-", LineLimit))
	fmt.Printf("Perso Firmware Hash:  %032x\n", deviceData.PersoFwSha256Hash)
	fmt.Printf("LC State:             %s\n", deviceData.DeviceLifeCycle)
	fmt.Printf("Wrapped RMA Token:    %x\n", deviceData.WrappedRmaUnlockToken)
	fmt.Println(strings.Repeat("-", LineLimit))
	fmt.Printf("Num Perso TLV Objects:  %d\n", deviceData.NumPersoTlvObjects)
	printPersoBlob(deviceData.PersoTlvData)
	fmt.Println(strings.Repeat("-", LineLimit))
	fmt.Println("Record endorsement (decode at: https://lapo.it/asn1js/): ")
	fmt.Printf("AuthPubkey    (ASN.1 DER): %x\n", rr.AuthPubkey)
	fmt.Printf("AuthSignature (ASN.1 DER): %x\n", rr.AuthSignature)
	fmt.Println(strings.Repeat("-", LineLimit))
}

func main() {
	// Check if a file path was provided
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run rr_parser.go <JSON registry record>")
		return
	}
	rrJSONPath := os.Args[1]

	// Open the registry record JSON file.
	rrFile, err := os.Open(rrJSONPath)
	if err != nil {
		fmt.Printf("Error opening file %s: %v\n", rrJSONPath, err)
		return
	}
	defer rrFile.Close()

	// Read the file contents of the registry record.
	rrBytes, err := io.ReadAll(rrFile)
	if err != nil {
		fmt.Println("Error reading registry record file:", err)
		return
	}

	// Parse the registry record into a struct.
	var pbRegistrationRequest proxybufferpb.DeviceRegistrationRequest
	decode_err := json.Unmarshal(rrBytes, &pbRegistrationRequest)
	if decode_err != nil {
		fmt.Println("Error unmarshaling registry record JSON:", decode_err)
		return
	}
	rr := pbRegistrationRequest.Record

	// Print registry record contents to console and save cert chains to a file.
	printRegistryRecord(rr)
	// TODO(moidx): verify cert chains files.
}
