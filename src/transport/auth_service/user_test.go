// Copyright lowRISC contributors (OpenTitan project).
// Licensed under the Apache License, Version 2.0, see LICENSE for details.
// SPDX-License-Identifier: Apache-2.0

// Package auth_service implements the user object

package auth_service

import (
	"reflect"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNewUserObject(t *testing.T) {
	type args struct {
		userID      string
		token       string
		sku         string
		authMethods []string
	}
	tests := []struct {
		name    string
		expCode codes.Code
		args    *args
		want    *User
	}{
		{
			// This is a simple user test. The fields
			// values should be updated if there is additional
			// logic added to the auth service.
			name:    "new user test",
			expCode: codes.OK,
			args: &args{
				userID:      "fakeUser",
				sku:         "1234",
				token:       "fakeToken",
				authMethods: []string{"fakeMethod"},
			},
			want: &User{
				userID:       "fakeUser",
				sku:          "1234",
				sessionToken: "fakeToken",
				authMethods:  []string{"fakeMethod"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewUserObject(tt.args.userID, tt.args.token, tt.args.sku, tt.args.authMethods)
			s, ok := status.FromError(err)
			if !ok {
				t.Fatal("unable to extract status code from error")
			}
			if s.Code() != tt.expCode {
				t.Errorf("NewUserObject() expected status code: %v, got: %v", tt.expCode, s.Code())
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewUserObject() = %v, want %v", got, tt.want)
			}

		})
	}
}

func TestUser_Clone(t *testing.T) {
	type fields struct {
		userID       string
		sku          string
		authMethods  []string
		sessionToken string
	}
	tests := []struct {
		name      string
		expCode   codes.Code
		fields    fields
		want      *User
		userError error
	}{
		{
			// This is a simple user test. The fields
			// values should be updated if there is additional
			// logic added to the auth service.
			name:      "ok",
			expCode:   codes.OK,
			fields:    fields{},
			want:      &User{},
			userError: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{
				userID:       tt.fields.userID,
				sku:          tt.fields.sku,
				authMethods:  tt.fields.authMethods,
				sessionToken: tt.fields.sessionToken,
			}
			if got := user.Clone(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("User.Clone() = %v, want %v", got, tt.want)
			}
		})
	}
}
