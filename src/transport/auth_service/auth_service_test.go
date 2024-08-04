// Copyright lowRISC contributors (OpenTitan project).
// Licensed under the Apache License, Version 2.0, see LICENSE for details.
// SPDX-License-Identifier: Apache-2.0

// Package auth_service implements the Auth Controller object

package auth_service

import (
	"reflect"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestAuthController_AddUser(t *testing.T) {
	type fields struct {
		enableTLS bool
		userStore *InMemoryUserStore
	}
	type args struct {
		userID      string
		token       string
		sku         string
		authMethods []string
	}

	dummyUser, _ := NewUserObject("dummyUser", "", "", []string{})

	tests := []struct {
		name    string
		expCode codes.Code
		fields  *fields
		args    *args
		want    *User
	}{
		// This is a simple AuthController add user test. The fields
		// values should be updated if there is additional
		// logic added to the AuthController.
		{
			name:    "ok",
			expCode: codes.OK,
			fields: &fields{
				enableTLS: true,
				userStore: NewInMemoryUserStore(),
			},
			args: &args{
				userID:      dummyUser.userID,
				token:       dummyUser.sessionToken,
				sku:         dummyUser.sku,
				authMethods: dummyUser.authMethods,
			},
			want: dummyUser,
		},
	}
	for _, tt := range tests {
		NewAuthControllerInstance(true)
		t.Run(tt.name, func(t *testing.T) {
			ctrl := &AuthController{
				enableTLS: tt.fields.enableTLS,
				userStore: tt.fields.userStore,
			}
			got, err := ctrl.AddUser(tt.args.userID, tt.args.token, tt.args.sku, tt.args.authMethods)
			s, ok := status.FromError(err)
			if !ok {
				t.Fatal("unable to extract status code from error")
			}
			if s.Code() != tt.expCode {
				t.Errorf("AuthController.AddUser() expected status code: %v, got: %v", tt.expCode, s.Code())
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AuthController.AddUser() = %v, want %v", got, tt.want)
			}
		})
	}
}
