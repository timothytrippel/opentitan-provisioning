// Copyright lowRISC contributors (OpenTitan project).
// Licensed under the Apache License, Version 2.0, see LICENSE for details.
// SPDX-License-Identifier: Apache-2.0

// Package etcd implements a connector to a etcd database.
package etcd

import (
	"context"
	"fmt"

	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/lowRISC/opentitan-provisioning/src/proxy_buffer/store/connector"
)

// etcdDB implements a `connector.Connector` database interface.
type etcdDB struct {
	// kv is an initialized key value etcd instance.
	kv clientv3.KV
}

// New creates a etcd connector with an initialized etcd clientv3 KV instance.
func New(kv clientv3.KV) connector.Connector {
	return &etcdDB{kv: kv}
}

// Insert adds a `key` `value` pair to the database. Multiple calls with the
// same key will succeed.
func (e *etcdDB) Insert(ctx context.Context, key string, value []byte) error {
	if _, err := e.kv.Put(ctx, key, string(value)); err != nil {
		return fmt.Errorf("failed to insert data with key: %q, error: %v", key, err)
	}
	return nil
}

// Get gets the latest insterted value associated with a given `key`.
func (e *etcdDB) Get(ctx context.Context, key string) ([]byte, error) {
	res, err := e.kv.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get data associated with key: %q, error: %v", key, err)
	}
	return res.Kvs[0].Value, nil
}
