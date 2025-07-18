// Copyright lowRISC contributors (OpenTitan project).
// Licensed under the Apache License, Version 2.0, see LICENSE for details.
// SPDX-License-Identifier: Apache-2.0
#include <gmock/gmock.h>
#include <google/protobuf/util/json_util.h>
#include <grpcpp/grpcpp.h>
#include <gtest/gtest.h>

#include <memory>
#include <string>

#include "absl/memory/memory.h"
#include "src/ate/ate_api.h"
#include "src/ate/proto/dut_commands.pb.h"
#include "src/pa/proto/pa.grpc.pb.h"
#include "src/pa/proto/pa_mock.grpc.pb.h"
#include "src/testing/test_helpers.h"

namespace {

using testing::EqualsProto;

class AteJsonTest : public ::testing::Test {};

TEST_F(AteJsonTest, TokensToJson) {
  dut_spi_frame_t frame;
  token_t wafer_auth_secret = {0};
  token_t test_unlock_token = {0};
  token_t test_exit_token = {0};

  wafer_auth_secret.size = sizeof(uint32_t) * 8;
  test_unlock_token.size = sizeof(uint64_t) * 2;
  test_exit_token.size = sizeof(uint64_t) * 2;

  wafer_auth_secret.data[0] = 1;
  test_unlock_token.data[0] = 1;
  test_exit_token.data[0] = 1;

  EXPECT_EQ(TokensToJson(&wafer_auth_secret, &test_unlock_token,
                         &test_exit_token, &frame),
            0);

  std::string json_string = std::string(reinterpret_cast<char*>(frame.payload),
                                        kDutRxSpiFrameSizeInBytes);

  ot::dut_commands::TokensJSON tokens_cmd;
  google::protobuf::util::JsonParseOptions options;
  options.ignore_unknown_fields = true;
  google::protobuf::util::Status status =
      google::protobuf::util::JsonStringToMessage(json_string, &tokens_cmd,
                                                  options);
  EXPECT_EQ(status.ok(), true);
  EXPECT_THAT(tokens_cmd, EqualsProto(R"pb(
                wafer_auth_secret: 1
                wafer_auth_secret: 0
                wafer_auth_secret: 0
                wafer_auth_secret: 0
                wafer_auth_secret: 0
                wafer_auth_secret: 0
                wafer_auth_secret: 0
                wafer_auth_secret: 0
                test_unlock_token_hash: 1
                test_unlock_token_hash: 0
                test_exit_token_hash: 1
                test_exit_token_hash: 0
              )pb"));
}

TEST_F(AteJsonTest, DeviceIdFromJson) {
  ot::dut_commands::DeviceIdJSON device_id_cmd;
  device_id_cmd.add_cp_device_id(0x12345678);
  device_id_cmd.add_cp_device_id(0x0);
  device_id_cmd.add_cp_device_id(0x0);
  device_id_cmd.add_cp_device_id(0x0);

  std::string command;
  google::protobuf::util::JsonOptions options;
  options.add_whitespace = false;
  options.always_print_primitive_fields = true;
  options.preserve_proto_field_names = true;
  google::protobuf::util::Status status =
      google::protobuf::util::MessageToJsonString(device_id_cmd, &command,
                                                  options);
  EXPECT_EQ(status.ok(), true);

  dut_spi_frame_t frame = {0};
  memcpy(frame.payload, command.data(), command.size());
  frame.size = command.size();

  device_id_bytes_t device_id = {0};
  EXPECT_EQ(DeviceIdFromJson(&frame, &device_id), 0);
  EXPECT_THAT(
      device_id.raw,
      testing::ElementsAreArray(
          {0x78, 0x56, 0x34, 0x12, 0x0,  0x0,  0x0,  0x0,  0x00, 0x00, 0x00,
           0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
           0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}));
}

TEST_F(AteJsonTest, RmaTokenWithoutCrc) {
  token_t rma_token = {0};
  rma_token.size = sizeof(uint64_t) * 2;
  rma_token.data[0] = 0x11;
  rma_token.data[1] = 0x22;

  dut_spi_frame_t ate_to_dut_frame;
  EXPECT_EQ(RmaTokenToJson(&rma_token, &ate_to_dut_frame, /*skip_crc=*/true),
            0);

  std::string json_string =
      std::string(reinterpret_cast<char*>(ate_to_dut_frame.payload),
                  kDutRxSpiFrameSizeInBytes);

  // Use the proto representation of RmaTokenJSON to verify the JSON string.
  ot::dut_commands::RmaTokenJSON rma_hash_cmd;
  google::protobuf::util::JsonParseOptions options;
  options.ignore_unknown_fields = true;
  google::protobuf::util::Status status =
      google::protobuf::util::JsonStringToMessage(json_string, &rma_hash_cmd,
                                                  options);
  EXPECT_EQ(status.ok(), true);
  EXPECT_THAT(rma_hash_cmd, EqualsProto(R"pb(
                hash: 8721 hash: 0
              )pb"));

  dut_spi_frame_t dut_to_ate_frame = {.payload = {0},
                                      .size = kDutTxMaxSpiFrameSizeInBytes};
  memcpy(dut_to_ate_frame.payload, ate_to_dut_frame.payload,
         kDutRxSpiFrameSizeInBytes);
  token_t rma_token_got = {0};
  EXPECT_EQ(RmaTokenFromJson(&dut_to_ate_frame, &rma_token_got), 0);
  EXPECT_THAT(rma_token_got.data, testing::ElementsAreArray(
                                      rma_token.data, sizeof(rma_token.data)));
  EXPECT_EQ(rma_token_got.size, sizeof(uint64_t) * 2);
}

TEST_F(AteJsonTest, RmaTokenWithCrc) {
  token_t rma_token = {0};
  rma_token.size = sizeof(uint64_t) * 2;
  rma_token.data[0] = 0x11;
  rma_token.data[1] = 0x22;

  dut_spi_frame_t frame_with_crc;
  EXPECT_EQ(RmaTokenToJson(&rma_token, &frame_with_crc, /*skip_crc=*/false), 0);
  dut_spi_frame_t frame_without_crc;
  EXPECT_EQ(RmaTokenToJson(&rma_token, &frame_without_crc, /*skip_crc=*/true),
            0);

  std::string json_string_without_crc =
      std::string(reinterpret_cast<char*>(frame_without_crc.payload),
                  kDutRxSpiFrameSizeInBytes);
  std::string json_string_with_crc =
      std::string(reinterpret_cast<char*>(frame_with_crc.payload),
                  kDutRxSpiFrameSizeInBytes);

  // Use the proto representation of RmaTokenJSON to verify the JSON string.
  ot::dut_commands::RmaTokenJSON rma_hash_cmd;
  google::protobuf::util::JsonParseOptions options;
  options.ignore_unknown_fields = true;
  google::protobuf::util::Status status =
      google::protobuf::util::JsonStringToMessage(json_string_without_crc,
                                                  &rma_hash_cmd, options);
  EXPECT_EQ(status.ok(), true);
  EXPECT_THAT(rma_hash_cmd, EqualsProto(R"pb(
                hash: 8721 hash: 0
              )pb"));

  dut_spi_frame_t dut_to_ate_frame_with_crc = {
      .payload = {0}, .size = kDutTxMaxSpiFrameSizeInBytes};
  memcpy(dut_to_ate_frame_with_crc.payload, frame_with_crc.payload,
         kDutRxSpiFrameSizeInBytes);
  token_t rma_token_got = {0};
  EXPECT_EQ(RmaTokenFromJson(&dut_to_ate_frame_with_crc, &rma_token_got), 0);
  EXPECT_THAT(rma_token_got.data, testing::ElementsAreArray(
                                      rma_token.data, sizeof(rma_token.data)));
  EXPECT_EQ(rma_token_got.size, sizeof(uint64_t) * 2);
}

TEST_F(AteJsonTest, CaSubjectKeys) {
  ca_subject_key_t dice_ca_key_id = {0};
  ca_subject_key_t aux_ca_key_id = {0};
  dice_ca_key_id.data[0] = 65;
  dice_ca_key_id.data[9] = 12;
  aux_ca_key_id.data[0] = 123;
  aux_ca_key_id.data[19] = 255;

  dut_spi_frame_t frame;
  EXPECT_EQ(CaSubjectKeysToJson(&dice_ca_key_id, &aux_ca_key_id, &frame), 0);

  std::string json_string = std::string(reinterpret_cast<char*>(frame.payload),
                                        kDutRxSpiFrameSizeInBytes);

  // Use the proto representation of CaSubjectKeysJSON to verify the JSON
  // string.
  ot::dut_commands::CaSubjectKeysJSON ca_key_ids_cmd;
  google::protobuf::util::JsonParseOptions options;
  options.ignore_unknown_fields = true;
  google::protobuf::util::Status status =
      google::protobuf::util::JsonStringToMessage(json_string, &ca_key_ids_cmd,
                                                  options);
  EXPECT_EQ(status.ok(), true);
  EXPECT_THAT(ca_key_ids_cmd, EqualsProto(R"pb(
                dice_auth_key_key_id: 65
                dice_auth_key_key_id: 0
                dice_auth_key_key_id: 0
                dice_auth_key_key_id: 0
                dice_auth_key_key_id: 0
                dice_auth_key_key_id: 0
                dice_auth_key_key_id: 0
                dice_auth_key_key_id: 0
                dice_auth_key_key_id: 0
                dice_auth_key_key_id: 12
                dice_auth_key_key_id: 0
                dice_auth_key_key_id: 0
                dice_auth_key_key_id: 0
                dice_auth_key_key_id: 0
                dice_auth_key_key_id: 0
                dice_auth_key_key_id: 0
                dice_auth_key_key_id: 0
                dice_auth_key_key_id: 0
                dice_auth_key_key_id: 0
                dice_auth_key_key_id: 0
                ext_auth_key_key_id: 123
                ext_auth_key_key_id: 0
                ext_auth_key_key_id: 0
                ext_auth_key_key_id: 0
                ext_auth_key_key_id: 0
                ext_auth_key_key_id: 0
                ext_auth_key_key_id: 0
                ext_auth_key_key_id: 0
                ext_auth_key_key_id: 0
                ext_auth_key_key_id: 0
                ext_auth_key_key_id: 0
                ext_auth_key_key_id: 0
                ext_auth_key_key_id: 0
                ext_auth_key_key_id: 0
                ext_auth_key_key_id: 0
                ext_auth_key_key_id: 0
                ext_auth_key_key_id: 0
                ext_auth_key_key_id: 0
                ext_auth_key_key_id: 0
                ext_auth_key_key_id: 255
              )pb"));
}

TEST_F(AteJsonTest, PersoBlob) {
  perso_blob_t blob = {0};
  blob.num_objects = 1;

  // Fill the blob with random data for testing.
  for (size_t i = 0; i < sizeof(blob.body); ++i) {
    blob.body[i] = static_cast<uint8_t>((i | 0x80) & 0xFF);
  }
  blob.next_free = sizeof(blob.body);

  constexpr size_t kNum256ByteFrames = 150;
  dut_spi_frame_t ate_to_dut_frames[kNum256ByteFrames] = {0};
  size_t num_frames = kNum256ByteFrames;
  EXPECT_EQ(PersoBlobToJson(&blob, ate_to_dut_frames, &num_frames), 0);
  EXPECT_EQ(num_frames, 129);

  // Translate the RX buffer into a TX buffer.
  const size_t kNum2020ByteFrames =
      ((kNum256ByteFrames * kDutRxSpiFrameSizeInBytes) + 2020 - 1) / 2020;
  dut_spi_frame_t dut_to_ate_frames[kNum2020ByteFrames] = {0};
  uint8_t tmp[kNum2020ByteFrames * 2020] = {0};
  memset(tmp, sizeof(tmp), ' ');
  for (size_t i = 0; i < kNum256ByteFrames; ++i) {
    memcpy(&tmp[i * kDutRxSpiFrameSizeInBytes], ate_to_dut_frames[i].payload,
           kDutRxSpiFrameSizeInBytes);
    ate_to_dut_frames[i].size = kDutRxSpiFrameSizeInBytes;
  }
  for (size_t i = 0; i < kNum2020ByteFrames; ++i) {
    memcpy(dut_to_ate_frames[i].payload, &tmp[i * 2020], 2020);
    dut_to_ate_frames[i].size = 2020;
  }

  perso_blob_t blob_got = {0};
  EXPECT_EQ(PersoBlobFromJson(dut_to_ate_frames, kNum2020ByteFrames, &blob_got),
            0);
  EXPECT_EQ(blob_got.num_objects, 1);
  EXPECT_EQ(blob_got.next_free, 8192);
  EXPECT_THAT(blob_got.body,
              testing::ElementsAreArray(blob.body, sizeof(blob.body)));
}

}  // namespace
