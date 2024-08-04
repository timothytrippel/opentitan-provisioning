// Copyright lowRISC contributors (OpenTitan project).
// Licensed under the Apache License, Version 2.0, see LICENSE for details.
// SPDX-License-Identifier: Apache-2.0

// Package grpconn implements the gRPC connection utility functions

package auth_service

import (
	"sync"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserStore interface {
	Save(user *User) error
	Find(userID string) (*User, error)
	Delete(user *User) error
}

type InMemoryUserStore struct {
	mutex sync.RWMutex
	users map[string]*User
}

func NewInMemoryUserStore() *InMemoryUserStore {
	return &InMemoryUserStore{
		users: make(map[string]*User),
	}
}

func (store *InMemoryUserStore) Save(user *User) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	res := store.users[user.userID]
	if res != nil {
		return status.Errorf(codes.Internal, "ErrAlreadyExists")
	}

	store.users[user.userID] = user.Clone()
	return nil
}

func (store *InMemoryUserStore) Delete(user *User) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	if _, ok := store.users[user.userID]; ok {
		delete(store.users, user.userID)
	}

	return nil
}

func (store *InMemoryUserStore) Find(userID string) (*User, error) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	user := store.users[userID]
	if user == nil {
		return nil, status.Errorf(codes.Internal, "user not found, user = %s", userID)
	}

	return user.Clone(), nil
}
