// Copyright lowRISC contributors (OpenTitan project).
// Licensed under the Apache License, Version 2.0, see LICENSE for details.
// SPDX-License-Identifier: Apache-2.0

// Package session_token implements a unique random string for each InitSession call.
package session_token

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// muToken is a mutex use to arbitrate token initialization access.
var muToken sync.RWMutex

var singleInstance *SessionToken

const (
	DictionaryBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
)

type SessionToken struct{}

func (st *SessionToken) Generate(tokenSize int) (string, error) {
	muToken.RLock()
	defer muToken.RUnlock()

	token := make([]byte, tokenSize)
	for i := range token {
		token[i] = DictionaryBytes[rand.Intn(len(DictionaryBytes))]
	}
	return string(token), nil
}

func NewSessionTokenInstance() *SessionToken {
	if singleInstance == nil {
		muToken.RLock()
		defer muToken.RUnlock()
		if singleInstance == nil {
			fmt.Println("Creating single instance now.")
			rand.Seed(time.Now().UnixNano())
			singleInstance = &SessionToken{}
		}
	}
	return singleInstance
}

func GetInstance() (*SessionToken, error) {
	if singleInstance == nil {
		return nil, status.Errorf(codes.Internal, "No instance of AuthController")
	}

	return singleInstance, nil
}
