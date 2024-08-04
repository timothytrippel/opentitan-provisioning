#!/bin/bash
# Copyright lowRISC contributors (OpenTitan project).
# Licensed under the Apache License, Version 2.0, see LICENSE for details.
# SPDX-License-Identifier: Apache-2.0
set -e

# Script used to generate endpoint credentials. This functionality will
# be implemented later in a Go utility to be able to leverage the HSM
# interface.
# See docs/spm.md for more details on how use the keys and certificates
# produced by this script.

readonly DEV_CONFIG_PATH=config/dev/certs/out
readonly DEV_CONFIG_TEMPLATE_PATH=config/dev/certs/templates
readonly CA_KEY=${DEV_CONFIG_PATH}/ca-key.pem
readonly CA_CERT=${DEV_CONFIG_PATH}/ca-cert.pem

mkdir -p ${DEV_CONFIG_PATH}

echo "Creating CA certificate"
openssl req -x509 -newkey rsa:4096 -days 365 -nodes \
    -keyout ${CA_KEY} \
    -out ${CA_CERT} \
    -config ${DEV_CONFIG_TEMPLATE_PATH}/ca.cnf


readonly SERVICE_CERT_REQ=${DEV_CONFIG_PATH}/pa-req.pem

create_key_and_cert () {
  ENDPOINT_CERT_REQ=${DEV_CONFIG_PATH}/${1}-req.pem
  echo "Creating ${1} private key and certificate signing request"
  openssl req -newkey rsa:4096 -nodes \
      -keyout ${DEV_CONFIG_PATH}/${1}-key.pem \
      -out ${ENDPOINT_CERT_REQ} \
      -config ${DEV_CONFIG_TEMPLATE_PATH}/endpoint_${1}.cnf

  echo "Signing ${1} certificate with CA key."
  openssl x509 -req \
    -in ${ENDPOINT_CERT_REQ} \
    -days 60 \
    -CA ${CA_CERT} \
    -CAkey ${CA_KEY} \
    -CAcreateserial \
    -out ${DEV_CONFIG_PATH}/${1}-cert.pem \
    -extensions req_ext \
    -extfile ${DEV_CONFIG_TEMPLATE_PATH}/endpoint_${1}.cnf

  rm ${ENDPOINT_CERT_REQ}
}

create_key_and_cert "pa-service"
create_key_and_cert "spm-service"
create_key_and_cert "ate-client"
