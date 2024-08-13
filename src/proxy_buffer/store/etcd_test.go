// Copyright lowRISC contributors (OpenTitan project).
// Licensed under the Apache License, Version 2.0, see LICENSE for details.
// SPDX-License-Identifier: Apache-2.0

// Package etcd_test implements unit tests for the etcd package.
package etcd_test

import (
	"context"
	"testing"

	mvccpb "go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/lowRISC/opentitan-provisioning/src/proxy_buffer/store/etcd"
)

// mockTxn implements the clientv3.Txn interface for testing purposes.
// This interface is required by `mockKV` below.
type mockTxn struct{}

func (m *mockTxn) If(cs ...clientv3.Cmp) clientv3.Txn {
	return m
}

func (m *mockTxn) Then(ops ...clientv3.Op) clientv3.Txn {
	return m
}

func (m *mockTxn) Else(ops ...clientv3.Op) clientv3.Txn {
	return m
}

func (m *mockTxn) Commit() (*clientv3.TxnResponse, error) {
	return new(clientv3.TxnResponse), nil
}

// mockKV implements the clientv3.KV interface for testing purposes.
type mockKV struct {
	// Return values associated with the `Put` function.
	putResponse clientv3.PutResponse
	putError    error

	// Return values associated with the `Get` function.
	getResponse clientv3.GetResponse
	getError    error

	// Return values associated with the `Delete` function.
	deleteResponse clientv3.DeleteResponse
	deleteError    error

	// Return values associated with the `Compact` function.
	compactResponse clientv3.CompactResponse
	compactError    error

	// Return values associated with the `Do` funciton.
	doResponse clientv3.OpResponse
	doError    error

	// Return value associated with the `Txn` function.
	mockTxn *mockTxn
}

// addKV adds a `key` `value` into the response of the mocked `Get` method.
func (m *mockKV) addKV(key, value string) {
	m.getResponse.Kvs = append(m.getResponse.Kvs, &mvccpb.KeyValue{
		Key:   []byte(key),
		Value: []byte(value),
	})
}

func (m *mockKV) Put(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
	return &m.putResponse, m.putError
}

func (m *mockKV) Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	return &m.getResponse, m.getError
}

func (m *mockKV) Delete(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error) {
	return &m.deleteResponse, m.deleteError
}

func (m *mockKV) Compact(ctx context.Context, rev int64, opts ...clientv3.CompactOption) (*clientv3.CompactResponse, error) {
	return &m.compactResponse, m.compactError
}

func (m *mockKV) Do(ctx context.Context, op clientv3.Op) (clientv3.OpResponse, error) {
	return m.doResponse, m.doError
}

func (m *mockKV) Txn(ctx context.Context) clientv3.Txn {
	return m.mockTxn
}

func TestInsert(t *testing.T) {
	kv := &mockKV{mockTxn: &mockTxn{}}
	connector := etcd.New(kv)

	kv.putError = nil
	if err := connector.Insert(context.Background(), "foo", []byte("bar")); err != nil {
		t.Fatalf("failed to insert data: %v", err)
	}
}

func TestGet(t *testing.T) {
	kv := &mockKV{mockTxn: &mockTxn{}}
	connector := etcd.New(kv)

	kv.addKV("foo", "bar")

	res, err := connector.Get(context.Background(), "foo")
	if err != nil {
		t.Errorf("failed to get record: %v", err)
	}

	if string(res) != "bar" {
		t.Errorf("expected: %q, got: %q", "bar", res)
	}
}
