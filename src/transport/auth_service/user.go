// Copyright lowRISC contributors (OpenTitan project).
// Licensed under the Apache License, Version 2.0, see LICENSE for details.
// SPDX-License-Identifier: Apache-2.0

// Package auth_service implements the user object

package auth_service

type User struct {
	userID       string
	sku          string
	authMethods  []string
	sessionToken string
}

func NewUserObject(userID, token, sku string, authMethods []string) (*User, error) {
	user := &User{
		userID:       userID,
		sku:          sku,
		sessionToken: token,
		authMethods:  authMethods,
	}

	return user, nil
}

func (user *User) Clone() *User {
	return &User{
		userID:       user.userID,
		sku:          user.sku,
		sessionToken: user.sessionToken,
		authMethods:  user.authMethods,
	}
}
