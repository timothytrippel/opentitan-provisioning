// Copyright lowRISC contributors (OpenTitan project).
// Licensed under the Apache License, Version 2.0, see LICENSE for details.
// SPDX-License-Identifier: Apache-2.0

// Package auth_service implements the store user object

package auth_service

import (
	"reflect"
	"sync"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func setStoreUser(user string) *User {
	ret, _ := NewUserObject(user, "", "", []string{})
	return ret
}

func TestInMemoryUserStore_Save(t *testing.T) {
	type fields struct {
		mutex sync.RWMutex
		users map[string]*User
	}
	type args struct {
		user *User
	}
	fakeMap := make(map[string]*User)

	tests := []struct {
		name    string
		expCode codes.Code
		fields  fields
		args    *args
	}{
		{
			// This is a simple user store test. The fields
			// values should be updated if there is additional
			// logic added to the auth service.
			name:    "ok",
			expCode: codes.OK,
			fields:  fields{users: fakeMap},
			args:    &args{user: setStoreUser("dummyUser")},
		},
		{
			// user store errors are converted to code.Internal.
			name:    "store already exists error",
			expCode: codes.Internal,
			fields:  fields{users: fakeMap},
			args:    &args{user: setStoreUser("dummyUser")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &InMemoryUserStore{
				mutex: tt.fields.mutex,
				users: tt.fields.users,
			}
			err := store.Save(tt.args.user)
			s, ok := status.FromError(err)
			if !ok {
				t.Fatal("unable to extract status code from error")
			}
			if s.Code() != tt.expCode {
				t.Errorf("InMemoryUserStore.Save() expected status code: %v, got: %v", tt.expCode, s.Code())
			}
		})
	}
}

func TestInMemoryUserStore_Delete(t *testing.T) {
	type fields struct {
		mutex sync.RWMutex
		users map[string]*User
	}
	type args struct {
		user *User
	}

	fakeMap := make(map[string]*User)
	dummyUser := setStoreUser("dummyUser")
	fakeMap[dummyUser.userID] = dummyUser.Clone()

	tests := []struct {
		name    string
		expCode codes.Code
		fields  fields
		args    *args
	}{
		{
			// This is a simple user delete test. The fields
			// values should be updated if there is additional
			// logic added to the auth service.
			name:    "ok",
			expCode: codes.OK,
			fields:  fields{users: fakeMap},
			args:    &args{user: dummyUser},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &InMemoryUserStore{
				mutex: tt.fields.mutex,
				users: tt.fields.users,
			}
			err := store.Delete(tt.args.user)
			s, ok := status.FromError(err)
			if !ok {
				t.Fatal("unable to extract status code from error")
			}
			if s.Code() != tt.expCode {
				t.Errorf("InMemoryUserStore.Save() expected status code: %v, got: %v", tt.expCode, s.Code())
			}
		})
	}
}

func TestInMemoryUserStore_Find(t *testing.T) {
	type fields struct {
		mutex sync.RWMutex
		users map[string]*User
	}
	type args struct {
		user *User
	}

	fakeMap := make(map[string]*User)
	dummyUser := setStoreUser("dummyUser")
	fakeMap[dummyUser.userID] = dummyUser.Clone()

	tests := []struct {
		name    string
		expCode codes.Code
		fields  fields
		args    *args
		want    *User
	}{
		{
			// This is a simple user for find test.
			name:    "ok",
			expCode: codes.OK,
			fields:  fields{users: fakeMap},
			args:    &args{user: setStoreUser("dummyUser")},
			want:    dummyUser,
		},
		{
			// This is a simple user for user not found test.
			name:    "user not found error",
			expCode: codes.Internal,
			fields:  fields{users: fakeMap},
			args:    &args{user: setStoreUser("unknownUser")},
			want:    nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &InMemoryUserStore{
				mutex: tt.fields.mutex,
				users: tt.fields.users,
			}
			got, err := store.Find(tt.args.user.userID)
			s, ok := status.FromError(err)
			if !ok {
				t.Fatal("unable to extract status code from error")
			}
			if s.Code() != tt.expCode {
				t.Errorf("InMemoryUserStore.Find() expected status code: %v, got: %v", tt.expCode, s.Code())
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InMemoryUserStore.Find() = %v, want %v", got, tt.want)
			}
		})
	}
}
