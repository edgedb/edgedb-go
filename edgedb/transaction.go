// This source file is part of the EdgeDB open source project.
//
// Copyright 2020-present EdgeDB Inc. and the EdgeDB authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package edgedb

import (
	"net"

	"github.com/edgedb/edgedb-go/edgedb/marshal"
	"github.com/edgedb/edgedb-go/edgedb/protocol/format"
)

// Transaction represents a transaction or save point block.
// Transactions are created by calling the Conn.Transaction() method.
// Most callers should use `Conn.RunInTransaction()` instead.
type Transaction struct {
	client *Client
	conn   net.Conn
}

// Start a transaction or save point.
func (tx *Transaction) Start() error {
	// todo handle nested blocks and other options.
	return tx.Execute("START TRANSACTION;")
}

// Commit the transaction or save point preserving changes.
func (tx *Transaction) Commit() (err error) {
	defer func() {
		e := tx.conn.Close()
		if e != nil && err == nil {
			err = e
		}
	}()

	// todo handle nested blocks etc.
	return tx.Execute("COMMIT;")
}

// RollBack the transaction or save point block discarding changes.
func (tx *Transaction) RollBack() (err error) {
	defer func() {
		e := tx.conn.Close()
		if e != nil && err == nil {
			err = e
		}
	}()

	// todo handle nested blocks etc.
	return tx.Execute("ROLLBACK;")
}

// Execute an EdgeQL command (or commands).
// Only valid if transaction has been started.
func (tx *Transaction) Execute(query string) error {
	return scriptFlow(tx.conn, query)
}

// Query runs a query and returns the results.
// Only valid if transaction has been started.
func (tx *Transaction) Query(
	query string,
	out interface{},
	args ...interface{},
) error {
	result, err := tx.client.granularFlow(tx.conn, query, format.Binary, args)
	if err != nil {
		return err
	}

	marshal.Marshal(&out, result)
	return nil
}

// Transaction creates a new trasaction struct.
func (c *Client) Transaction() (*Transaction, error) {
	// todo support transaction options
	conn, err := c.pool.Get()
	if err != nil {
		return nil, err
	}

	return &Transaction{c, conn}, nil
}

// RunInTransaction runs a function in a transaction.
// If function returns an error transaction is rolled back,
// otherwise transaction is committed.
func (c *Client) RunInTransaction(fn func() error) error {
	// see https://pkg.go.dev/github.com/go-pg/pg/v10#DB.RunInTransaction
	panic("RunInTransaction() not implemented") // todo
}
